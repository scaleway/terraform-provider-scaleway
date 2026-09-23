---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_database_engines"
---

# scaleway_rdb_database_engines

Gets information about available RDB database engines and versions.

For further information refer to the Managed Databases for PostgreSQL and MySQL [API documentation](https://developers.scaleway.com/en/products/rdb/api/#database-instance)

## Example Usage

```terraform
data "scaleway_rdb_database_engines" "pg" {
  region  = "fr-par"
  name    = "PostgreSQL"
  version = "16"
}

locals {
  selected_engine = data.scaleway_rdb_database_engines.pg.engines[0].versions[0].name
}
```

## Argument Reference

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which to list database engines.

- `name` - (Optional) Filter by database engine name (e.g. `PostgreSQL`, `MySQL`).

- `version` - (Optional) Filter by database engine version (e.g. `16`).

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the data source (the region).

- `engines` - List of available database engines.
    - `name` - Name of the database engine.
    - `logo_url` - URL of the database engine logo.
    - `region` - Region of the database engine.
    - `versions` - Available versions for the database engine.
        - `version` - Database engine version number.
        - `name` - Full engine name including version (e.g. `PostgreSQL-16`).
        - `end_of_life` - End of life date of the engine version (RFC3339).
        - `disabled` - Whether the engine version is disabled and cannot be created.
        - `beta` - Whether the engine version is in beta.
