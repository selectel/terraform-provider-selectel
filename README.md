Terraform Selectel Provider
=========================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img alt="Terraform Selectel Provider" src="https://selectel.ru/blog/wp-content/uploads/2019/03/PR-9299_Terraform_555x278_blog_preview@2x.png" width="600px">

Maintainers
-----------

This provider plugin is maintained by:

* Sergei Kolyshkin ([@kolsean](https://github.com/kolsean))
* Nikita Kunets ([@dkder3k](https://github.com/dkder3k))
* Ilya Kulakov ([@TawR1024](https://github.com/TawR1024))
* Tatyana Voloshina ([@T-Sh](https://github.com/T-Sh))

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.20 (to build the provider plugin)

Building The Provider
---------------------

Clone the repository

```sh
$ git clone git@github.com:selectel/terraform-provider-selectel.git
```

Enter the provider directory and build the provider

```sh
$ cd terraform-provider-selectel
$ make build
```

Using the provider
----------------------

Getting Started with Terraform at Selectel: [kb.selectel.com](https://kb.selectel.com/docs/selectel-cloud-platform/main-services/instructions/how_to_use_terraform/).
Documentation is available at: [docs/providers/selectel](https://www.terraform.io/docs/providers/selectel/index.html).  
You can find examples in this repository: [examples](https://github.com/terraform-providers/terraform-provider-selectel/tree/master/examples).  
Also there are lots of examples in the [selectel/terraform-examples](https://github.com/selectel/terraform-examples).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](https://golang.org) installed on your machine (version 1.17+ is *required*).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the current directory.

```sh
$ make build
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

Releasing the Provider
----------------------

This repository contains a GitHub Action configured to automatically build and
publish assets for release when a tag is pushed that matches the pattern `v*`
(ie. `v0.1.0`).

A [Gorelaser](https://goreleaser.com/) configuration is provided that produce
build artifacts matching the [layout required](https://www.terraform.io/docs/registry/providers/publishing.html#manually-preparing-a-release)
to publish the provider in the Terraform Registry.

Releases will as drafts. Once marked as published on the GitHub Releases page,
they will become available via the Terraform Registry.
