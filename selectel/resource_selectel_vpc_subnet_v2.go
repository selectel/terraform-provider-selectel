package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceDeprecatedVPCSubnetV2() *schema.Resource {
	resource := resourceVPCPublicSubnetV2()
	resource.DeprecationMessage = "selectel_vpc_subnet_v2 is deprecated; use selectel_vpc_public_subnet_v2 instead. Existing configurations can continue using this resource."

	return resource
}
