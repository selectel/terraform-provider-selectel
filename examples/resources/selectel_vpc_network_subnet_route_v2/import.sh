export OS_DOMAIN_NAME=<domain_name>
export OS_USERNAME=<username>
export OS_PASSWORD=<password>
export INFRA_PROJECT_ID=<project_id>
export INFRA_REGION=<region>

terraform import selectel_vpc_network_subnet_route_v2.route_1 <subnet_id>-route-<destination_cidr>-<next_hop>
