# Acceptance tests

## Running Acceptance tests

All acceptance tests require the following environment variables to be set:

```sh
export OS_AUTH_URL="example.com"
export OS_REGION_NAME="region-1"
export OS_DOMAIN_NAME="123456"          # your account ID
export OS_USERNAME="your-service-user"
export OS_PASSWORD="your-service-password"
```

For tests that require a project scope, also set:

```sh
export INFRA_PROJECT_ID="your-project-id"
```

## Running Acceptance tests for Block Storage

Block Storage tests use existing, dedicated test projects and service users. The `OS_*` user must have permission to create, update, and delete volumes and snapshots in these projects. The tests create billable resources and attempt to delete them during cleanup; inspect the projects if a test process is interrupted.

In addition to the `OS_*` credentials above, configure the following inputs. All rows are required for a complete Block Storage acceptance run; individual tests only require their own fixtures.

| Environment variable | Used by / requirement |
| --- | --- |
| `INFRA_PROJECT_ID` | All Block Storage tests: existing primary project ID. |
| `INFRA_REGION` | All Block Storage tests: pool containing the test volumes; also required for import. |
| `SELECTEL_ACC_BLOCKSTORAGE_AVAILABILITY_ZONE` | Volume resource and lookup tests: pool segment in `INFRA_REGION`. |
| `SELECTEL_ACC_BLOCKSTORAGE_VOLUME_TYPE` | Volume resource and lookup tests: volume type available in that segment. |
| `SELECTEL_ACC_BLOCKSTORAGE_SECOND_PROJECT_ID` | `Replacement/project_id`: another existing project accessible to the write user; must differ from `INFRA_PROJECT_ID`. |
| `SELECTEL_ACC_BLOCKSTORAGE_SECOND_REGION` | `Replacement/region`: another available pool; must differ from `INFRA_REGION`. |
| `SELECTEL_ACC_BLOCKSTORAGE_SECOND_REGION_AVAILABILITY_ZONE` | `Replacement/region`: pool segment in the second pool. |
| `SELECTEL_ACC_BLOCKSTORAGE_SECOND_REGION_VOLUME_TYPE` | `Replacement/region`: volume type available in the second segment. |
| `SELECTEL_ACC_BLOCKSTORAGE_REGIONAL_REGION` | `RegionalTypeName`: pool supporting the regional type alias being tested. |
| `SELECTEL_ACC_BLOCKSTORAGE_REGIONAL_ZONE` | `RegionalTypeName`: segment in that pool. |
| `SELECTEL_ACC_BLOCKSTORAGE_REGIONAL_TYPE` | `RegionalTypeName`: regional type alias that the API resolves to a zonal name. |
| `SELECTEL_ACC_BLOCKSTORAGE_VIEWER_USER` | Roles test: existing service user with read-only access to the primary project in the same account. |
| `SELECTEL_ACC_BLOCKSTORAGE_VIEWER_PASSWORD` | Roles test: password of the viewer user. |

The `Sources/snapshot` scenario and the snapshot data source tests create their own source volumes and snapshots, wait for readiness, and clean them up after the target volume is destroyed. External snapshot IDs and sizes are no longer required.

The write/viewer users and their project roles are base test infrastructure. The role test does not create IAM users or require IAM administration privileges. It starts its viewer subprocess automatically; do not configure the internal `SELECTEL_ACC_BLOCKSTORAGE_VIEWER_PROCESS`, `SELECTEL_ACC_BLOCKSTORAGE_VIEWER_VOLUME_ID`, or `SELECTEL_ACC_BLOCKSTORAGE_VIEWER_VOLUME_NAME` variables yourself.

Block Storage inputs formerly named `SELECTEL_BLOCKSTORAGE_*` now use `SELECTEL_ACC_BLOCKSTORAGE_*`; `VIEWER_USER` and `VIEWER_PASSWORD` are also replaced by the names in the table. Set these variables in your local test environment before running the complete suite. Existing `OS_*` and `INFRA_*` names are unchanged.

Run the Block Storage suite with:

```sh
make testacc TESTARGS="-run '^TestAccSelectelCompute(VolumeV3|VolumeLookupV3|VolumeTypeV3|SnapshotV3|BlockStorageRoles)' -count=1"
```

Missing or invalid prerequisites fail the selected test instead of silently skipping it. For a narrower run, select the intended test with `-run` and supply its inputs.

### Cleaning up interrupted Block Storage tests

The sweeper is a separate, destructive command; normal tests do not invoke it. Stop all acceptance runs and other users of the selected test project before starting it. Never point it at a shared or production project. Use the write credentials above and explicitly select one dedicated project and its pool:

```sh
export SELECTEL_ACC_BLOCKSTORAGE_SWEEP_PROJECT_ID="your-dedicated-test-project-id"
env -u TF_ACC go test ./selectel -sweep="ru-1" -sweep-run=selectel_compute_volume_v3 -timeout=60m
```

Repeat for every project/pool used by replacement and regional-type scenarios. The sweeper does not infer projects from `INFRA_PROJECT_ID` or enumerate other projects. Run it only after stopping tests; it cannot distinguish an active fixture from an orphan in the same project.

Eligible volumes must have a reserved `tf-acc-selectel-volume-` or `tf-acc-selectel-snapshot-` name and either the provider's `selectel_tf_create_token` metadata or `purpose=snapshot-acceptance`. Snapshots must have a reserved name, `purpose=snapshot-acceptance`, and an eligible source volume. The sweeper removes cloned volumes, then snapshots, then source volumes, waiting for deletion at each step. It refuses attached test volumes and propagates API/cleanup failures. Unmarked objects, including fixtures from older runs without these markers, require manual inspection; it does not force-delete them.

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
