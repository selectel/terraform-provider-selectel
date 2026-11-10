package selectel

const (
	secGroupIDFromResource      = "Retrieved from the [selectel_vpc_secgroup_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_secgroup_v2) resource."
	portIDFromResource          = "Retrieved from the [selectel_vpc_port_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_port_v2) resource."
	routerIDFromResource        = "Retrieved from the [selectel_vpc_router_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_router_v2) resource."
	networkSubnetIDFromResource = "Retrieved from the [selectel_vpc_network_subnet_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_subnet_v2) resource."
	networkIDFromResource       = "Retrieved from the [selectel_vpc_network_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_network_v2) resource."
	floatingIPIDFromResource    = "Retrieved from the [selectel_vpc_floatingip_v2](https://registry.terraform.io/providers/selectel/selectel/latest/docs/resources/vpc_floatingip_v2) resource."

	networkIDFromControlPanel              = "To get the network ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Private networks** tab ⟶ the network page ⟶ copy the ID under the network name."
	networkSubnetIDFromControlPanel        = "To get the subnet ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Private networks** tab ⟶ the network page ⟶ the **Subnets** tab ⟶ copy the ID of the required subnet."
	portIDFromControlPanel                 = "To get the port ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Private networks** tab ⟶ the network page ⟶ the **Ports** tab ⟶ copy the ID of the required port."
	secGroupIDFromControlPanel             = "To get the security group ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Security groups** ⟶ copy the ID in the card of the required security group."
	secGroupRegionFromControlPanel         = "To get the pool, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Security groups**. The pool is in the card of the security group, before the ID."
	routerIDFromControlPanel               = "To get the router ID, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Cloud routers** tab ⟶ copy the ID in the card of the required router."
	routerRegionFromControlPanel           = "To get the pool, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Cloud routers** tab. The pool is in the card of the router, under the router name."
	floatingIPRegionFromControlPanel       = "To get the pool, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Public floating IP addresses** tab. The pool is in the card of the public floating IP address, under the IP address."
	floatingIPIDFromControlPanel           = "To get the ID of the public floating IP address, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Public floating IP addresses** tab ⟶ copy the ID in the card of the required public floating IP address."
	subnetRouteDestinationFromControlPanel = "To get the destination CIDR, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Private networks** tab ⟶ the network page ⟶ the **Subnets** tab ⟶ **Automatic network settings** ⟶ the **Static routes** table ⟶ copy the value from the **Destination subnet** column."
	subnetRouteNextHopFromControlPanel     = "To get the next hop, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Private networks** tab ⟶ the network page ⟶ the **Subnets** tab ⟶ **Automatic network settings** ⟶ the **Static routes** table ⟶ copy the value from the **Gateway (next-hop)** column."
	networkRegionFromControlPanel          = "To get the pool, in the [Control panel](https://my.selectel.ru/vpc/), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Private networks** tab ⟶ the network page. The pool is under the network name, before the ID."
	publicPortIDFromControlPanel           = "To get the public port ID, in the [Control panel](https://my.selectel.ru/vpc/default/networks), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Direct public IP addresses** tab ⟶ copy the ID of the public port on the right side of the public port card."
	publicPortRegionFromControlPanel       = "To get the pool, in the [Control panel](https://my.selectel.ru/vpc/default/networks), go to **Products** ⟶ **Cloud Servers** ⟶ **Network** ⟶ the **Direct public IP addresses** tab. The pool is under the IP address."

	secGroupRuleIDFromOpenStackCLI         = "To get the security group rule ID, use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack security group rule list` and copy the `ID` field."
	rbacPolicyIDFromOpenStackCLI           = "To get the RBAC policy ID, use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack network rbac list` and copy the `ID` field."
	routerRouteDestinationFromOpenStackCLI = "To get the destination CIDR, use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack router show <router_id>` and find the destination of the route in the `routes` field."
	routerRouteNextHopFromOpenStackCLI     = "To get the next hop, use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack router show <router_id>` and find the next hop of the route in the `routes` field."
	portForwardingIDFromOpenStackCLI       = "To get the rule ID, use [OpenStack CLI](https://docs.selectel.ru/en/cloud-servers/tools/openstack-cli/) command `openstack floating ip port forwarding list <floatingip_id>` and copy the `ID` field."

	securityGroupRulesLearnMore = "Learn more about [Security group rules](https://docs.selectel.ru/en/cloud-servers/security-groups/about-security-groups/)."

	securityGroupsMoreInfo  = "For more information about security groups, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/security-groups/)."
	staticRoutesMoreInfo    = "For more information about static routes, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/static-routes/)."
	cloudRoutersMoreInfo    = "For more information about cloud routers, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/cloud-routers/)."
	privateNetworksMoreInfo = "For more information about private networks, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/private-networks-and-subnets/)."
	portsMoreInfo           = "For more information about ports, see the [official Selectel documentation](https://docs.selectel.ru/en/cloud-servers/cloud-networks/ports/)."
)
