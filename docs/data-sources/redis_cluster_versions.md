---
subcategory: "Redis"
page_title: "Scaleway: scaleway_redis_cluster_versions"
---

# scaleway_redis_cluster_versions

Gets information about available Redis™ cluster versions.

For further information refer to the Managed Database for Redis™ [API documentation](https://developers.scaleway.com/en/products/redis/api/v1alpha1/#clusters-a85816).

## Example Usage

```terraform
data "scaleway_redis_cluster_versions" "redis" {
  zone               = "fr-par-1"
  include_disabled   = false
  include_beta       = false
  include_deprecated = false
}

locals {
  selected_version = data.scaleway_redis_cluster_versions.redis.versions[0].version
}
```

## Argument Reference

- `zone` - (Optional, Computed, Defaults to [provider](../index.md#arguments-reference) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which to list cluster versions.

- `include_disabled` - (Optional) Whether to include disabled Redis™ engine versions. Defaults to `false`.

- `include_beta` - (Optional) Whether to include beta Redis™ engine versions. Defaults to `false`.

- `include_deprecated` - (Optional) Whether to include deprecated Redis™ engine versions. Defaults to `false`.

- `version` - (Optional) Filter Redis™ engine versions that match a given name pattern.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the data source (the zone).

- `versions` - List of available Redis™ cluster versions.
    - `version` - Redis™ engine version.
    - `end_of_life_at` - End of life date of the version (RFC3339).
    - `released_at` - Release date of the version (RFC3339).
    - `logo_url` - URL of the Redis™ logo.
