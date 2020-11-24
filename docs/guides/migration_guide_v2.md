---
page_title: "Migration Guide"
description: |-
  Migrating your Scaleway provider from v1 to v2.
---

# Migrating from v1 to v2

-> **Note:** The version 2 is not released yet but versions `v1.11+` allow you to do a smooth migration to the `v2`.
In other words, there will be no breaking change between `v1.11+` and `v2`.
The `v2` roadmap is available [here](https://github.com/terraform-providers/terraform-provider-scaleway/issues/125).

This page guides you through the process of migrating your version 1 resources to their version 2 equivalent.
To prepare the launch of all new Scaleway products, we completely changed the naming of all resources (as well as their attributes) in version 2 of the Terraform provider.

## Provider

### Version configuration

-> **Note:** Before upgrading to `v2+`, it is recommended to upgrade to the most recent `1.X` version of the provider (`v1.11.0`) and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html) without unexpected change or deprecation notice.

It is recommended to use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#version-provider-versions). If you are following these recommendation, update the version constraints in your Terraform configuration and run [`terraform init`](https://www.terraform.io/docs/commands/init.html) to download the new version.

Update to latest `1.X` version:

```hcl
provider "scaleway" {
  # ... other configuration ...

  version = "~> 1.11"
}
```

Update to latest 2.X version:

```hcl
provider "scaleway" {
  # ... other configuration ...

  version = "~> 2.0"
}
```

### Provider configuration

In order to unify configuration management across all scaleway developer tools, we changed the configuration management in version 2.

Below you find an overview of changes in the provider config:

| Old provider attribute | New provider attribute |
| --------------------- | --------------------- |
| `access_key`          | `access_key`          |
| `token`               | `secret_key`          |
| `organization`        | `organization_id`     |

~> **Important:** `access_key` should now only be used for your access key (e.g. `SCWZFD9BPQ4TZ14SM1YS`).
Your secret key (previously known as _token_) must be set in `secret_key` (`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`).

Below you find an overview of the changes in environment variables:

| Old env variable        | New env variable                            |
| ----------------------- | ------------------------------------------- |
| `SCALEWAY_ACCESS_KEY`   | `SCW_ACCESS_KEY`                            |
| `SCALEWAY_TOKEN`        | `SCW_SECRET_KEY`                            |
| `SCALEWAY_ORGANIZATION` | `SCW_DEFAULT_ORGANIZATION_ID`               |
| `SCALEWAY_REGION`       | `SCW_DEFAULT_REGION` and `SCW_DEFAULT_ZONE` |
| `SCW_TLSVERIFY`         | `SCW_INSECURE`                              |
| `SCW_ORGANIZATION`      | `SCW_DEFAULT_ORGANIZATION_ID`               |
| `SCW_REGION`            | `SCW_DEFAULT_REGION`                        |
| `SCW_TOKEN`             | `SCW_SECRET_KEY`                            |

~> **Important:** `SCALEWAY_ACCESS_KEY` was changed to `SCW_ACCESS_KEY`.
This should be your access key (e.g. `SCWZFD9BPQ4TZ14SM1YS`).
Your secret key (previously known as _token_) must be set in `SCW_SECRET_KEY` (`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`).

Terraform can also read standard Scaleway configuration files.
By doing so, you can use the same configuration between different tools such as the [CLI](https://github.com/scaleway/scaleway-cli) or [Packer](https://www.packer.io/docs/builders/scaleway).

## Resources

All resources are from now on prefixed by `scaleway`, their product category and their product name (`scaleway_{product-category-name}_{product-name}_{resource-name}`).
For instances an S3 bucket belongs to the `Storage` product category and is a resource of the `Object` product.
Hence it is named: `scaleway_object_bucket`.

### How can I migrate from existing code?

Because the resources changed their name, we cannot using automatic state migration.
We will first manually remove the resource from the terraform state and then use [`terraform import`](https://www.terraform.io/docs/import/usage.html) to import existing resources to a renamed resource.

For instance, let's suppose that you have resource in `fr-par-1` such as:

```hcl-terraform
provider "scaleway" {
    zone= "fr-par-1"
}

resource scaleway_server main {
  name  = "foobar"
  type  = "DEV1-S"
  image = "cf44b8f5-77e2-42ed-8f1e-09ed5bb028fc"
}
```

First, let's delete the resource from your terraform state using the [`terraform state`](https://www.terraform.io/docs/commands/state/index.html) command.
You can do it using: `terraform state rm scaleway_server.main`.

Once this is done, refactor your terraform code to:

```hcl-terraform
provider "scaleway" {
    zone= "fr-par-1"
}

resource scaleway_instance_server main {
  name  = "foobar"
  type  = "DEV1-S"
  image = "cf44b8f5-77e2-42ed-8f1e-09ed5bb028fc"
}
```

and run `terraform import scaleway_instance_server.main fr-par-1/11111111-1111-1111-1111-111111111111` where `11111111-1111-1111-1111-111111111111` is the id of your resource.
After importing, you can verify using `terraform apply` that you are in a desired state and that no changes need to be done.

### Instance

All the old instance resources have been regrouped under a new name: `Instance`.
This means that all old instance resources are now prefixed with `scaleway_instance_`.

#### Renamed: `scaleway_server` -> `scaleway_instance_server`

`scaleway_server` was renamed to `scaleway_instance_server`.

In version 1, attachments of volumes where done on the volume resource.
From now on, this is done on the `scaleway_instance_server` resource.

Thus, to create a server with a volume attached:

```hcl
resource "scaleway_instance_volume" "data" {
  size_in_gb = 100
}

resource "scaleway_instance_server" "web" {
  type = "DEV1-L"
  image = "ubuntu_focal"

  tags = [ "hello", "public" ]

  root_volume {
    delete_on_termination = false
  }

  additional_volume_ids = [ scaleway_instance_volume.data.id ]
}
```

#### Renamed: `scaleway_ip` -> `scaleway_instance_ip`

`scaleway_ip` was renamed to `scaleway_instance_ip` and the `server` attribute, used to attach an IP has been moved to `scaleway_instance_server.id_id`

```hcl
resource "scaleway_instance_ip" "test_ip" {
}
```

#### Renamed: `scaleway_volume` -> `scaleway_instance_volume`

`scaleway_volume` was renamed to `scaleway_instance_volume`.
The former attributes can still be used on the new volume resource.

Additionally, from now on, you can also create new volumes based on other volumes or snapshots.
For more information check the [new volume `scaleway_instance_volume` resource](../resources/instance_volume.md).

#### Renamed: `scaleway_ssh_key` -> `scaleway_account_ssk_key`

`scaleway_ssh_key` was renamed to `scaleway_account_ssk_key`
The `key` attribute has been renamed to `public_key`.
A `name` required attribute and an `organization_id` optional attribute have been added.

#### Removed: `scaleway_user_data`

`scaleway_user_data` is now part of the `scaleway_instance_server` resource.

#### Removed: `scaleway_token`

The `scaleway_token` was removed in version 2.

Tokens should be created in the console.

#### Renamed: `scaleway_ip_reverse_dns` -> `scaleway_instance_ip_reverse_dns`

`scaleway_ip_reverse_dns` was renamed to `scaleway_instance_ip_reverse_dns`.

#### Removed: `scaleway_volume_attachment`

The `scaleway_volume_attachment` was removed in version 2.

Volumes can in version 2 only be attached on the server resource.
The [above example](#scaleway_server-gt-scaleway_instance_server) shows how this works.

### Storage

#### Renamed: `scaleway_bucket` -> `scaleway_object_bucket`

The `scaleway_bucket` was moved to the `object` product in the `storage` product category.

It's behaviour remained the same, but we also added an [`acl` attribute](../resources/object_bucket.md#acl).
This attribute takes canned ACLs.
