---
page_title: "Scaleway: scaleway_rdb_instance"
description: |-
  Manages Scaleway Database Instances.
---

# scaleway_rdb_instance

Creates and manages Scaleway Database Instances.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api).

## Examples

### Basic

```hcl
resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-11"
  is_ha_cluster  = true
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
}

# with backup schedule
resource "scaleway_rdb_instance" "main" {
  name          = "test-rdb"
  node_type     = "DB-DEV-S"
  engine        = "PostgreSQL-11"
  is_ha_cluster = true
  user_name     = "my_initial_user"
  password      = "thiZ_is_v&ry_s3cret"
  
  disable_backup = true
  backup_schedule_frequency = 24 # every day
  backup_schedule_retention = 7  # keep it one week
}

# with private network and dhcp configuration
resource scaleway_vpc_private_network pn02 {
    name = "my_private_network"
}

resource scaleway_vpc_public_gateway_dhcp main {
    subnet = "192.168.1.0/24"
}

resource scaleway_vpc_public_gateway_ip main {
}

resource scaleway_vpc_public_gateway main {
    name = "foobar"
    type = "VPC-GW-S"
    ip_id = scaleway_vpc_public_gateway_ip.main.id
}

resource scaleway_vpc_public_gateway_pat_rule main {
    gateway_id = scaleway_vpc_public_gateway.main.id
    private_ip = scaleway_vpc_public_gateway_dhcp.main.address
    private_port = scaleway_rdb_instance.main.private_network.0.port
    public_port = 42
    protocol = "both"
    depends_on = [scaleway_vpc_gateway_network.main, scaleway_vpc_private_network.pn02]
}

resource scaleway_vpc_gateway_network main {
    gateway_id = scaleway_vpc_public_gateway.main.id
    private_network_id = scaleway_vpc_private_network.pn02.id
    dhcp_id = scaleway_vpc_public_gateway_dhcp.main.id
    cleanup_dhcp = true
    enable_masquerade = true
    depends_on = [scaleway_vpc_public_gateway_ip.main, scaleway_vpc_private_network.pn02]
}

resource scaleway_rdb_instance main {
    name = "test-rdb"
    node_type = "db-dev-s"
    engine = "PostgreSQL-11"
    is_ha_cluster = false
    disable_backup = true
    user_name = "my_initial_user"
    password = "thiZ_is_v&ry_s3cret"
    region= "fr-par"
    tags = [ "terraform-test", "scaleway_rdb_instance", "volume", "rdb_pn" ]
    volume_type = "bssd"
    volume_size_in_gb = 10
    private_network {
        ip_net = "192.168.1.254/24" #pool high
        pn_id = "${scaleway_vpc_private_network.pn02.id}"
    }
}
```

## Arguments Reference

The following arguments are supported:

- `node_type` - (Required) The type of database instance you want to create (e.g. `db-dev-s`).

~> **Important:** Updates to `node_type` will upgrade the Database Instance to the desired `node_type` without any interruption. Keep in mind that you cannot downgrade a Database Instance.

- `engine` - (Required) Database Instance's engine version (e.g. `PostgreSQL-11`).

~> **Important:** Updates to `engine` will recreate the Database Instance.

- `volume_type` - (Optional, default to `lssd`) Type of volume where data are stored (`bssd` or `lssd`).

- `volume_size_in_gb` - (Optional) Volume size (in GB) when `volume_type` is set to `bssd`. Must be a multiple of 5000000000.

- `user_name` - (Optional) Identifier for the first user of the database instance.

~> **Important:** Updates to `user_name` will recreate the Database Instance.

- `password` - (Optional) Password for the first user of the database instance.

- `is_ha_cluster` - (Optional) Enable or disable high availability for the database instance.

~> **Important:** Updates to `is_ha_cluster` will recreate the Database Instance.

- `name` - (Optional) The name of the Database Instance.

- `disable_backup` - (Optional) Disable automated backup for the database instance.

- `backup_schedule_frequency` - (Optional) Backup schedule frequency in hours.

- `backup_schedule_retention` - (Optional) Backup schedule retention in days.

- `backup_same_region` - (Optional) Boolean to store logical backups in the same region as the database instance.

- `settings` - (Optional) Map of engine settings to be set. Using this option will override default config. Available settings for your engine can be found on scaleway console or fetched using [rdb engine list route](https://developers.scaleway.com/en/products/rdb/api/#get-1eafb7)

- `tags` - (Optional) The tags associated with the Database Instance.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the Database Instance should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the Database Instance is associated with.

## Private Network

~> **Important:** Updates to `private_network` will recreate the attachment Instance.

- `ip_net` - (Required) The IP network where to con.
- `pn_id` - (Required) The ID of the private network. If not provided it will be randomly generated.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Database Instance.
- `endpoint_ip` - (Deprecated) The IP of the Database Instance.
- `endpoint_port` - (Deprecated) The port of the Database Instance.
- `read_replicas` - List of read replicas of the database instance.
    - `ip` - IP of the replica.
    - `port` - Port of the replica.
    - `name` - Name of the replica.
- `load_balancer` - List of load balancer endpoints of the database instance.
    - `endpoint_id` - The ID of the endpoint of the load balancer.
    - `ip` - IP of the endpoint.
    - `port` - Port of the endpoint.
    - `name` - Name of the endpoint.
    - `hostname` - Name of the endpoint.
- `private_network` - List of private networks endpoints of the database instance.
    - `endpoint_id` - The ID of the endpoint of the private network.
    - `ip` - IP of the endpoint.
    - `port` - Port of the endpoint.
    - `name` - Name of the endpoint.
    - `hostname` - Name of the endpoint.
- `certificate` - Certificate of the database instance.
- `organization_id` - The organization ID the Database Instance is associated with.

## Limitations

The Managed Database product is only compliant with the private network in the default availability zone (AZ).
i.e `fr-par-1`, `nl-ams-1`, `pl-waw-1`. To know more about check our section [How to connect a PostgreSQL and MySQL Database Instance to a Private Network](https://www.scaleway.com/en/docs/managed-databases/postgresql-and-mysql/how-to/connect-database-private-network/)

## Import

Database Instance can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_rdb_instance.rdb01 fr-par/11111111-1111-1111-1111-111111111111
```
