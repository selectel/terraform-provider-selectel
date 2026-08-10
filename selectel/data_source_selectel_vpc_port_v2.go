package selectel

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/port"
)

func dataSourceVPCPortV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVPCPortV2Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"port_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"admin_state_up": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"device_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"device_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			// fixed_ip matches a port that has this address among its fixed IPs,
			// not one whose whole set equals it.
			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// security_group_ids matches a port that has at least one of the
			// given groups assigned.
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"all_fixed_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"all_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"allowed_address_pairs": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mac_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"extra_dhcp_option": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_version": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"binding": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vnic_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_assignment": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"all_tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"port_security_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"qos_policy_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dhcp_blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCPortV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := networkingV2Client(meta, d.Get("project_id").(string), d.Get("region").(string))
	if err != nil {
		return diag.FromErr(err)
	}

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
	if v, ok := d.GetOkExists("admin_state_up"); ok {
		filters["admin_state_up"] = []string{strconv.FormatBool(v.(bool))}
	}
	if v, ok := d.GetOk("tags"); ok {
		filters["tags"] = []string{strings.Join(expandToStringSlice(v.(*schema.Set).List()), ",")}
	}

	found, err := port.List(ctx, client, vpc.ListOptions{
		SelectionOptions: vpc.SelectionOptions{Filters: filters},
	})
	if err != nil {
		return diag.FromErr(errGettingObjects(networkingPortV2Object, err))
	}

	matched := filterNetworkingPortV2(d, found)

	switch len(matched) {
	case 0:
		return diag.FromErr(fmt.Errorf(
			"no %s found for the given search criteria, please adjust them and try again",
			networkingPortV2Object))
	case 1:
	default:
		return diag.FromErr(fmt.Errorf(
			"%d %ss found for the given search criteria, please specify a more precise search",
			len(matched), networkingPortV2Object))
	}

	result := matched[0]

	d.SetId(result.ID)
	d.Set("port_id", result.ID)
	flattenNetworkingPortV2(d, &result)

	log.Printf("[DEBUG] found networking port %s", result.ID)

	return nil
}

// filterNetworkingPortV2 applies the two selectors the list API cannot express:
// a single fixed IP that must be among the port's addresses, and a set of
// security groups of which the port must have at least one.
func filterNetworkingPortV2(d *schema.ResourceData, candidates []port.Port) []port.Port {
	wantFixedIP, hasFixedIP := d.GetOk("fixed_ip")

	var wantGroups []string
	if v, ok := d.GetOk("security_group_ids"); ok {
		wantGroups = expandToStringSlice(v.(*schema.Set).List())
	}

	matched := make([]port.Port, 0, len(candidates))

	for _, candidate := range candidates {
		if hasFixedIP && !networkingPortV2HasFixedIP(candidate, wantFixedIP.(string)) {
			continue
		}
		if len(wantGroups) > 0 && !networkingPortV2HasAnyGroup(candidate, wantGroups) {
			continue
		}

		matched = append(matched, candidate)
	}

	return matched
}

func networkingPortV2HasFixedIP(candidate port.Port, address string) bool {
	for _, fixedIP := range candidate.FixedIPs {
		if fixedIP.IPAddress == address {
			return true
		}
	}

	return false
}

func networkingPortV2HasAnyGroup(candidate port.Port, groups []string) bool {
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
