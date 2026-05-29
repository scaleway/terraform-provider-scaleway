---
subcategory: "MongoDB®"
page_title: "Scaleway: scaleway_mongodb_databases"
---

# scaleway_mongodb_databases

Gets information about databases on a MongoDB® instance.

For further information refer to the Managed Databases for MongoDB® [API documentation](https://developers.scaleway.com/en/products/mongodb/api/)

## Example Usage

```hcl
resource "scaleway_mongodb_instance" "main" {
  name        = "foobar"
  version     = "7.0"
  node_type   = "MGDB-PLAY2-NANO"
  node_number = 1
  user_name   = "my_initial_user"
  password    = "thiZ_is_v&ry_s3cret"
}

data "scaleway_mongodb_databases" "db" {
  instance_id = scaleway_mongodb_instance.main.id
  region      = "fr-par"
}

output "database_names" {
  value = [for database in data.scaleway_mongodb_databases.db.databases : database.name]
}
```

## Argument Reference

- `instance_id` - (Required) The MongoDB® instance ID. Can be a plain UUID or a regional ID.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the MongoDB® instance exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `databases` - List of databases on the MongoDB® instance.
    - `name` - Name of the database.
