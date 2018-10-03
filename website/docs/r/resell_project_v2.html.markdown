---
layout: "selvpc"
page_title: "SelVPC: selvpc_resell_project_v2"
sidebar_current: "docs-selvpc-resource-resell-project-v2"
description: |-
  Manages a V2 project resource within Resell Selectel VPC.
---

# selvpc\_resell\_project_v2

Manages a V2 project resource within Resell Selectel VPC.

## Example Usage

```hcl
resource "selvpc_resell_project_v2" "kubernetes_cluster" {
  name       = "kubernetes_cluster"
  custom_url = "kubernetes-cluster-123.selvpc.ru"
  theme {
    color = "2753E9"
  }
  quotas     = [
    {
      resource_name = "compute_cores"
      resource_quotas = [
        {
          region = "ru-3"
          zone = "ru-3a"
          value = 12
        }
      ]
    },
    {
      resource_name   = "compute_ram"
      resource_quotas = [
        {
          region = "ru-3"
          zone = "ru-3a"
          value = 20480
        }
      ]
    },
    {
      resource_name   = "volume_gigabytes_fast"
      resource_quotas = [
        {
          region = "ru-3"
          zone = "ru-3a"
          value = 100
        }
      ]
    }
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the project.

* `custom_url` - (Optional) The custom url for the project. Needs to be the 3d-level
  domain for the `selvpc.ru`. Example: `terraform-project-001.selvpc.ru`.

* `theme` - (Optional) An additional theme settings for this project. The structure is
  described below.

* `auto_quotas` - (Optional) A boolean parameter that specifies if project should
  get automatically calculated quotas.

* `quotas` - (Optional) An array of desired quotas for this project. The structure is
  described below.

The `theme` block supports:

* `color` - (Optional) A background color in hex format.

* `logo` - (Optional) An url of the project header logo.

The `quotas` block supports:

* `resource_name` - (Required) A name of the billing resource to set quotas for.

* `resource_quotas` - (Required) An array of desired billing quotas for this particular
  resource. The structure is described below.

The `resource_quotas` block supports:

* `region` - (Optional) A Selectel VPC region for the resource quota.

* `zone` - (Optional) A Selectel VPC zone for the resource quota.

* `value` - (Required) A value of the resource quota.

## Attributes Reference

The following attributes are exported:

* `name` - See Argument Reference above.

* `url` - An url of the Selectel VP project. It is set by the Selectel and can't
  be changed by the user.

* `custom_url` - See Argument Reference above.

* `enabled` - Shows if project is active or it was disabled by the Selectel.

* `theme` - See Argument Reference above.

* `quotas` - See Argument Reference above.

* `all_quotas` - Contains all quotas. They can differ from the configurable `quota`
  argument since the project will have all available resource quotas automatically applied.

## Import

Projects can be imported using the `id`, e.g.

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN terraform import selvpc_resell_project_v2.project_1 0a343062504b4d06a0fac375e466db25
```
