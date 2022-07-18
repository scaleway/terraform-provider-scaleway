---
page_title: "Scaleway: scaleway_redis_cluster"
description: |-
Manages Scaleway Redis Clusters.
---

# scaleway_redis_cluster

Creates and manages Scaleway Redis Clusters.
For more information, see [the documentation](https://developers.scaleway.com/en/products/redis/api/v1alpha1/).

## Examples

### Basic

```hcl
resource "scaleway_redis_cluster" "main" {
  name = "test_redis_basic"
  version = "6.2.6"
  node_type = "MDB-BETA-M"
  user_name = "my_initial_user"
  password = "thiZ_is_v&ry_s3cret"
  tags = [ "test", "redis" ]
  cluster_size = 1
  tls_enabled = "true"

  acl {
    ip = "0.0.0.0/0"
    description = "Allow all"
  }
}
```

### With settings

```hcl
resource "scaleway_redis_cluster" "main" {
  name = "test_redis_basic"
  version = "6.2.6"
  node_type = "MDB-BETA-M"
  user_name = "my_initial_user"
  password = "thiZ_is_v&ry_s3cret"

  settings = {
    "maxclients" = "1000"
    "tcp-keepalive" = "120"
  }
}
```

### With a private network

```hcl
resource "scaleway_vpc_private_network" "pn" { 
  name = "private-network"
}

resource "scaleway_redis_cluster" "main" {
  name         = "test_redis_endpoints"
  version      = "6.2.6"
  node_type    = "MDB-BETA-M"
  user_name    = "my_initial_user"
  password     = "thiZ_is_v&ry_s3cret"
  cluster_size = 1
  private_network {
    id          = "${scaleway_vpc_private_network.pn.id}"
    service_ips = [
      "10.12.1.1/20",
    ]
  }
  depends_on = [
    scaleway_vpc_private_network.pn
  ]
}
```

## Arguments Reference

The following arguments are supported:

- `version` - (Required) Redis's Cluster version (e.g. `6.2.6`).

~> **Important:** Updates to `version` will migrate the Redis Cluster to the desired `version`. Keep in mind that you cannot downgrade a Redis Cluster.

- `node_type` - (Required) The type of Redis Cluster you want to create (e.g. `MDB-BETA-M`).

~> **Important:** Updates to `node_type` will migrate the Redis Cluster to the desired `node_type`. Keep in mind that you cannot downgrade a Redis Cluster.

- `user_name` - (Required) Identifier for the first user of the Redis Cluster.

- `password` - (Required) Password for the first user of the Redis Cluster.

- `name` - (Optional) The name of the Redis Cluster.

- `tags` - (Optional) The tags associated with the Redis Cluster.

- `zone` - (Defaults to [provider](../index.md) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Redis Cluster should be created.

- `cluster_size` - (Optional) The number of nodes in the Redis Cluster.

~> **Important:** You cannot set `cluster_size` to 2, you either have to choose Standalone mode (1 node) or Cluster mode which is minimum 3 (1 main node + 2 secondary nodes)

~> **Important:** You can set a bigger `cluster_size` than you initially did, it will migrate the Redis Cluster, but keep in mind that you cannot downgrade a Redis Cluster so setting a smaller `cluster_size` will not have any effect.

- `tls_enabled` - (Defaults to false) Whether TLS is enabled or not.

- `project_id` - (Defaults to [provider](../index.md) `project_id`) The ID of the project the Redis Cluster is associated with.

- `acl` - (Optional) List of acl rules, this is cluster's authorized IPs.

  ~> The `acl` block supports:

    - `ip` - (Required) The ip range to whitelist in [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation)
    - `description` - (Optional) A text describing this rule. Default description: `Allow IP`

- `settings` - (Optional) Map of settings for redis cluster. Available settings can be found by listing redis versions with scaleway API or CLI

- `private_network` - (Optional) Describes the private network you want to connect to your cluster. If not set, a public network will be provided.

  ~> The `private_network` block supports :

    - `id` - (Required) The UUID of the private network resource.
    - `service_ips` - (Required) Endpoint IPv4 addresses in [CIDR notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing#CIDR_notation). You must provide at least one IP per node.

~> **Important:** The way to use private networks differs whether you are using redis in standalone or cluster mode.

- Standalone mode (`cluster_size` = 1) : you can attach as many private networks as you want (each must be a separate block). If you detach your only private network, your cluster won't be reachable until you define a new private or public network. You can modify your private_network and its specs, you can have both a private and public network side by side.

- Cluster mode (`cluster_size` > 1) : you can define a single private network as you create your cluster, you won't be able to edit or detach it afterwards, unless you create another cluster. Your `service_ips` must be listed as follows:

```hcl
  service_ips = [
    "10.12.1.10/20",
    "10.12.1.11/20",
    "10.12.1.12/20",
]
```

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Database Instance.
- `created_at` - The date and time of creation of the Redis Cluster.
- `updated_at` - The date and time of the last update of the Redis Cluster.


## Import

Redis Cluster can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_redis_cluster.redis01 fr-par/11111111-1111-1111-1111-111111111111
```
