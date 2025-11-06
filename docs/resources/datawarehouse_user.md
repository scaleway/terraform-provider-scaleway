---
subcategory: "Data Warehouse"
page_title: "Scaleway: scaleway_datawarehouse_user"
---

# Resource: scaleway_datawarehouse_user

Creates and manages Scaleway Data Warehouse users within a deployment.
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

resource "scaleway_datawarehouse_user" "main" {
  deployment_id = scaleway_datawarehouse_deployment.main.id
  name          = "my_user"
  password      = "user_password_123"
}
```

### Admin User

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

resource "scaleway_datawarehouse_user" "admin" {
  deployment_id = scaleway_datawarehouse_deployment.main.id
  name          = "admin_user"
  password      = "admin_password_456"
  is_admin      = true
}
```

## Argument Reference

The following arguments are supported:

- `deployment_id` - (Required) ID of the Data Warehouse deployment to which this user belongs.
- `name` - (Required) Name of the ClickHouse user.
- `password` - (Required) Password for the ClickHouse user.
- `is_admin` - (Optional) Whether the user has administrator privileges. Defaults to `false`.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the user should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the user (format: `{region}/{deployment_id}/{name}`).

## Import

Data Warehouse users can be imported using the `{region}/{deployment_id}/{name}`, e.g.

```bash
terraform import scaleway_datawarehouse_user.main fr-par/11111111-1111-1111-1111-111111111111/my_user
```

