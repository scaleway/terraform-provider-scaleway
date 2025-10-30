---
subcategory: "Data Warehouse"
page_title: "Scaleway: scaleway_datawarehouse_database"
---

# Resource: scaleway_datawarehouse_database

Creates and manages Scaleway Data Warehouse databases within a deployment.
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

resource "scaleway_datawarehouse_database" "main" {
  deployment_id = scaleway_datawarehouse_deployment.main.id
  name          = "my_database"
}
```

## Argument Reference

The following arguments are supported:

- `deployment_id` - (Required) ID of the Data Warehouse deployment to which this database belongs.
- `name` - (Required) Name of the database.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the database should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the database (format: `{region}/{deployment_id}/{name}`).
- `size` - Size of the database in GB.

## Import

Data Warehouse databases can be imported using the `{region}/{deployment_id}/{name}`, e.g.

```bash
terraform import scaleway_datawarehouse_database.main fr-par/11111111-1111-1111-1111-111111111111/my_database
```

