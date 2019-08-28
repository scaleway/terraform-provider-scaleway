---
layout: "scaleway"
page_title: "Migration Guide"
description: |-
  Migrating your Scaleway provider from v1 to v2.
---

# Migrating from v1 to v2

This page guides you through the process of migrating your version 1 resources to their version 2 equivalent.
To prepare the launch of all new Scaleway products, we completely changed the naming of all resources (as well as their attributes) in version 2 of the Terraform provider.

## Configuration

In order to unify configuration management across all scaleway developer tools, we changed the configuration management in version 2.

Below you find an overview of changes in the provider config:

| Old provider argument | New provider argument |
| --------------------- | --------------------- |
| `access_key`          | `access_key`          |
| `token`               | `secret_key`          |
| `organization`        | `project_id`          |

~> **Important:** `access_key` should now only be used for your access key (e.g. `SCWZFD9BPQ4TZ14SM1YS`). Your secret key (previously known as *token*) must be set in `secret_key` (`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`).

Below you find an overview of the changes in environment variables:

| Old env variable        | New env variable                            |
| ----------------------- | ------------------------------------------- |
| `SCALEWAY_ACCESS_KEY`   | `SCW_ACCESS_KEY`                            |
| `SCALEWAY_TOKEN`        | `SCW_SECRET_KEY`                            |
| `SCALEWAY_ORGANIZATION` | `SCW_DEFAULT_PROJECT_ID`                    |
| `SCALEWAY_REGION`       | `SCW_DEFAULT_REGION` and `SCW_DEFAULT_ZONE` |
| `SCW_TLSVERIFY`         | `SCW_INSECURE`                              |
| `SCW_ORGANIZATION`      | `SCW_DEFAULT_PROJECT_ID`                    |
| `SCW_REGION`            | `SCW_DEFAULT_REGION`                        |
| `SCW_TOKEN`             | `SCW_SECRET_KEY`                            |

~> **Important:** `SCALEWAY_ACCESS_KEY` was changed to `SCW_ACCESS_KEY`. This should be your access key (e.g. `SCWZFD9BPQ4TZ14SM1YS`). Your secret key (previously known as *token*) must be set in `SCW_SECRET_KEY` (`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`).

## Resources

All resources are from now on prefixed by `scaleway`, their product category and their product name (`scaleway_{product-category-name}_{product-name}_{resource-name}`). For instances an S3 bucket belongs to the `Storage` product category and is a resource of the `Object` product. Hence it is named: `scaleway_storage_object_bucket`.

### Compute

All the old compute resources have been regrouped under a new name: `Instance`. `Compute` is now the general product category name of all server related resources. 
This means that all old resources are now prefixed with `scaleway_compute_instance_`.

#### Renamed: `scaleway_server` -> `scaleway_instance_server`

`scaleway_server` was renamed to `scaleway_instance_server`.

In version 1, attachments of volumes where done on the volume resource. But from now on, this is done on the `scaleway_instance_server` resource.

Thus, to create a server with a volume attached:

```hcl
resource "scaleway_compute_instance_volume" "data" {
  size_in_gb = 100
}

resource "scaleway_instance_server" "web" {
  type = "DEV1-L"
  image_id = "f974feac-abae-4365-b988-8ec7d1cec10d"

  tags = [ "hello", "public" ]

  root_volume {
    delete_on_termination = false
  }

  additional_volume_ids = [ "${scaleway_compute_instance_volume.data.id}" ]
}
```

#### Renamed: `scaleway_ip` -> `scaleway_instance_ip`

`scaleway_ip` was renamed to `scaleway_instance_ip` and the argument `server` was renamed to `server_id`.

```hcl
resource "scaleway_instance_ip" "test_ip" {
  server_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

#### Renamed: `scaleway_volume` -> `scaleway_compute_instance_volume`

`scaleway_volume` was renamed to `scaleway_compute_instance_volume`.
The former arguments can still be used on the new volume resource.

Additionally, from now on, you can also create new volumes based on other volumes or snapshots. For more information check the [new volume `scaleway_compute_instance_volume` resource](../r/compute_instance_volume.html).

#### Removed: `scaleway_user_data`

`scaleway_user_data` is now part of the `scaleway_instance_server` resource.

#### Removed: `scaleway_token`

The `scaleway_token` was removed in version 2.

Tokens should be created in the console.

#### Removed: `scaleway_ssh_key`

The `scaleway_ssh_key` was removed in version 2.

SSH keys should be uploaded in the console.

#### Removed: `scaleway_ip_reverse_dns`

The `scaleway_ip_reverse_dns` was removed in version 2.

Reverse DNS must be set on the IP resource itself:

```hcl
resource "scaleway_instance_ip" "test_ip" {
  reverse = "scaleway.com"
}
```

#### Removed: `scaleway_volume_attachment`

The `scaleway_volume_attachment` was removed in version 2.

Volumes can in version 2 only be attached on the server resource. The [above example](#scaleway_server-gt-scaleway_instance_server) shows how this works.

### Storage

#### Renamed: `scaleway_bucket` -> `scaleway_storage_object_bucket`

The `scaleway_bucket` was moved to the `object` product in the `storage` product category.

It's behaviour remained the same, but we also added an [`acl` argument](../r/storage_object_bucket.html#acl). This argument takes canned ACLs.