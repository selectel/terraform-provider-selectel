## 1.1.0 (Unreleased)

FEATURES:

* __New Resource:__ `selvpc_resell_keypair_v2` [GH-29]
* __New Resource:__ `selvpc_resell_vrrp_subnet_v2` [GH-35]

IMPROVEMENTS:

* Added tuned HTTP client to prevent errors when making call to the Resell API [GH-30]
* Added the same format for all debug messages [GH-32]
* Remove the `type` argument of the `selvpc_resell_subnet_v2` from the documentation as it doesn't exist [GH-36]
* Updated Go-selvpcclient dependency to `v1.6.0` [GH-33]
* Updated Go in Travis CI to `v1.11.4` [GH-38]
* Updated GolangCI-Lint in Travis CI to `v1.12.5` [GH-37]

## 1.0.0 (December 19, 2018)

FEATURES:

* __New Resource:__ `selvpc_resell_role_v2` ([#4](https://github.com/terraform-providers/terraform-provider-aws/issues/4))
* __New Resource:__ `selvpc_resell_subnet_v2` ([#1](https://github.com/terraform-providers/terraform-provider-aws/issues/1))
* __New Resource:__ `selvpc_resell_token_v2` ([#2](https://github.com/terraform-providers/terraform-provider-aws/issues/2))
* __New Resource:__ `selvpc_resell_user_v2` ([#3](https://github.com/terraform-providers/terraform-provider-aws/issues/3))

IMPROVEMENTS:

* Updated `Building The Provider` and `Using the provider` sections in the Readme ([#6](https://github.com/terraform-providers/terraform-provider-aws/issues/6))
* Added `GolangCI-Lint` in the `TravisCI`, removed separated linters scripts and cleaned up `GNUmakefile` ([#12](https://github.com/terraform-providers/terraform-provider-aws/issues/12))
* Added more context into error messages ([#17](https://github.com/terraform-providers/terraform-provider-aws/issues/17))
* Added tuned HTTP timeouts instead of the default ones from Go's `net/http` package ([#14](https://github.com/terraform-providers/terraform-provider-aws/issues/14))
* Updated `go-selvpcclient` dependency to `v1.5.0` ([#14](https://github.com/terraform-providers/terraform-provider-aws/issues/14))

## 0.3.0 (November 26, 2018)

IMPROVEMENTS:

* Updated `go-selvpcclient` dependency to `v1.4.0` ([#51](https://github.com/selectel/terraform-provider-selvpc/issues/51))
* Updated documentation for `floatingip_v2`, `license_v2` and `project_v2` resources ([#50](https://github.com/selectel/terraform-provider-selvpc/issues/50))
* Changed `TypeList` to `TypeSet` for the `servers`, `quotas`, `all_quotas`, `resource_quotas` attributes ([#48](https://github.com/selectel/terraform-provider-selvpc/issues/48))
* Added a check for error on setting non-scalars ([#52](https://github.com/selectel/terraform-provider-selvpc/issues/52))
* Added a check for if resources don’t exist during read with unsetting the ID ([#53](https://github.com/selectel/terraform-provider-selvpc/issues/53))
* Grouped attributes at the top of resources followed by the optional attributes ([#54](https://github.com/selectel/terraform-provider-selvpc/issues/54)) 

BUG FIXES: 

* Fixed `golint` URL in the TravisCI configuration ([#49](https://github.com/selectel/terraform-provider-selvpc/issues/49))
* Fixed `all_quotas` attribute checking in the `TestAccResellV2ProjectAutoQuotas` ([#57](https://github.com/selectel/terraform-provider-selvpc/issues/57)), ([#62](https://github.com/selectel/terraform-provider-selvpc/issues/62))
* Fixed quotas in the created project of the `selvpc_resell_floatingip_v2` resource ([#58](https://github.com/selectel/terraform-provider-selvpc/issues/58))
* Fixed `structLitKeyOrder` errors in the CI ([#60](https://github.com/selectel/terraform-provider-selvpc/issues/60))

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
