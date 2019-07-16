---
layout: "scaleway"
page_title: "Provider: Scaleway"
description: |-
  The Scaleway provider is used to manage Scaleway resources. The provider needs to be configured with the proper credentials before it can be used.
---

# Scaleway Provider

The Scaleway provider is used to manage Scaleway resources.
The provider needs to be configured with the proper credentials before it can be used.

**This is the documentation for the version `>= 1.11.0` of the provider. If you come from `< v1.11.0`, checkout to [migration guide](./guides/migration_guide.html).**

Use the navigation to the left to read about the available resources.

## Example

Here is an example that will setup a web server with an additional volume, a public IP and a security group.

You can test this config by creating a `test.tf` and run terraform commands from this directory:

- Get your [Scaleway credentials](https://console.scaleway.com/account/credentials) 
- Initialize a Terraform working directory: `terraform init`
- Generate and show the execution plan: `terraform plan`
- Build the infrastructure: `terraform apply`

```hcl
provider "scaleway" {
  access_key = "<SCALEWAY-ACCESS-KEY>"
  secret_key = "<SCALEWAY-SECRET-KEY>"
  project_id = "<SCALEWAY-PROJECT-ID>" # aka. Organization ID
  zone       = "fr-par-1"
  region     = "fr-par"
}

resource "scaleway_compute_instance_ip" "public_ip" {
  server_id = "${scaleway_compute_instance_server.web.id}"
}

resource "scaleway_compute_instance_volume" "data" {
  size_in_gb = 100
}

resource "scaleway_compute_instance_security_group" "www" {
  inbound_default_policy = "drop"
  outbound_default_policy = "accept"

  inbound {
    action = "accept"
    port = "22"
    ip = "212.47.225.64"
  }

  inbound {
    action = "accept"
    port = "80"
  }

  inbound {
    action = "accept"
    port = "443"
  }
}

resource "scaleway_compute_instance_server" "web" {
  type = "DEV1-L"
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"

  tags = [ "front", "web" ]

  additional_volume_ids = [ "${scaleway_compute_instance_volume.data.id}" ]

  security_group_id= "${scaleway_compute_instance_security_group.www.id}"
}
```

## Authentication

The Scaleway authentication is based on an **access key** and a **secret key**.
Since secret keys are only revealed one time (when it is first created) you might
need to create a new one in the section "API Tokens" of the
[Scaleway console](https://console.scaleway.com/account/credentials).
Click on the "Generate new token" button to create them. Giving it a friendly-name is recommended.

The Scaleway provider offers three ways of providing these credentials. The following methods are supported, in this priority order:

1. [Static credentials](#static-credentials)
2. [Environment variables](#environment-variables)
3. [Shared configuration file](#shared-configuration-file)

### Static credentials

!> **Warning**: Hard-coding credentials into any Terraform configuration is not recommended, and risks secret leakage should this file ever be committed to a public version control system.

Static credentials can be provided by adding `access_key` and `secret_key` attributes in-line in the Scaleway provider block:

Example:

```hcl
provider "scaleway" {
  access_key = "my-access-key"
  secret_key = "my-secret-key"
}
```

### Environment variables

You can provide your credentials via the `SCW_ACCESS_KEY`, `SCW_SECRET_KEY` environment variables.

Example:

```hcl
provider "scaleway" {}
```

Usage:

```bash
$ export SCW_ACCESS_KEY="my-access-key"
$ export SCW_SECRET_KEY="my-secret-key"
$ terraform plan
```

### Shared configuration file

It is a YAML configuration file shared between the majority of the
[Scaleway developer tools](https://developers.scaleway.com/en/community-tools/#official-repos).
Its default location is `$HOME/.config/scw/config.yaml` (`%USERPROFILE%/.config/scw/config.yaml` on Windows).
If it fails to detect credentials inline, or in the environment, Terraform will check this file.

You can optionally specify a different location with `SCW_CONFIG_PATH` environment variable.
You can find more information about this configuration [in the documentation](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md#scaleway-config).

## Arguments Reference

In addition to [generic provider arguments](https://www.terraform.io/docs/configuration/providers.html) (e.g. `alias` and `version`), the following arguments are supported in the Scaleway provider block:

- `access_key` - (Optional) The Scaleway access key. It must be provided, but it can also be sourced from
the `SCW_ACCESS_KEY` [environment variable](#environment-variables), or via a [shared configuration file](#shared-configuration-file),
in this priority order.

- `secret_key` - (Optional) The Scaleway secert key. It must be provided, but it can also be sourced from
the `SCW_SECRET_KEY` [environment variable](#environment-variables), or via a [shared configuration file](#shared-configuration-file),
in this priority order.

- `project_id` - (Optional) The project ID that will be used as default value for all resources. It can also be sourced from
the `SCW_DEFAULT_PROJECT_ID` [environment variable](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md#environment-variables), or via a [shared configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md#scaleway-config),
in this priority order.

- `region` - (Optional) The [region](./guides/regions_and_zones.html#regions)  that will be used as default value for all resources. It can also be sourced from
the `SCW_DEFAULT_REGION` [environment variable](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md#environment-variables), or via a [shared configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md#scaleway-config),
in this priority order.

- `zone` - (Optional) The [zone](./guides/regions_and_zones.html#zones) that will be used as default value for all resources. It can also be sourced from
the `SCW_DEFAULT_ZONE` [environment variable](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md#environment-variables), or via a [shared configuration file](https://github.com/scaleway/scaleway-sdk-go/blob/master/scw/README.md#scaleway-config),
in this priority order.

## Scaleway S3-compatible

[Scaleway object storage](https://www.scaleway.com/object-storage/) can be used to store your Terraform state.
Configure your backend as:

```
terraform {
  backend "s3" {
    bucket      = "terraform_state"
    key         = "my_state.tfstate"
    region      = "fr-par"
    endpoint    = "https://s3.fr-par.scw.cloud"
    access_key = "my-access-key"
    secret_key = "my-secret-key"
    skip_credentials_validation = true
    skip_region_validation = true
  }
}
```

Beware as no locking mechanism are yet supported.
Using scaleway object storage as terraform backend is not suitable if you work in a team with à risk of simultaneously access to the same plan.
