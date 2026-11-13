---
subcategory: "Databases"
page_title: "Scaleway: scaleway_sdb_sql_versions"
---

# scaleway_sdb_sql_versions

Gets information about available Serverless SQL Database versions.

For further information refer to the Serverless SQL Databases [API documentation](https://www.scaleway.com/en/developers/api/serverless-databases/).

## Example Usage

```terraform
data "scaleway_sdb_sql_versions" "pg" {
  name = "16"
}

locals {
  selected_version = data.scaleway_sdb_sql_versions.pg.versions[0].name
}
```

## Argument Reference

- `region` - (Optional, Computed, Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which to list versions.

- `name` - (Optional) Filter Serverless SQL Database versions that match a given name pattern (e.g. `16`).

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the data source (the region).

- `versions` - List of available Serverless SQL Database versions.
    - `name` - The major version of the PostgreSQL engine.
    - `end_of_life_at` - End of life date of the version (RFC3339).
    - `srn` - The Scaleway Resource Name (SRN) of the version.
