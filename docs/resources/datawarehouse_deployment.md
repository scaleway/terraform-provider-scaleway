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

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the Data Warehouse deployment.
- `version` - (Required, Forces new resource) ClickHouse version to use (e.g., "v25"). Changing this forces recreation of the deployment.
- `replica_count` - (Required) Number of replicas.
- `cpu_min` - (Required) Minimum CPU count. Must be less than or equal to `cpu_max`.
- `cpu_max` - (Required) Maximum CPU count. Must be greater than or equal to `cpu_min`.
- `ram_per_cpu` - (Required) RAM per CPU in GB.
- `password` - (Optional) Password for the first user of the deployment. If not specified, a random password will be generated. Note: password is only used during deployment creation.
- `tags` - (Optional) List of tags to apply to the deployment.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the deployment should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the deployment is associated with.

~> **Important:** Private endpoints are not yet supported by the API. A public endpoint is always created automatically.

~> **Note:** During the private beta phase, modifying `cpu_min`, `cpu_max`, and `replica_count` has no effect until the feature is launched in general availability.

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

## Import

Data Warehouse deployments can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_datawarehouse_deployment.main fr-par/11111111-1111-1111-1111-111111111111
```
