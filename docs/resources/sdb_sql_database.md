---
subcategory: "Databases"
page_title: "Scaleway: scaleway_sdb_sql_database"
---

# scaleway_sdb_sql_database

Creates and manages Scaleway Serverless SQL Databases. For more information, see [the documentation](https://www.scaleway.com/en/developers/api/serverless-databases/).

## Example Usage

### Basic

```hcl
resource "scaleway_sdb_sql_database" "database" {
  name = "my-database"
  min_cpu = 0
  max_cpu = 8
}
```

### With IAM Token

#### TODO

## Arguments Reference

The following arguments are supported:

- `name` - (Required) Name of the database (e.g. `my-new-database`).

    ~> **Important:** Updates to `name` will recreate the Database.

- `min_cpu` - (Optional) The minimum number of CPU units for your Database. Defaults to 0.
- `max_cpu` - (Optional) The maximum number of CPU units for your Database. Defaults to 15.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the resource exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the database, which is of the form `{region}/{id}` e.g. `fr-par/11111111-1111-1111-1111-111111111111`
- `endpoint` - Endpoint of the database

## Import

RDB Database can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_sdb_sql_database.database fr-par/11111111-1111-1111-1111-111111111111
```
