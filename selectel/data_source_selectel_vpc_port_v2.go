package selectel

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/schemaconfig"
)

func dataSourceVPCPortV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCPortV2Read,
		Description: "Provides a port. " +
			"For more information about ports, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/ports/).\n\n" +
			upgradeGuideNote,
		Schema: map[string]*schema.Schema{
			"project_id": projectIDDataSourceSchema(),
			"region":     regionDataSourceSchema("port"),
			"port_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Unique identifier of the port. " +
					"To get the port ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ the network page ⟶ the **Ports** tab ⟶ copy the ID of the required port. " +
					"As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack port list` and copy the `ID` field.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the port.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Description of the port.",
			},
			"network_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Unique identifier of the network the port belongs to. " +
					"To get the network ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Cloud Platform** ⟶ **Network** ⟶ the **Private Network** tab ⟶ copy the ID of the required network. " +
					"As an alternative, you can use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack network list` and copy the `ID` field.",
			},
			"admin_state_up": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Specifies if the port is administratively up.",
			},
			"device_owner": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Owner of the device the port is attached to.",
			},
			"device_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Unique identifier of the device the port is attached to.",
			},
			"mac_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsMACAddress,
				Description:  "MAC address of the port, for example `fa:16:3e:4c:2c:30`.",
			},
			"status": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Status of the port. Available statuses are `ACTIVE` or `DOWN`.",
			},
			"dns_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "DNS name of the port. " + privateDNSLearnMore,
			},
			"fixed_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Address among the port fixed IP addresses, for example `192.168.1.10`.",
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Unique identifiers of security groups of which the port must have at least one assigned.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags of the port.",
			},
			"all_fixed_ips": vpcComputedFixedIPsSchema("All fixed IPs of the port."),
			"all_security_group_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All unique identifiers of security groups assigned to the port.",
			},
			"allowed_address_pairs": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Additional IP and MAC address pairs allowed on the port.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP address or CIDR of the pair.",
						},
						"mac_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "MAC address of the pair.",
						},
					},
				},
			},
			"extra_dhcp_option": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Extra DHCP options configured on the port.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the option.",
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Value of the option.",
						},
						"ip_version": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Internet protocol version of the option.",
						},
					},
				},
			},
			"all_tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All tags of the port.",
			},
			"port_security_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if port security is enabled.",
			},
			"dhcp_blocked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if DHCP is administratively blocked on the port.",
			},
			"blocked": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if the port is administratively blocked.",
			},
		},
	}
}

func dataSourceVPCPortV2Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client, err := vpcV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	filters := vpcPortV2ListFilters(d)

	found, err := port.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(vpcPortV2Object, err))
	}

	matched := filterVPCPortV2BySecurityGroups(d, found)

	switch len(matched) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			vpcPortV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(matched), vpcPortV2Object))
	}

	result := matched[0]

	d.SetId(result.ID)
	d.Set("port_id", result.ID)
	flattenVPCPortV2(d, &result)

	log.Printf("[DEBUG] found VPC port %s", result.ID)

	return nil
}

func vpcPortV2ListFilters(d *schema.ResourceData) map[string][]string {
	filters := make(map[string][]string)
	for schemaKey, queryKey := range map[string]string{
		"port_id":      "id",
		"name":         "name",
		"description":  "description",
		"network_id":   "network_id",
		"device_owner": "device_owner",
		"device_id":    "device_id",
		"mac_address":  "mac_address",
		"status":       "status",
		"dns_name":     "dns_name",
	} {
		if v, ok := d.GetOk(schemaKey); ok {
			filters[queryKey] = []string{v.(string)}
		}
	}
	if schemaconfig.IsSet(d, "admin_state_up") {
		filters["admin_state_up"] = []string{strconv.FormatBool(d.Get("admin_state_up").(bool))}
	}
	if v, ok := d.GetOk("fixed_ip"); ok {
		filters["fixed_ips"] = []string{"ip_address=" + v.(string)}
	}
	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(convert.ToStringSlice(v.(*schema.Set).List()), ",")}
	}

	return filters
}

func filterVPCPortV2BySecurityGroups(d *schema.ResourceData, candidates []port.Port) []port.Port {
	var wantGroups []string
	if v, ok := d.GetOk("security_group_ids"); ok {
		wantGroups = convert.ToStringSlice(v.(*schema.Set).List())
	}

	matched := make([]port.Port, 0, len(candidates))

	for _, candidate := range candidates {
		if len(wantGroups) > 0 && !vpcPortV2HasAnyGroup(candidate, wantGroups) {
			continue
		}

		matched = append(matched, candidate)
	}

	return matched
}

func vpcPortV2HasAnyGroup(candidate port.Port, groups []string) bool {
	assigned := make(map[string]struct{}, len(candidate.SecurityGroups))
	for _, group := range candidate.SecurityGroups {
		assigned[group] = struct{}{}
	}

	for _, group := range groups {
		if _, ok := assigned[group]; ok {
			return true
		}
	}

	return false
}
