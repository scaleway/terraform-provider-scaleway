---
subcategory: "Redis"
page_title: "Scaleway: scaleway_redis_instance"
---

# scaleway_redis_cluster

Gets information about a Redis™ cluster.

For further information refer to the Managed Database for Redis™ [API documentation](https://developers.scaleway.com/en/products/redis/api/v1alpha1/#clusters-a85816).

## Example Usage

```hcl
# Get info by name
data "scaleway_redis_cluster" "my_cluster" {
  name = "foobar"
}

# Get info by cluster ID
data "scaleway_redis_cluster" "my_cluster" {
  cluster_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the Redis cluster.

- `cluster_id` - (Optional) The Redis cluster ID.

  -> **Note** You must specify at least one: `name` and/or `cluster_id`.

- `zone` - (Default to [provider](../index.md) `region`) The [zone](../guides/regions_and_zones.md#zones) in which the server exists.

- `project_id` - (Optional) The ID of the project the Redis cluster is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Redis cluster.
- `version` - Redis's Cluster version (e.g. `6.2.7`).
- `user_name` -  The first user of the Redis Cluster.
- `password` - Password of the first user of the Redis Cluster.
- `created_at` - The date and time of creation of the Redis Cluster.
- `updated_at` - The date and time of the last update of the Redis Cluster.
- `cluster_size` - The number of nodes in the Redis Cluster.
- `node_type` - The type of Redis Cluster (e.g. `RED1-M`).
- `public_network` - Public network details.
- `private_network` - List of private networks endpoints of the Redis Cluster.
- `endpoint_id` - The ID of the endpoint.
- `tls_enabled` -  Whether TLS is enabled or not.
- `acl` - List of acl rules.
- `settings` -  Map of settings for redis cluster.
- `certificate` - The PEM of the certificate used by redis, only when `tls_enabled` is true.
- `tags` - The tags associated with the Redis Cluster.


~> **Important:** Redis™ cluster IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`
