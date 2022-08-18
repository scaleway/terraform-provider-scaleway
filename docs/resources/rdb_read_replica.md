---
page_title: "Scaleway: scaleway_rdb_read_replica"
description: |-
Manages Scaleway Database read replicas.
---

# scaleway_rdb_instance

Creates and manages Scaleway Database read replicas.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api).

## Examples

### Basic

```hcl
resource scaleway_rdb_instance instance {
  name = "test-rdb-rr-update"
  node_type = "db-dev-s"
  engine = "PostgreSQL-14"
  is_ha_cluster = false
  disable_backup = true
  user_name = "my_initial_user"
  password = "thiZ_is_v&ry_s3cret"
  tags = [ "terraform-test", "scaleway_rdb_read_replica", "minimal" ]
}

resource "scaleway_rdb_read_replica" "replica" {
  instance_id = scaleway_rdb_instance.instance.id
  direct_access {}
}
```

### Private network

```hcl
resource scaleway_rdb_instance instance {
  name = "rdb_instance"
  node_type = "db-dev-s"
  engine = "PostgreSQL-14"
  is_ha_cluster = false
  disable_backup = true
  user_name = "my_initial_user"
  password = "thiZ_is_v&ry_s3cret"
}

resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_rdb_read_replica" "replica" {
  instance_id = scaleway_rdb_instance.instance.id
  private_network {
    private_network_id = scaleway_vpc_private_network.pn.id
    service_ip         = "192.168.1.254/24"
  }
}
```

## Arguments Reference

The following arguments are supported:

- `instance_id` - (Required) Id of the rdb instance to replicate.

~> **Important:** When creating a replica, it musts contains at least one of direct_access or private_network. It can contain both.

- `direct_access` - (Optional) Creates a direct access endpoint to rdb replica.

- `private_network` - (Optional) Create an endpoint in a private network.
    - `private_network_id` - (Required) UUID of the private network to be connected to the read replica.
    - `service_ip` - (Required) Endpoint IPv4 address with a CIDR notation. Check documentation about IP and subnet limitations. (IP network).

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the Database read replica should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Database read replica.
- `read_replicas` - List of read replicas of the database instance.
    - `ip` - IP of the replica.
    - `port` - Port of the replica.
    - `name` - Name of the replica.
- `direct_access` - List of load balancer endpoints of the database instance.
    - `endpoint_id` - The ID of the endpoint of the read replica.
    - `ip` - IPv4 address of the endpoint (IP address). Only one of ip and hostname may be set.
    - `port` - TCP port of the endpoint.
    - `name` - Name of the endpoint.
    - `hostname` - Hostname of the endpoint. Only one of ip and hostname may be set.
- `private_network` - List of private networks endpoints of the database instance.
    - `endpoint_id` - The ID of the endpoint of the read replica.
    - `ip` - IPv4 address of the endpoint (IP address). Only one of ip and hostname may be set.
    - `port` - TCP port of the endpoint.
    - `name` - Name of the endpoint.
    - `hostname` - Hostname of the endpoint. Only one of ip and hostname may be set.


## Import

Database Instance can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_rdb_read_replica.rr fr-par/11111111-1111-1111-1111-111111111111
```
