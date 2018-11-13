## 0.3.0 (Unreleased)

IMPROVEMENTS:

* Updated `go-selvpcclient` dependency to `v1.4.0` ([GH-51])
* Updated documentation for `floatingip_v2`, `license_v2` and `project_v2` resources ([GH-50])
* Changed `TypeList` to `TypeSet` for the `servers`, `quotas`, `all_quotas`, `resource_quotas` attributes ([GH-48])
* Added a check for error on setting non-scalars ([GH-52])
* Added a check for if resources don’t exist during read with unsetting the ID ([GH-53])
* Grouped attributes at the top of resources followed by the optional attributes ([GH-54]) 

BUG FIXES: 

* Fixed `golint` URL in the TravisCI configuration ([GH-49])
* Fixed `all_quotas` attribute checking in the `TestAccResellV2ProjectAutoQuotas` ([GH-57]), ([GH-62])
* Fixed quotas in the created project of the `selvpc_resell_floatingip_v2` resource ([GH-58])
* Fixed `structLitKeyOrder` errors in the CI ([GH-60])

## 0.2.0 (Oct 3, 2018)

FEATURES:

* Added `auto_quotas` attribute for the `selvpc_resell_project_v` resource ([#41](https://github.com/selectel/terraform-provider-selvpc/issues/41))

IMPROVEMENTS:

* Added `critic` target in the `GNUmakefile` that will run `gocritic` linter. This target will be called by the Travis CI ([#43](https://github.com/selectel/terraform-provider-selvpc/issues/43))
* Updated Go version to the `1.11.1` in the Travis CI configuration ([#44](https://github.com/selectel/terraform-provider-selvpc/issues/44))

## 0.1.0 (May 13, 2018)

FEATURES:

* __New Resource:__ `selvpc_resell_project_v2` ([#3](https://github.com/selectel/terraform-provider-selvpc/issues/3))
* __New Resource:__ `selvpc_resell_floatingip_v2` ([#34](https://github.com/selectel/terraform-provider-selvpc/issues/34))
* __New Resource:__ `selvpc_resell_license_v2` ([#33](https://github.com/selectel/terraform-provider-selvpc/issues/33))
