---
layout: "selvpc"
page_title: "Provider: SelVPC"
sidebar_current: "docs-selvpc-index"
description: |-
  The SelVPC provider is used to interact with the Selectel VPC resources. The provider needs the Selectel API key token to authorize its requests.
---

# SelVPC provider

The SelVPC provider is used to interact with the Selectel VPC resources. The provider
needs the Selectel API key token to authorize its requests.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the SelVPC Provider
provider "selvpc" {
  token = "SELECTEL_API_TOKEN_KEY"
}

# Create a project
resource "selvpc_resell_project_v2" "project_1" {
  # ...
}
```

## Configuration Reference

The following arguments are supported:

* `token` - (Required) The Selectel API key token. If omitted, the `SEL_TOKEN`
  environment variable is used.

* `endpoint` - (Optional) The Selectel VPC endpoint. Needed only if this provider
  is used for tests environment. If omitted, the provider will use the official
  Selectel VPC endpoint automatically.

## Additional Logging

To enable debug logging, set the `TF_LOG` environment variable to `DEBUG`:

```shell
$ env TF_LOG=DEBUG terraform apply
```

## Testing and Development

In order to run the Acceptance Tests for development you need to set
the `SEL_TOKEN` environment variable:

```shell
$ env SEL_TOKEN=SELECTEL_API_TOKEN TF_ACC=1 go test -v ./selvpc/...
```

Please create an issue describing a new feature or bug prior creating a pull
request.
