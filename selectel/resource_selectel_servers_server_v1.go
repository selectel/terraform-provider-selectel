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
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/api/servers"
	waiters "github.com/terraform-providers/terraform-provider-selectel/selectel/waiters/servers"
)

func resourceServersServerV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServersServerV1Create,
		ReadContext:   resourceServersServerV1Read,
		UpdateContext: resourceServersServerV1Update,
		DeleteContext: resourceServersServerV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServersServerV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(80 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: resourceServersServerV1Schema(),
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			_ = d.Clear(serversServerSchemaForceUpdateAdditionalParams)
			return nil
		},
	}
}

func resourceServersServerV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	partitionsConfigFromSchema, err := resourceServersServerV1ReadPartitionsConfig(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"failed to read partitions config: %w", err,
		))
	}

	var (
		locationID      = d.Get(serversServerSchemaKeyLocationID).(string)
		osID            = d.Get(serversServerSchemaKeyOSID).(string)
		configurationID = d.Get(serversServerSchemaKeyConfigurationID).(string)
		pricePlanName   = d.Get(serversServerSchemaKeyPricePlanName).(string)
		sshKeyName, _   = d.Get(serversServerSchemaKeyOSSSHKeyName).(string)

		publicSubnetID, _ = d.Get(serversServerSchemaKeyPublicSubnetID).(string)
		privateSubnet, _  = d.Get(serversServerSchemaKeyPrivateSubnet).(string)
	)

	data, err := resourceServersServerV1CreateLoadData(
		ctx, dsClient, locationID, osID, configurationID, publicSubnetID, privateSubnet,
		sshKeyName, pricePlanName, partitionsConfigFromSchema,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// validating availability of the server, OS, price plan and balance, partitions config

	var (
		userData, _ = d.Get(serversServerSchemaKeyOSUserData).(string)
		sshKeyPK, _ = d.Get(serversServerSchemaKeyOSSSHKey).(string)
	)

	if data.sshKeyByName != nil {
		sshKeyPK = data.sshKeyByName.PublicKey
	}

	err = resourceServersServerV1CreateValidatePreconditions(
		ctx, dsClient, data, locationID, data.pricePlan.UUID, configurationID, osID, userData != "",
		sshKeyPK != "" || data.sshKeyByName != nil, privateSubnet != "",
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// creating

	var (
		hostName = resourceServersServerV1GenerateHostNameIfNotPresented(d)

		password, _ = d.Get(serversServerSchemaKeyOSPassword).(string)

		req = &servers.ServerBillingPostPayload{
			ServiceUUID:      configurationID,
			PricePlanUUID:    data.pricePlan.UUID,
			PayCurrency:      data.billingPayCurrency,
			LocationUUID:     locationID,
			Quantity:         1,
			IPList:           data.ipsPublic,
			LocalIPList:      data.ipsPrivate,
			LocalSubnetUUID:  data.localSubnetUUID,
			ProjectUUID:      d.Get(serversServerSchemaKeyProjectID).(string),
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

	log.Print(msgCreate(objectServer, req.CopyWithoutSensitiveData()))

	billingRes, _, err := dsClient.ServerBilling(ctx, req, data.server.IsServerChip)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectServer, err))
	}

	switch {
	case len(billingRes) > 1:
		return diag.FromErr(fmt.Errorf(
			"failed to create one %s %s: multiple resources created: %#v", objectServer, configurationID, billingRes,
		))

	case len(billingRes) == 0:
		return diag.FromErr(fmt.Errorf(
			"failed to create %s %s: no resource returned", objectServer, configurationID,
		))
	}

	uuid := billingRes[0].UUID

	d.SetId(uuid)

	log.Printf("[DEBUG] waiting for server %s to become 'ACTIVE'", uuid)

	timeout := d.Timeout(schema.TimeoutCreate)
	err = waiters.WaitForServersServerV1ActiveState(ctx, dsClient, uuid, timeout)
	if err != nil {
		return diag.FromErr(errCreatingObject(objectServer, err))
	}

	return nil
}

type serversServerV1CreateData struct {
	os                 *servers.OperatingSystem
	server             *servers.Server
	partitions         servers.PartitionsConfig
	ipsPublic          []net.IP
	ipsPrivate         []net.IP
	localSubnetUUID    string
	sshKeyByName       *servers.SSHKey
	billing            *servers.ServiceBilling
	billingPayCurrency string
	pricePlan          *servers.PricePlan
}

func resourceServersServerV1CreateLoadData(
	ctx context.Context, dsClient *servers.ServiceClient,
	locationID, osID, configurationID, publicSubnetID, privateSubnet, sshKeyName, pricePlanName string,
	partitionsConfigFromSchema *PartitionsConfig,
) (*serversServerV1CreateData, error) {
	operatingSystems, _, err := dsClient.OperatingSystems(ctx, &servers.OperatingSystemsQuery{
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

	partitionsConfig, err := resourceServersServerV1LoadPartitions(ctx, partitionsConfigFromSchema, dsClient, os, configurationID)
	if err != nil {
		return nil, err
	}

	var publicIPs []net.IP
	if publicSubnetID != "" {
		// also validating the sufficiency of free addresses
		publicIP, err := resourceServersServerV1GetFreePublicIPs(ctx, dsClient, locationID, publicSubnetID)
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

		privateIP, localSubnetUUID, err = resourceServersServerV1GetFreePrivateIPs(ctx, dsClient, locationID, privateSubnet)
		if err != nil {
			return nil, err
		}

		privateIPs = append(privateIPs, privateIP)
	}

	var sshKey *servers.SSHKey
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

	billing, _, err := dsClient.ServerCalculateBilling(ctx, configurationID, locationID, pricePlan.UUID, servers.ServiceBillingPayCurrencyMain, isServerChip)
	if err != nil {
		return nil, fmt.Errorf("can't calculate billing for %s %s: %w", objectServer, configurationID, err)
	}

	billingPayCurrency := servers.ServiceBillingPayCurrencyMain

	if !billing.HasEnoughBalance {
		billing, _, err = dsClient.ServerCalculateBilling(ctx, configurationID, locationID, pricePlan.UUID, servers.ServiceBillingPayCurrencyBonus, isServerChip)
		if err != nil {
			return nil, fmt.Errorf("can't calculate billing for %s %s: %w", objectServer, configurationID, err)
		}

		billingPayCurrency = servers.ServiceBillingPayCurrencyBonus
	}

	return &serversServerV1CreateData{
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

func resourceServersServerV1LoadPartitions(
	ctx context.Context, partitionsConfigFromSchema *PartitionsConfig, dsClient *servers.ServiceClient, os *servers.OperatingSystem,
	configurationID string,
) (servers.PartitionsConfig, error) {
	if !partitionsConfigFromSchema.IsEmpty() || os.Partitioning {
		if !os.Partitioning { // in case of configured partitions
			return nil, fmt.Errorf(
				"%s %s does not support partitions config", objectOS, os.OSValue,
			)
		}

		localDrives, _, err := dsClient.LocalDrives(ctx, configurationID)
		if err != nil {
			return nil, fmt.Errorf(
				"error getting local drives for %s %s: %w", objectServer, configurationID, err,
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

func resourceServersServerV1CreateValidatePreconditions(
	ctx context.Context, dsClient *servers.ServiceClient,
	data *serversServerV1CreateData,
	locationID, pricePlanID, configurationID, osID string,
	needUserData, sshKey bool,
	needPrivateIP bool,
) error {
	switch {
	case !data.server.IsLocationAvailable(locationID):
		return fmt.Errorf(
			"%s %s is not available for %s %s", objectLocation, locationID, objectServer, configurationID,
		)

	case !data.server.IsPricePlanAvailableForLocation(pricePlanID, locationID):
		return fmt.Errorf(
			"price-plan %s is not available for %s %s in %s %s",
			pricePlanID, objectServer, configurationID, objectLocation, locationID,
		)

	case data.os == nil:
		return fmt.Errorf(
			"%s %s is not available for %s %s in %s %s",
			objectOS, osID, objectServer, configurationID, objectLocation, locationID,
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
			objectServer, configurationID, pricePlanID, objectLocation, locationID,
		)

	case needPrivateIP && !data.server.IsPrivateNetworkAvailable():
		return fmt.Errorf(
			"%s %s does not support private network", objectServer, configurationID,
		)

	case needPrivateIP && !data.os.IsPrivateNetworkAvailable():
		return fmt.Errorf(
			"%s %s does not support private network", objectOS, osID,
		)
	}

	_, _, err := dsClient.PartitionsValidate(ctx, data.partitions, configurationID)
	if err != nil {
		return fmt.Errorf(
			"failed to validate partitions config for %s %s: %w", objectServer, configurationID, err,
		)
	}

	return nil
}

func resourceServersServerV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgGet(objectServer, d.Id()))

	rd, _, err := dsClient.ResourceDetails(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errGettingObject(objectServer, d.Id(), err))
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
		sshKeyName, _ = d.Get(serversServerSchemaKeyOSSSHKeyName).(string)
		key           = keys.FindOneByPK(resourceOS.UserSSHKey)
	)
	switch {
	case key != nil && sshKeyName != "":
		_ = d.Set("ssh_key_name", key.Name)

	default:
		_ = d.Set("ssh_key", resourceOS.UserSSHKey)
	}

	operatingSystems, _, err := dsClient.OperatingSystems(ctx, &servers.OperatingSystemsQuery{
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

func resourceServersServerV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Print(msgDelete(objectServer, d.Id()))

	_, err := dsClient.DeleteResource(ctx, d.Id())
	if err != nil {
		return diag.FromErr(errDeletingObject(objectServer, d.Id(), err))
	}

	log.Printf("[DEBUG] waiting for server %s to become 'EXPIRING'", d.Id())

	timeout := d.Timeout(schema.TimeoutDelete)
	err = waiters.WaitForServersServerV1RefusedToRenewState(ctx, dsClient, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(errDeletingObject(objectServer, d.Id(), err))
	}

	return nil
}

func resourceServersServerV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dsClient, diagErr := getServersClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	var (
		locationID      = d.Get(serversServerSchemaKeyLocationID).(string)
		configurationID = d.Get(serversServerSchemaKeyConfigurationID).(string)
		osID            = d.Get(serversServerSchemaKeyOSID).(string)
		sshKeyName, _   = d.Get(serversServerSchemaKeyOSSSHKeyName).(string)
	)

	data, err := resourceServersServerV1UpdateLoadData(ctx, dsClient, d, locationID, osID, configurationID, sshKeyName)
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		userData, _ = d.Get(serversServerSchemaKeyOSUserData).(string)
		sshKeyPK, _ = d.Get(serversServerSchemaKeyOSSSHKey).(string)
	)

	if data.sshKeyByName != nil {
		sshKeyPK = data.sshKeyByName.PublicKey
	}

	err = resourceServersServerV1UpdateValidatePreconditions(
		ctx, d, dsClient, data.os, data.partitions, userData != "", sshKeyPK != "" || data.sshKeyByName != nil,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	var (
		hostName = resourceServersServerV1GenerateHostNameIfNotPresented(d)

		password, _ = d.Get(serversServerSchemaKeyOSPassword).(string)

		payload = &servers.InstallNewOSPayload{
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

	log.Print(msgUpdate(objectServer, d.Id(), payload.CopyWithoutSensitiveData()))

	_, err = dsClient.InstallNewOS(ctx, payload, d.Id())
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectServer, d.Id(), err))
	}

	log.Printf("[DEBUG] waiting for server %s to become 'ACTIVE'", d.Id())

	timeout := d.Timeout(schema.TimeoutUpdate)
	err = waiters.WaitForServersServerInstallNewOSV1ActiveState(ctx, dsClient, d.Id(), timeout)
	if err != nil {
		return diag.FromErr(errUpdatingObject(objectServer, d.Id(), err))
	}

	return nil
}

type serversServerV1UpdateData struct {
	os           *servers.OperatingSystem
	partitions   servers.PartitionsConfig
	sshKeyByName *servers.SSHKey
}

func resourceServersServerV1UpdateLoadData(
	ctx context.Context, dsClient *servers.ServiceClient, d *schema.ResourceData,
	locationID, osID, configurationID, sshKeyName string,
) (*serversServerV1UpdateData, error) {
	operatingSystems, _, err := dsClient.OperatingSystems(ctx, &servers.OperatingSystemsQuery{
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

	partitionsConfigFromSchema, err := resourceServersServerV1ReadPartitionsConfig(d)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read partitions config: %w", err,
		)
	}

	partitionsConfig, err := resourceServersServerV1LoadPartitions(ctx, partitionsConfigFromSchema, dsClient, os, configurationID)
	if err != nil {
		return nil, err
	}

	var sshKey *servers.SSHKey
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

	return &serversServerV1UpdateData{
		os:           os,
		partitions:   partitionsConfig,
		sshKeyByName: sshKey,
	}, nil
}

func resourceServersServerV1UpdateValidatePreconditions(
	ctx context.Context, d *schema.ResourceData, dsClient *servers.ServiceClient,
	os *servers.OperatingSystem, partitions servers.PartitionsConfig,
	needUserData, needSSHKey bool,
) error {
	var (
		osID                           = d.Get(serversServerSchemaKeyOSID).(string)
		forceUpdateAdditionalParams, _ = d.Get(serversServerSchemaForceUpdateAdditionalParams).(bool)

		isAdditionalParamsChanged = d.HasChange(serversServerSchemaKeyOSHostName) ||
			d.HasChange(serversServerSchemaKeyOSSSHKey) ||
			d.HasChange(serversServerSchemaKeyOSSSHKeyName) ||
			d.HasChange(serversServerSchemaKeyOSPassword) ||
			d.HasChange(serversServerSchemaKeyOSPartitionsConfig) ||
			d.HasChange(serversServerSchemaKeyOSUserData)
	)

	switch {
	case !(d.HasChange(serversServerSchemaKeyOSID) || (forceUpdateAdditionalParams && isAdditionalParamsChanged)):
		return fmt.Errorf("can't update cause os configuration has not changed")

	case d.HasChange(serversServerSchemaKeyProjectID):
		prevID, _ := d.GetChange(serversServerSchemaKeyProjectID)

		return fmt.Errorf("can't update cause project ID has changed, use previous id %s", prevID)

	case d.HasChange(serversServerSchemaKeyLocationID):
		prevID, _ := d.GetChange(serversServerSchemaKeyLocationID)

		return fmt.Errorf("can't update cause location ID has changed, use previous id %s", prevID)

	case d.HasChange(serversServerSchemaKeyConfigurationID):
		prevID, _ := d.GetChange(serversServerSchemaKeyConfigurationID)

		return fmt.Errorf("can't update cause configuration ID has changed, use previous id %s", prevID)

	case d.HasChange(serversServerSchemaKeyPricePlanName):
		prevName, _ := d.GetChange(serversServerSchemaKeyPricePlanName)

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

	diagErr := resourceServersServerV1UpdateValidatePreconditionsAdditionalOSParams(d, forceUpdateAdditionalParams || d.HasChange(serversServerSchemaKeyOSID))
	if diagErr != nil {
		return diagErr
	}

	configurationID := d.Get(serversServerSchemaKeyConfigurationID).(string)

	_, _, err := dsClient.PartitionsValidate(ctx, partitions, configurationID)
	if err != nil {
		return fmt.Errorf(
			"failed to validate partitions config: %w", err,
		)
	}

	return nil
}

func resourceServersServerV1UpdateValidatePreconditionsAdditionalOSParams(
	d *schema.ResourceData, canUpdateAdditionalOSParams bool,
) error {
	switch {
	case !canUpdateAdditionalOSParams && d.HasChange(serversServerSchemaKeyOSHostName):
		prevName, _ := d.GetChange(serversServerSchemaKeyOSHostName)

		return fmt.Errorf("can't update cause host name has changed, use previous name %s or %s flag", prevName, serversServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(serversServerSchemaKeyOSSSHKey):
		prevKey, _ := d.GetChange(serversServerSchemaKeyOSSSHKey)

		return fmt.Errorf("can't update cause ssh key has changed, use previous key %s or %s flag", prevKey, serversServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(serversServerSchemaKeyOSSSHKeyName):
		prevName, _ := d.GetChange(serversServerSchemaKeyOSSSHKeyName)

		return fmt.Errorf("can't update cause ssh key name has changed, use previous name %s or %s flag", prevName, serversServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(serversServerSchemaKeyOSPassword):
		return fmt.Errorf("can't update cause os password has changed, use previous password or %s flag", serversServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(serversServerSchemaKeyOSPartitionsConfig):
		return fmt.Errorf("can't update cause partitions has changed or %s flag", serversServerSchemaForceUpdateAdditionalParams)

	case !canUpdateAdditionalOSParams && d.HasChange(serversServerSchemaKeyOSUserData):
		prevScript, _ := d.GetChange(serversServerSchemaKeyOSUserData)

		return fmt.Errorf("can't update cause user data has changed, use previous data %s or %s flag", prevScript, serversServerSchemaForceUpdateAdditionalParams)
	}

	return nil
}

func resourceServersServerV1ImportState(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, errors.New("project_id must be set for the resource import")
	}

	_ = d.Set("project_id", config.ProjectID)

	return []*schema.ResourceData{d}, nil
}
