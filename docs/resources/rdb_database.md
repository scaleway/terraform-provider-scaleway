---
page_title: "Scaleway: scaleway_rdb_database"
description: |-
  Manages Scaleway RDB Database.
---

# scaleway_rdb_database

Creates and manages Scaleway RDB database.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api).

## Examples

### Basic

```hcl
resource "scaleway_rdb_database" "main" {
  instance_id    = scaleway_rdb_instance.main.id
  name           = "my-new-database"
}
```

## Arguments Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the instance where to create the database.

~> **Important:** Updates to `instance_id` will recreate the Database.

- `name` - (Required) Name of the database (e.g. `my-new-database`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `owner` - The name of the owner of the database.
- `managed` - Whether or not the database is managed or not.
- `size` - Size of the database (in bytes).

## Import

RDB Database can be imported using the `{region}/{id}/{DBNAME}`, e.g.

```bash
$ terraform import scaleway_rdb_database.rdb01_mydb fr-par/11111111-1111-1111-1111-111111111111/mydb
```
