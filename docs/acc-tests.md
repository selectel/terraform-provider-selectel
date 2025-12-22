# Acceptance tests

## Running Acceptance tests for Global Router

Acceptance tests for Global Router requires definition Environment variables, because tests use references on existing Network resources in Cloud and Dedicated servers.

The full list of required environment values is:
```sh
export GLOBAL_ROUTER_DEDICATED_REGION=SPB-1
export GLOBAL_ROUTER_DEICATED_NETWORK_VLAN=123

export GLOBAL_ROUTER_SUBNET_CIDR=10.1.11.0/24
export GLOBAL_ROUTER_SUBNET_GATEWAY=10.1.11.2
export GLOBAL_ROUTER_SUBNET_SERVICE_ADDR1=10.1.11.253
export GLOBAL_ROUTER_SUBNET_SERVICE_ADDR2=10.1.11.254

export GLOBAL_ROUTER_CLOUD_REGION=ru-1
export GLOBAL_ROUTER_CLOUD_PROJECT_ID=222222222222222222222222222222222

export GLOBAL_ROUTER_STATIC_ROUTE_CIDR=0.0.0.0/0
export GLOBAL_ROUTER_STATIC_ROUTE_NEXT_HOP=10.1.11.3
```