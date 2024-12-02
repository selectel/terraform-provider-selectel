---
layout: "selectel"
page_title: "Upgrading Terraform Selectel Provider to version 6.0.0"
sidebar_current: "docs-selectel-guide-upgrade-guide-v6"
description: |-
  How to upgrade Terraform Selectel Provider to version 6.0.0.
---

# Upgrading Terraform Selectel Provider to version 6.0.0

In version 6.0.0, Terraform Selectel Provider introduces the following changes:

- removes deprecated API resources:
  - selectel_vpc_role_v2;
  - selectel_vpc_user_v2;
  - selectel_vpc_vrrp_subnet_v2;
  - selectel_vpc_crossregion_subnet_v2;
- сhanges the names of environment variables `SEL_PROJECT_ID` and `SEL_REGION` to `INFRA_PROJECT_ID` and `INFRA_REGION` respectively;
- makes authentication parameters `auth_region` and `auth_url` required for authentication.

Before upgrading to version 6.0.0, [upgrade to the most recent 5.X version of the provider](https://registry.terraform.io/providers/selectel/selectel/latest/docs/guides/upgrading_to_version_5) and ensure that your environment successfully runs `terraform plan`. You should not see changes you do not expect or deprecation notices.

## Check authentication and rename environment variables

1. In the Terraform configuration, update the version constraints:

   ```hcl
   terraform {
   required_providers {
     selectel  = {
       source  = "selectel/selectel"
       version = "~> 6.0"
     }
     openstack = {
       source  = "terraform-provider-openstack/openstack"
       version = "1.54.0"
     }
   }
   }
   ```

2. Ensure that the required authentication parameters `auth_region` and `auth_url` are in the configuration:

```hcl

provider "selectel" {
  domain_name = "123456"
  username    = "user"
  password    = "password"
  auth_region = "pool"
  auth_url    = "https://cloud.api.selcloud.ru/identity/v3/"
}

```

3. If you use environment variables `SEL_PROJECT_ID` or `SEL_REGION`, rename them to `INFRA_PROJECT_ID` and `INFRA_REGION` respectively.
4. To download the new version, initialize the Terraform configuration.

   ```bash
   terraform init -upgrade
   ```

## Replace the removed resources

Replace the removed resources if you have them in the configuration.

1. Backup the `.tfstate` file.
2. Remove the `selectel_vpc_vrrp_subnet_v2` resource from the `.tfstate` file:

   ```bash
   terraform state rm $(terraform state list | grep selectel_vpc_vrrp_subnet_v2)
   ```

3. Remove the `selectel_vpc_crossregion_subnet_v2` resource from the `.tfstate` file:

   ```bash
   terraform state rm $(terraform state list | grep selectel_vpc_crossregion_subnet_v2)
   ```

4. In the configuration files, remove the `selectel_vpc_vrrp_subnet_v2` and `selectel_vpc_crossregion_subnet_v2` resources.
5. To ensure that Terraform applies the required changes, preview the changes:

   ```bash
   terraform plan
   ```

6. Apply the changes:

   ```bash
   terraform apply
   ```
