---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_instance_log"
---

# Data Source: scaleway_rdb_instance_log

Gets information about a specific Scaleway Managed Database Instance log.

Refer to the Managed Databases for PostgreSQL and MySQL [documentation](https://www.scaleway.com/en/docs/managed-databases-for-postgresql-and-mysql/) and [API documentation](https://www.scaleway.com/en/developers/api/managed-databases-for-postgresql-and-mysql/) for more information.

## Example Usage

```terraform
data "scaleway_rdb_instance_log" "by_id" {
  instance_log_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `instance_log_id` - (Required) The ID of the Database Instance log. Can be a plain UUID or a regional ID (`{region}/{id}`).
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which the log exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Database Instance log, in the `{region}/{id}` format.
- `status` - Status of the log (`unknown`, `ready`, `creating`, `error`).
- `node_name` - Name of the underlying node.
- `download_url` - Presigned Object Storage URL to download the log file.
- `created_at` - Creation date of the log (RFC 3339 format).
- `expires_at` - Expiration date of the log (RFC 3339 format).
