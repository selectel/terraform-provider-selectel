package selectel

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/selectel/mks-go/pkg/v1/kubeoptions"
)

func dataSourceMKSFeatureGatesV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMKSFeatureGateTypeV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
					ru9Region,
				}, false),
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kube_version": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"feature_gates": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kube_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"names": {
							Type:     schema.TypeSet,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceMKSFeatureGateTypeV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mksClient, diagErr := getMKSClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	featureGates, _, err := kubeoptions.ListFeatureGates(ctx, mksClient)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectFeatureGates, err))
	}

	filterSet := d.Get("filter").(*schema.Set)
	if filterSet.Len() == 0 {
		d.SetId("mks_feature_gates")
		flatFG := flattenFeatureGates(featureGates)
		if err := d.Set("feature_gates", flatFG); err != nil {
			return diag.FromErr(err)
		}

		return nil
	}

	filterMap := filterSet.List()[0].(map[string]interface{})
	kubeVersion := filterMap["kube_version"].(string)

	if kubeVersion == "" {
		return diag.Errorf("kubernetes version is not set")
	}
	kubeMinorVersion, err := kubeVersionTrimToMinor(kubeVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	availableFeatureGates, err := filterKubeOptionsByKubeVersion(featureGates, kubeMinorVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	flatFG := flattenFeatureGatesFromSlice(kubeMinorVersion, availableFeatureGates)
	if err := d.Set("feature_gates", flatFG); err != nil {
		return diag.FromErr(err)
	}

	checksum, err := stringListChecksum(availableFeatureGates)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func filterKubeOptionsByKubeVersion(options []*kubeoptions.View, version string) ([]string, error) {
	if version == "" {
		return nil, fmt.Errorf("kubernetes version is not set")
	}

	for _, view := range options {
		if view.KubeVersion == version {
			return view.Names, nil
		}
	}

	return nil, fmt.Errorf("available kubernetes options for kubernetes version %q is not found", version)
}

func dataSourceMKSAdmissionControllersV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMKSAdmissionControllersV1Read,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ru1Region,
					ru2Region,
					ru3Region,
					ru7Region,
					ru8Region,
					ru9Region,
				}, false),
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kube_version": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"admission_controllers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kube_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"names": {
							Type:     schema.TypeSet,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceMKSAdmissionControllersV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	mksClient, diagErr := getMKSClient(ctx, d, meta)
	if diagErr != nil {
		return diagErr
	}

	admissionControllers, _, err := kubeoptions.ListAdmissionControllers(ctx, mksClient)
	if err != nil {
		return diag.FromErr(errGettingObjects(objectAdmissionControllers, err))
	}

	filterSet := d.Get("filter").(*schema.Set)
	if filterSet.Len() == 0 {
		return setAllAdmissionControllers(d, admissionControllers)
	}

	filterMap := filterSet.List()[0].(map[string]interface{})
	kubeVersion := filterMap["kube_version"].(string)

	if kubeVersion == "" {
		return diag.Errorf("kubernetes version is not set")
	}
	kubeMinorVersion, err := kubeVersionTrimToMinor(kubeVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	availableAdmissionControllers, err := filterKubeOptionsByKubeVersion(admissionControllers, kubeMinorVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	flatFG := flattenAdmissionControllersFromSlice(kubeMinorVersion, availableAdmissionControllers)
	if err := d.Set("admission_controllers", flatFG); err != nil {
		return diag.FromErr(err)
	}

	checksum, err := stringListChecksum(availableAdmissionControllers)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(checksum)

	return nil
}

func setAllAdmissionControllers(d *schema.ResourceData, availableAdmissionControllers []*kubeoptions.View) diag.Diagnostics {
	d.SetId("mks_admission_controllers")
	flatFG := flattenAdmissionControllers(availableAdmissionControllers)
	if err := d.Set("admission_controllers", flatFG); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
