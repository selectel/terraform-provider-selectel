package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/dedicatedservers"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/servers"
)

func resourceDedicatedServersServerV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDedicatedServersServerV1Create,
		ReadContext:   resourceDedicatedServersServerV1Read,
		UpdateContext: resourceDedicatedServersServerV1Update,
		DeleteContext: resourceDedicatedServersServerV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDedicatedServersServerV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(80 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: resourceDedicatedServersServerV1Schema(),
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			_ = d.Clear(dedicatedServersServerSchemaForceUpdateAdditionalParams)
			return nil
		},
	}
}

func resourceDedicatedServersServerV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	partitionsConfigFromSchema, err := resourceDedicatedServersServerV1ReadPartitionsConfig(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"failed to read partitions config: %w", err,
		))
	}

	var (
		locationID      = d.Get(dedicatedServersServerSchemaKeyLocationID).(string)
		osID            = d.Get(dedicatedServersServerSchemaKeyOSID).(string)
		configurationID = d.Get(dedicatedServersServerSchemaKeyConfigurationID).(string)
		pricePlanName   = d.Get(dedicatedServersServerSchemaKeyPricePlanName).(string)
		sshKeyName, _   = d.Get(dedicatedServersServerSchemaKeyOSSSHKeyName).(string)

		publicSubnetID, _ = d.Get(dedicatedServersServerSchemaKeyPublicSubnetID).(string)
		privateSubnet, _  = d.Get(dedicatedServersServerSchemaKeyPrivateSubnet).(string)
	)

	data, err := resourceDedicatedServersServerV1CreateLoadData(
		ctx, dsClient, locationID, osID, configurationID, publicSubnetID, privateSubnet,
		sshKeyName, pricePlanName, partitionsConfigFromSchema,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// validating availability of the server, OS, price plan and balance, partitions config

	var (
		userData, _ = d.Get(dedicatedServersServerSchemaKeyOSUserData).(string)
		sshKeyPK, _ = d.Get(dedicatedServersServerSchemaKeyOSSSHKey).(string)
	)

	if data.sshKeyByName != nil {
		sshKeyPK = data.sshKeyByName.PublicKey
	}

	err = resourceDedicatedServersServerV1CreateValidatePreconditions(
		ctx, dsClient, data, locationID, data.pricePlan.UUID, configurationID, osID, userData != "",
		sshKeyPK != "" || data.sshKeyByName != nil, privateSubnet != "",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// creating

	var (
		hostName = resourceDedicatedServersServerV1GenerateHostNameIfNotPresented(d)

		password, _ = d.Get(dedicatedServersServerSchemaKeyOSPassword).(string)

		req = &dedicatedservers.ServerBillingPostPayload{
			ServiceUUID:      configurationID,
			PricePlanUUID:    data.pricePlan.UUID,
			PayCurrency:      data.billingPayCurrency,
			LocationUUID:     locationID,
			Quantity:         1,
			IPList:           data.ipsPublic,
			LocalIPList:      data.ipsPrivate,
			LocalSubnetUUID:  data.localSubnetUUID,
			ProjectUUID:      d.Get(dedicatedServersServerSchemaKeyProjectID).(string),
			PartitionsConfig: data.partitions,
			OSVersion:        data.os.VersionValue,
			OSTemplate:       data.os.OSValue,
			OSArch:           data.os.Arch,
			UserSSHKey:       sshKeyPK,
			UserHostname:     hostName,
			UserDesc:         hostName,
			Password:         password,
			UserData:         userData,
		}
	)

	log.Print(msgCreate(objectDedicatedServer, req.CopyWithoutSensitiveData()))

	billingRes, _, err := dsClient.ServerBilling(ctx, req, data.server.IsServerChip)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDedicatedServer, err))
	}

	switch {
	case len(billingRes) > 1:
		return diag.FromErr(fmt.Errorf(
			"failed to create one %s %s: multiple resources created: %#v", objectDedicatedServer, configurationID, billingRes,
		))

	case len(billingRes) == 0:
		return diag.FromErr(fmt.Errorf(
			"failed to create %s %s: no resource returned", objectDedicatedServer, configurationID,
		))
	}

	uuid := billingRes[0].UUID

	d.SetId(uuid)

	log.Printf("[DEBUG] waiting for server %s to become 'ACTIVE'", uuid)

	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForServersServerV1ActiveState(ctx, dsClient, uuid, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectDedicatedServer, err))
	}

	return nil
}

type serversDedicatedServerV1CreateData struct {
	os                 *dedicatedservers.OperatingSystem
	server             *dedicatedservers.Server
	partitions         dedicatedservers.PartitionsConfig
	ipsPublic          []net.IP
	ipsPrivate         []net.IP
	localSubnetUUID    string
	sshKeyByName       *dedicatedservers.SSHKey
	billing            *dedicatedservers.ServiceBilling
	billingPayCurrency string
	pricePlan          *dedicatedservers.PricePlan
}

func resourceDedicatedServersServerV1CreateLoadData(
	ctx context.Context, dsClient *dedicatedservers.ServiceClient,
	locationID, osID, configurationID, publicSubnetID, privateSubnet, sshKeyName, pricePlanName string,
	partitionsConfigFromSchema *PartitionsConfig,
) (*serversDedicatedServerV1CreateData, error) {
	operatingSystems, _, err := dsClient.OperatingSystems(ctx, &dedicatedservers.OperatingSystemsQuery{
		LocationID: locationID,
		ServiceID:  configurationID,
	})
	if err != nil {
		return nil, fmt.Errorf("getting os: %w", err)
	}

	os := operatingSystems.FindOneByID(osID)
	if os == nil {
		return nil, fmt.Errorf("os %s not found", osID)
	}

	service, _, err := dsClient.Service(ctx, configurationID)
	if err != nil {
		return nil, fmt.Errorf("getting service %s: %w", configurationID, err)
	}

	isServerChip := service.IsServerChip()
	isServer := service.IsServer()
	if !isServer && !isServerChip {
		return nil, errors.New(
			"configuration is neither a server nor a server chip",
		)
	}

	server, _, err := dsClient.ServerByID(ctx, configurationID, isServerChip)
	if err != nil {
		return nil, fmt.Errorf("getting server %s: %w", configurationID, err)
	}

	partitionsConfig, err := resourceDedicatedServersServerV1LoadPartitions(ctx, partitionsConfigFromSchema, dsClient, os, configurationID)
	if err != nil {
		return nil, err
	}

	var publicIPs []net.IP
	if publicSubnetID != "" {
		// also validating the sufficiency of free addresses
		publicIP, err := resourceDedicatedServersServerV1GetFreePublicIPs(ctx, dsClient, locationID, publicSubnetID)
		if err != nil {
			return nil, err
		}

		publicIPs = append(publicIPs, publicIP)
	}

	var (
		privateIPs      []net.IP
		localSubnetUUID string
	)
	if privateSubnet != "" {
		// also validating the sufficiency of free addresses
		var privateIP net.IP

		privateIP, localSubnetUUID, err = resourceDedicatedServersServerV1GetFreePrivateIPs(ctx, dsClient, locationID, privateSubnet)
		if err != nil {
			return nil, err
		}

		privateIPs = append(privateIPs, privateIP)
	}

	var sshKey *dedicatedservers.SSHKey
	if sshKeyName != "" {
		sshKeys, _, err := dsClient.SSHKeys(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting SSH keys: %w", err)
		}

		sshKey = sshKeys.FindOneByName(sshKeyName)
		if sshKey == nil {
			return nil, fmt.Errorf("SSH key %s not found", sshKeyName)
		}
	}

	pricePlans, _, err := dsClient.PricePlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting price plans: %w", err)
	}

	pricePlan := pricePlans.FindOneByName(pricePlanName)
	if pricePlan == nil {
		return nil, fmt.Errorf("price plan %s not found", pricePlanName)
	}

	billing, _, err := dsClient.ServerCalculateBilling(ctx, configurationID, locationID, pricePlan.UUID, dedicatedservers.ServiceBillingPayCurrencyMain, isServerChip)
	if err != nil {
		return nil, fmt.Errorf("can't calculate billing for %s %s: %w", objectDedicatedServer, configurationID, err)
	}

	billingPayCurrency := dedicatedservers.ServiceBillingPayCurrencyMain

	if !billing.HasEnoughBalance {
		billing, _, err = dsClient.ServerCalculateBilling(ctx, configurationID, locationID, pricePlan.UUID, dedicatedservers.ServiceBillingPayCurrencyBonus, isServerChip)
		if err != nil {
			return nil, fmt.Errorf("can't calculate billing for %s %s: %w", objectDedicatedServer, configurationID, err)
		}

		billingPayCurrency = dedicatedservers.ServiceBillingPayCurrencyBonus
	}

	return &serversDedicatedServerV1CreateData{
		os:                 os,
		server:             server,
		partitions:         partitionsConfig,
		ipsPublic:          publicIPs,
		ipsPrivate:         privateIPs,
		localSubnetUUID:    localSubnetUUID,
		sshKeyByName:       sshKey,
		billing:            billing,
		billingPayCurrency: billingPayCurrency,
		pricePlan:          pricePlan,
	}, nil
}

func resourceDedicatedServersServerV1LoadPartitions(
	ctx context.Context, partitionsConfigFromSchema *PartitionsConfig, dsClient *dedicatedservers.ServiceClient, os *dedicatedservers.OperatingSystem,
	configurationID string,
) (dedicatedservers.PartitionsConfig, error) {
	if !partitionsConfigFromSchema.IsEmpty() || os.Partitioning {
		if !os.Partitioning { // in case of configured partitions
			return nil, fmt.Errorf(
				"%s %s does not support partitions config", objectOS, os.OSValue,
			)
		}

		localDrives, _, err := dsClient.LocalDrives(ctx, configurationID)
		if err != nil {
			return nil, fmt.Errorf(
				"error getting local drives for %s %s: %w", objectDedicatedServer, configurationID, err,
			)
		}

		partitionsConfig, err := partitionsConfigFromSchema.CastToAPIPartitionsConfig(localDrives, os.DefaultPartitions)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to read partitions config input: %w", err,
			)
		}

		return partitionsConfig, nil
	}

	return nil, nil
}

func resourceDedicatedServersServerV1CreateValidatePreconditions(
	ctx context.Context, dsClient *dedicatedservers.ServiceClient,
	data *serversDedicatedServerV1CreateData,
	locationID, pricePlanID, configurationID, osID string,
	needUserData, sshKey bool,
	needPrivateIP bool,
) error {
	switch {
	case !data.server.IsLocationAvailable(locationID):
		return fmt.Errorf(
			"%s %s is not available for %s %s", objectLocation, locationID, objectDedicatedServer, configurationID,
		)

	case !data.server.IsPricePlanAvailableForLocation(pricePlanID, locationID):
		return fmt.Errorf(
			"price-plan %s is not available for %s %s in %s %s",
			pricePlanID, objectDedicatedServer, configurationID, objectLocation, locationID,
		)

	case data.os == nil:
		return fmt.Errorf(
			"%s %s is not available for %s %s in %s %s",
			objectOS, osID, objectDedicatedServer, configurationID, objectLocation, locationID,
		)

	case needUserData && !data.os.ScriptAllowed:
		return fmt.Errorf(
			"%s %s does not allow scripts", objectOS, osID,
		)

	case sshKey && !data.os.IsSSHKeyAllowed:
		return fmt.Errorf(
			"%s %s does not allow SSH keys", objectOS, osID,
		)

	case data.partitions != nil && !data.os.Partitioning:
		return fmt.Errorf(
			"%s %s does not support partitions config", objectOS, data.os.OSValue,
		)

	case !data.billing.HasEnoughBalance:
		return fmt.Errorf(
			"%s %s is not available for price-plan %s in %s %s because of insufficient balance (main, bonus)",
			objectDedicatedServer, configurationID, pricePlanID, objectLocation, locationID,
		)

	case needPrivateIP && !data.server.IsPrivateNetworkAvailable():
		return fmt.Errorf(
			"%s %s does not support private network", objectDedicatedServer, configurationID,
		)

	case needPrivateIP && !data.os.IsPrivateNetworkAvailable():
		return fmt.Errorf(
			"%s %s does not support private network", objectOS, osID,
		)
	}

	_, _, err := dsClient.PartitionsValidate(ctx, data.partitions, configurationID)
	if err != nil {
		return fmt.Errorf(
			"failed to validate partitions config for %s %s: %w", objectDedicatedServer, configurationID, err,
		)
	}

	return nil
}

func resourceDedicatedServersServerV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectDedicatedServer, d.Id()))

	rd, _, err := dsClient.ResourceDetails(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectDedicatedServer, d.Id(), err))
	}

	_ = d.Set("location_id", rd.LocationUUID)
	_ = d.Set("configuration_id", rd.ServiceUUID)
	_ = d.Set("price_plan_name", rd.Billing.CurrentPricePlan.Name)

	resourceOS, _, err := dsClient.OperatingSystemByResource(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"error getting OS for server %s: %w", d.Id(), err,
		))
	}

	_ = d.Set("os_host_name", resourceOS.UserHostName)
	_ = d.Set("user_data", resourceOS.UserData)
	_ = d.Set("os_password", resourceOS.Password)

	keys, _, err := dsClient.SSHKeys(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"error getting SSH keys: %w", err,
		))
	}

	var (
		sshKeyName, _ = d.Get(dedicatedServersServerSchemaKeyOSSSHKeyName).(string)
		key           = keys.FindOneByPK(resourceOS.UserSSHKey)
	)
	switch {
	case key != nil && sshKeyName != "":
		_ = d.Set("ssh_key_name", key.Name)

	default:
		_ = d.Set("ssh_key", resourceOS.UserSSHKey)
	}

	operatingSystems, _, err := dsClient.OperatingSystems(ctx, &dedicatedservers.OperatingSystemsQuery{
		LocationID: rd.LocationUUID,
		ServiceID:  rd.ServiceUUID,
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"error getting operation systems: %w", err,
		))
	}

	os := operatingSystems.FindOneByArchAndVersionAndOs(resourceOS.Arch, resourceOS.Version, resourceOS.OSValue)
	if os == nil {
		return diag.FromErr(
			fmt.Errorf("failed to find OS %s with arch %s and version %s", resourceOS.OSValue, resourceOS.Arch, resourceOS.Version),
		)
	}

	_ = d.Set("os_id", os.UUID)

	return nil
}

func resourceDedicatedServersServerV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectDedicatedServer, d.Id()))

	_, err := dsClient.DeleteResource(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDedicatedServer, d.Id(), err))
	}

	log.Printf("[DEBUG] waiting for server %s to become 'EXPIRING'", d.Id())

	timeout := d.Timeout(schema.TimeoutDelete)
	err = waiters.WaitForServersServerV1RefusedToRenewState(ctx, dsClient, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectDedicatedServer, d.Id(), err))
	}

	return nil
}

func resourceDedicatedServersServerV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getDedicatedServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	var (
		locationID      = d.Get(dedicatedServersServerSchemaKeyLocationID).(string)
		configurationID = d.Get(dedicatedServersServerSchemaKeyConfigurationID).(string)
		osID            = d.Get(dedicatedServersServerSchemaKeyOSID).(string)
		sshKeyName, _   = d.Get(dedicatedServersServerSchemaKeyOSSSHKeyName).(string)
	)

	data, err := resourceDedicatedServersServerV1UpdateLoadData(ctx, dsClient, d, locationID, osID, configurationID, sshKeyName)
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		userData, _ = d.Get(dedicatedServersServerSchemaKeyOSUserData).(string)
		sshKeyPK, _ = d.Get(dedicatedServersServerSchemaKeyOSSSHKey).(string)
	)

	if data.sshKeyByName != nil {
		sshKeyPK = data.sshKeyByName.PublicKey
	}

	err = resourceDedicatedServersServerV1UpdateValidatePreconditions(
		ctx, d, dsClient, data.os, data.partitions, userData != "", sshKeyPK != "" || data.sshKeyByName != nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		hostName = resourceDedicatedServersServerV1GenerateHostNameIfNotPresented(d)

		password, _ = d.Get(dedicatedServersServerSchemaKeyOSPassword).(string)

		payload = &dedicatedservers.InstallNewOSPayload{
			OSVersion:        data.os.VersionValue,
			OSTemplate:       data.os.OSValue,
			OSArch:           data.os.Arch,
			UserSSHKey:       sshKeyPK,
			UserHostname:     hostName,
			Password:         password,
			PartitionsConfig: data.partitions,
			UserData:         userData,
		}
	)

	log.Print(msgUpdate(objectDedicatedServer, d.Id(), payload.CopyWithoutSensitiveData()))

	_, err = dsClient.InstallNewOS(ctx, payload, d.Id())
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectDedicatedServer, d.Id(), err))
	}

	log.Printf("[DEBUG] waiting for server %s to become 'ACTIVE'", d.Id())

	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForServersServerInstallNewOSV1ActiveState(ctx, dsClient, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectDedicatedServer, d.Id(), err))
	}

	return nil
}

type serversDedicatedServerV1UpdateData struct {
	os           *dedicatedservers.OperatingSystem
	partitions   dedicatedservers.PartitionsConfig
	sshKeyByName *dedicatedservers.SSHKey
}

func resourceDedicatedServersServerV1UpdateLoadData(
	ctx context.Context, dsClient *dedicatedservers.ServiceClient, d *schema.ResourceData,
	locationID, osID, configurationID, sshKeyName string,
) (*serversDedicatedServerV1UpdateData, error) {
	operatingSystems, _, err := dsClient.OperatingSystems(ctx, &dedicatedservers.OperatingSystemsQuery{
		LocationID: locationID,
		ServiceID:  configurationID,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting operating systems: %w", err)
	}

	os := operatingSystems.FindOneByID(osID)

	if os == nil {
		return nil, fmt.Errorf("error finding operating system '%s'", osID)
	}

	partitionsConfigFromSchema, err := resourceDedicatedServersServerV1ReadPartitionsConfig(d)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read partitions config: %w", err,
		)
	}

	partitionsConfig, err := resourceDedicatedServersServerV1LoadPartitions(ctx, partitionsConfigFromSchema, dsClient, os, configurationID)
	if err != nil {
		return nil, err
	}

	var sshKey *dedicatedservers.SSHKey
	if sshKeyName != "" {
		sshKeys, _, err := dsClient.SSHKeys(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"error getting SSH keys: %w", err,
			)
		}

		sshKey = sshKeys.FindOneByName(sshKeyName)
		if sshKey == nil {
			return nil, fmt.Errorf(
				"SSH key %s not found", sshKeyName,
			)
		}
	}

	return &serversDedicatedServerV1UpdateData{
		os:           os,
		partitions:   partitionsConfig,
		sshKeyByName: sshKey,
	}, nil
}

func resourceDedicatedServersServerV1UpdateValidatePreconditions(
	ctx context.Context, d *schema.ResourceData, dsClient *dedicatedservers.ServiceClient,
	os *dedicatedservers.OperatingSystem, partitions dedicatedservers.PartitionsConfig,
	needUserData, needSSHKey bool,
) error {
	var (
		osID                           = d.Get(dedicatedServersServerSchemaKeyOSID).(string)
		forceUpdateAdditionalParams, _ = d.Get(dedicatedServersServerSchemaForceUpdateAdditionalParams).(bool)

		isAdditionalParamsChanged = d.HasChange(dedicatedServersServerSchemaKeyOSHostName) ||
			d.HasChange(dedicatedServersServerSchemaKeyOSSSHKey) ||
			d.HasChange(dedicatedServersServerSchemaKeyOSSSHKeyName) ||
			d.HasChange(dedicatedServersServerSchemaKeyOSPassword) ||
			d.HasChange(dedicatedServersServerSchemaKeyOSPartitionsConfig) ||
			d.HasChange(dedicatedServersServerSchemaKeyOSUserData)
	)

	switch {
	case !(d.HasChange(dedicatedServersServerSchemaKeyOSID) || (forceUpdateAdditionalParams && isAdditionalParamsChanged)):
		return fmt.Errorf("can't update cause os configuration has not changed")

	case d.HasChange(dedicatedServersServerSchemaKeyProjectID):
		prevID, _ := d.GetChange(dedicatedServersServerSchemaKeyProjectID)

		return fmt.Errorf("can't update cause project ID has changed, use previous id %s", prevID)

	case d.HasChange(dedicatedServersServerSchemaKeyLocationID):
		prevID, _ := d.GetChange(dedicatedServersServerSchemaKeyLocationID)

		return fmt.Errorf("can't update cause location ID has changed, use previous id %s", prevID)

	case d.HasChange(dedicatedServersServerSchemaKeyConfigurationID):
		prevID, _ := d.GetChange(dedicatedServersServerSchemaKeyConfigurationID)

		return fmt.Errorf("can't update cause configuration ID has changed, use previous id %s", prevID)

	case d.HasChange(dedicatedServersServerSchemaKeyPricePlanName):
		prevName, _ := d.GetChange(dedicatedServersServerSchemaKeyPricePlanName)

		return fmt.Errorf("can't update cause price plan name has changed, use previous name %s", prevName)

	case needUserData && !os.ScriptAllowed:
		return fmt.Errorf(
			"%s %s does not allow scripts", objectOS, osID,
		)

	case needSSHKey && !os.IsSSHKeyAllowed:
		return fmt.Errorf(
			"%s %s does not allow SSH keys", objectOS, osID,
		)

	case partitions != nil && !os.Partitioning:
		return fmt.Errorf(
			"%s %s does not support partitions config", objectOS, os.OSValue,
		)
	}

	diagErr := resourceDedicatedServersServerV1UpdateValidatePreconditionsAdditionalOSParams(d, forceUpdateAdditionalParams || d.HasChange(dedicatedServersServerSchemaKeyOSID))
	if diagErr != nil {
		return diagErr
	}

	configurationID := d.Get(dedicatedServersServerSchemaKeyConfigurationID).(string)

	_, _, err := dsClient.PartitionsValidate(ctx, partitions, configurationID)
	if err != nil {
		return fmt.Errorf(
			"failed to validate partitions config: %w", err,
		)
	}

	return nil
}

func resourceDedicatedServersServerV1UpdateValidatePreconditionsAdditionalOSParams(
	d *schema.ResourceData, canUpdateAdditionalOSParams bool,
) error {
	switch {
	case !canUpdateAdditionalOSParams && d.HasChange(dedicatedServersServerSchemaKeyOSHostName):
		prevName, _ := d.GetChange(dedicatedServersServerSchemaKeyOSHostName)

		return fmt.Errorf("can't update cause host name has changed, use previous name %s or %s flag", prevName, dedicatedServersServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(dedicatedServersServerSchemaKeyOSSSHKey):
		prevKey, _ := d.GetChange(dedicatedServersServerSchemaKeyOSSSHKey)

		return fmt.Errorf("can't update cause ssh key has changed, use previous key %s or %s flag", prevKey, dedicatedServersServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(dedicatedServersServerSchemaKeyOSSSHKeyName):
		prevName, _ := d.GetChange(dedicatedServersServerSchemaKeyOSSSHKeyName)

		return fmt.Errorf("can't update cause ssh key name has changed, use previous name %s or %s flag", prevName, dedicatedServersServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(dedicatedServersServerSchemaKeyOSPassword):
		return fmt.Errorf("can't update cause os password has changed, use previous password or %s flag", dedicatedServersServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(dedicatedServersServerSchemaKeyOSPartitionsConfig):
		return fmt.Errorf("can't update cause partitions has changed or %s flag", dedicatedServersServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(dedicatedServersServerSchemaKeyOSUserData):
		prevScript, _ := d.GetChange(dedicatedServersServerSchemaKeyOSUserData)

		return fmt.Errorf("can't update cause user data has changed, use previous data %s or %s flag", prevScript, dedicatedServersServerSchemaForceUpdateAdditionalParams)
	}

	return nil
}

func resourceDedicatedServersServerV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("project_id must be set for the resource import")
	}

	_ = d.Set("project_id", config.ProjectID)

	return []*schema.ResourceData{d}, nil
}
