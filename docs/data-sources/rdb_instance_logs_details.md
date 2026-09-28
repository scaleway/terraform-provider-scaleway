---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_instance_logs_details"
---

# Data Source: scaleway_rdb_instance_logs_details

Gets remote log details for a Scaleway Managed Database Instance.

Refer to the Managed Databases for PostgreSQL and MySQL [documentation](https://www.scaleway.com/en/docs/managed-databases-for-postgresql-and-mysql/) and [API documentation](https://www.scaleway.com/en/developers/api/managed-databases-for-postgresql-and-mysql/) for more information.

## Example Usage

```terraform
resource "scaleway_rdb_instance" "main" {
  name           = "rdb-instance-logs-details"
  node_type      = "db-dev-s"
  engine         = "PostgreSQL-15"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

data "scaleway_rdb_instance_logs_details" "details" {
  instance_id = scaleway_rdb_instance.main.id
}
```

## Argument Reference

- `instance_id` - (Required) The ID of the Database Instance. Can be a plain UUID or a regional ID (`{region}/{id}`).
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which the Database Instance exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of this data source, in the `{region}/{instance_id}` format.
- `details` - Remote Database Instance logs details.
  - `log_name` - Name of the remote log.
  - `size` - Size of the remote log in bytes.
