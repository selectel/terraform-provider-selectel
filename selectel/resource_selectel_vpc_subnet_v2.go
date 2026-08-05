package selectel

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// resourceDeprecatedVPCSubnetV2 preserves the original public-subnet resource
// type for existing configurations. New configurations should use the explicit
// selectel_vpc_public_subnet_v2 name.
func resourceDeprecatedVPCSubnetV2() *schema.Resource {
	resource := resourceVPCPublicSubnetV2()
	resource.DeprecationMessage = "selectel_vpc_subnet_v2 is deprecated for new configurations; use selectel_vpc_public_subnet_v2 instead. Existing configurations can continue using this resource."

	return resource
}
