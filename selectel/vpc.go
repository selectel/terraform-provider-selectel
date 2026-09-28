package selectel

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpc "github.com/selectel/vpc-go/pkg/v2"
	"github.com/selectel/vpc-go/pkg/v2/floatingip"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/internal/convert"
)

const (
	vpcV2ServiceType = "network"

	vpcFloatingIPV2Object = "VPC floating ip"
)

const (
	vpcV2StatusActive        = "ACTIVE"
	vpcV2StatusDown          = "DOWN"
	vpcV2StatusError         = "ERROR"
	vpcV2StatusBuild         = "BUILD"
	vpcV2StatusAllocating    = "ALLOCATING"
	vpcV2StatusPendingCreate = "PENDING_CREATE"
	vpcV2StatusPendingUpdate = "PENDING_UPDATE"
)

const (
	vpcV2StatusPollInterval = 2 * time.Second
	vpcV2StatusTimeout      = time.Minute
)

var errVPCV2StatusTimeout = errors.New("timeout waiting for status")

type vpcV2StatusRefreshFunc func(ctx context.Context) (string, error)

func vpcV2Client(meta any, projectID, region string) (*vpc.Client, error) {
	config := meta.(*Config)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get project-scope selvpc client for VPC api: %w", err)
	}

	endpoint, err := selvpcClient.Catalog.GetEndpoint(vpcV2ServiceType, region)
	if err != nil {
		return nil, fmt.Errorf("can't resolve the %s service endpoint in region '%s': %w",
			vpcV2ServiceType, region, err)
	}

	client, err := vpc.NewClient(vpc.Config{
		Endpoint:  endpoint.URL,
		Token:     selvpcClient.GetXAuthToken(),
		UserAgent: config.UserAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("can't build the VPC client for region '%s': %w", region, err)
	}

	return client, nil
}

func vpcCheckDeleted(d *schema.ResourceData, err error, object string) diag.Diagnostics {
	if err == nil {
		return nil
	}

	if vpc.IsErrorClass(err, vpc.ErrorClassNotFound) {
		d.SetId("")

		return nil
	}

	return diag.FromErr(errGettingObject(object, d.Id(), err))
}

func waitForVPCV2Status(
	ctx context.Context, object, id string, pending []string,
	interval, timeout time.Duration, refresh vpcV2StatusRefreshFunc,
) (string, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		status, err := refresh(ctx)
		if err != nil {
			return "", fmt.Errorf("error getting %s %s status: %w", object, id, err)
		}
		if status == vpcV2StatusError {
			return status, fmt.Errorf("%s %s is in %s status", object, id, status)
		}
		if !slices.Contains(pending, status) {
			return status, nil
		}
		if time.Now().After(deadline) {
			return status, fmt.Errorf("%w: %s %s is still %s after %s",
				errVPCV2StatusTimeout, object, id, status, timeout)
		}

		log.Printf("[DEBUG] %s %s is %s, waiting", object, id, status)

		select {
		case <-ctx.Done():
			return status, ctx.Err()
		case <-ticker.C:
		}
	}
}

func waitForVPCFloatingIPV2Active(ctx context.Context, client *vpc.Client, id string) diag.Diagnostics {
	status, err := waitForVPCV2Status(
		ctx, vpcFloatingIPV2Object, id, []string{vpcV2StatusDown},
		vpcV2StatusPollInterval, vpcV2StatusTimeout,
		func(ctx context.Context) (string, error) {
			found, err := floatingip.Get(ctx, client, id)
			if err != nil {
				return "", err
			}

			return found.Status, nil
		},
	)
	if errors.Is(err, errVPCV2StatusTimeout) {
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("%s %s is still %s", vpcFloatingIPV2Object, id, status),
			Detail: "The API accepted the change, but the network backend has not " +
				"reported the floating ip as ACTIVE yet. Check its status before " +
				"relying on the address.",
		}}
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenVPCTagsV2(d *schema.ResourceData, tags []string) {
	if err := d.Set("all_tags", tags); err != nil {
		log.Print(errSettingComplexAttr("all_tags", err))
	}

	allTags := d.Get("all_tags").(*schema.Set)
	desiredTags := d.Get("tags").(*schema.Set)
	actualTags := allTags.Intersection(desiredTags)

	if !actualTags.Equal(desiredTags) {
		if err := d.Set("tags", convert.ToStringSlice(actualTags.List())); err != nil {
			log.Print(errSettingComplexAttr("tags", err))
		}
	}
}

func expandVPCUpdateTagsV2(d *schema.ResourceData) []string {
	oldTagsRaw, newTagsRaw := d.GetChange("tags")

	return mergeVPCTagsV2(
		d.Get("all_tags").(*schema.Set),
		oldTagsRaw.(*schema.Set),
		newTagsRaw.(*schema.Set),
	)
}

func mergeVPCTagsV2(allTags, oldTags, newTags *schema.Set) []string {
	return convert.ToStringSlice(allTags.Difference(oldTags).Union(newTags).List())
}

func vpcComputedFixedIPsSchema() *schema.Schema {
	return &schema.Schema{
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
	}
}

func vpcComputedRoutesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination_cidr": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"next_hop": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}
