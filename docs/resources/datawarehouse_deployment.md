---
subcategory: "Data Warehouse"
page_title: "Scaleway: scaleway_datawarehouse_deployment"
---

# Resource: scaleway_datawarehouse_deployment

Creates and manages Scaleway Data Warehouse deployments.
For more information refer to the [product documentation](https://www.scaleway.com/en/docs/data-warehouse/).

## Example Usage

### Basic

```terraform
resource "scaleway_datawarehouse_deployment" "main" {
  name           = "my-datawarehouse"
  version        = "v25"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  password       = "thiZ_is_v&ry_s3cret"
}
```

### With Tags

```terraform
resource "scaleway_datawarehouse_deployment" "main" {
  name           = "my-datawarehouse"
  version        = "v25"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  password       = "thiZ_is_v&ry_s3cret"
  tags           = ["production", "analytics"]
}
```

### With Private Network

```terraform
resource "scaleway_vpc" "main" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_datawarehouse_deployment" "main" {
  name           = "my-datawarehouse"
  version        = "v25"
  replica_count  = 1
  cpu_min        = 2
  cpu_max        = 4
  ram_per_cpu    = 4
  password       = "thiZ_is_v&ry_s3cret"

  private_network {
    pn_id = scaleway_vpc_private_network.pn.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the Data Warehouse deployment.
- `version` - (Required, Forces new resource) ClickHouse version to use (e.g., "v25"). Changing this forces recreation of the deployment.
- `replica_count` - (Required) Number of replicas. Can be updated in place via the deployment configuration API.
- `cpu_min` - (Required) Minimum CPU count (autoscaling lower bound). Must be less than or equal to `cpu_max`. Can be updated in place.
- `cpu_max` - (Required) Maximum CPU count (autoscaling upper bound). Must be greater than or equal to `cpu_min`. Can be updated in place.
- `ram_per_cpu` - (Required) RAM per CPU in GB.
- `started` - (Optional, defaults to `true`) Whether the deployment should be running. When set to `false`, the provider calls the Stop deployment API after create or update; when set to `true`, it calls Start deployment if the deployment is stopped. Scaling fields (`replica_count`, `cpu_min`, `cpu_max`) require the deployment to be running; if it is stopped, the provider starts it to apply the change, then stops it again when `started` is `false`.
- `password` - (Optional) Password for the first user of the deployment. If not specified, a random password will be generated. Only one of `password` or `password_wo` should be specified. Note: plain `password` is only used during deployment creation; it is not rotated on update.
- `password_wo` - (Optional) Password for the first user of the deployment in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`. Updates are applied via the Users API to the initial user (an administrator when present, otherwise the first user by name).
- `password_wo_version` - (Optional) The version of the [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) password. To update the `password_wo`, you must also update the `password_wo_version`.
- `tags` - (Optional) List of tags to apply to the deployment.
- `private_network` - (Optional, Forces new resource) Private network configuration to expose your deployment. Changing this forces recreation of the deployment.
    - `pn_id` - (Required) The ID of the private network. Format: `{region}/{id}` or just `{id}`.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the deployment should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the deployment is associated with.

~> **Note:** A public endpoint is always created automatically alongside any private network configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the deployment.
- `status` - The status of the deployment (e.g., "ready", "provisioning").
- `created_at` - Date and time of deployment creation (RFC 3339 format).
- `updated_at` - Date and time of deployment last update (RFC 3339 format).
- `public_network` - Public endpoint information (always created automatically).
    - `id` - The ID of the public endpoint.
    - `dns_record` - DNS record for the public endpoint.
    - `services` - List of services exposed on the public endpoint.
        - `protocol` - Service protocol (e.g., "tcp", "https", "mysql").
        - `port` - TCP port number.
- `private_network` - Private endpoint information (only present if configured).
    - `pn_id` - The ID of the private network.
    - `id` - The ID of the private endpoint.
    - `dns_record` - DNS record for the private endpoint.
    - `services` - List of services exposed on the private endpoint.
        - `protocol` - Service protocol (e.g., "tcp", "https", "mysql").
        - `port` - TCP port number.

## Import

Data Warehouse deployments can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_datawarehouse_deployment.main fr-par/11111111-1111-1111-1111-111111111111
```
