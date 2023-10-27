---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_instance"
---

# scaleway_rdb_instance

Creates and manages Scaleway Database Instances.
For more information, see [the documentation](https://developers.scaleway.com/en/products/rdb/api).

## Examples

### Example Basic

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
```

### Example With IPAM

```hcl
resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-11"
  is_ha_cluster  = true
  disable_backup = true
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
  private_network {
    pn_id = scaleway_vpc_private_network.pn.id
  }
}
```

### Example with Settings

```hcl
resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "db-dev-s"
  disable_backup = true
  engine         = "MySQL-8"
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
  init_settings  = {
    "lower_case_table_names" = 1
  }
  settings = {
    "max_connections" = "350"
  }
}
```

### Example with backup schedule

```hcl
resource "scaleway_rdb_instance" "main" {
  name          = "test-rdb"
  node_type     = "DB-DEV-S"
  engine        = "PostgreSQL-11"
  is_ha_cluster = true
  user_name     = "my_initial_user"
  password      = "thiZ_is_v&ry_s3cret"

  disable_backup            = false
  backup_schedule_frequency = 24 # every day
  backup_schedule_retention = 7  # keep it one week
}
```

### Example with custom private network

```hcl
# VPC PRIVATE NETWORK
resource "scaleway_vpc_private_network" "pn" {
  name = "my_private_network"
  ipv4_subnet {
    subnet = "172.16.20.0/22"
  }
}

# RDB INSTANCE CONNECTED ON A CUSTOM PRIVATE NETWORK
resource "scaleway_rdb_instance" "main" {
  name              = "test-rdb"
  node_type         = "db-dev-s"
  engine            = "PostgreSQL-11"
  is_ha_cluster     = false
  disable_backup    = true
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  region            = "fr-par"
  tags              = ["terraform-test", "scaleway_rdb_instance", "volume", "rdb_pn"]
  volume_type       = "bssd"
  volume_size_in_gb = 10
  private_network {
    ip_net = "172.16.20.4/22" # IP address within a given IP network
    pn_id  = scaleway_vpc_private_network.pn.id
  }
}


### Configuring Logs config
resource "scaleway_rdb_instance" "instance_with_logs_policy" {
  # other configs ..
  logs_policy {
    max_age_retention    = 30
    total_disk_retention = 100000000
  }
}
```

## Arguments Reference

The following arguments are supported:

- `node_type` - (Required) The type of database instance you want to create (e.g. `db-dev-s`).

~> **Important:** Updates to `node_type` will upgrade the Database Instance to the desired `node_type` without any
interruption. Keep in mind that you cannot downgrade a Database Instance.

- `engine` - (Required) Database Instance's engine version (e.g. `PostgreSQL-11`).

~> **Important:** Updates to `engine` will recreate the Database Instance.

- `volume_type` - (Optional, default to `lssd`) Type of volume where data are stored (`bssd` or `lssd`).

- `volume_size_in_gb` - (Optional) Volume size (in GB) when `volume_type` is set to `bssd`.

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

- `init_settings` - (Optional) Map of engine settings to be set at database initialisation.

~> **Important:** Updates to `init_settings` will recreate the Database Instance.

- `settings` - (Optional) Map of engine settings to be set. Using this option will override default config.

- `logs_policy` (Optional) List of logs policy. More about it in the [logs policy](#logs-policy) section
-
- `tags` - (Optional) The tags associated with the Database Instance.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions)
  in which the Database Instance should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the Database
  Instance is associated with.

## Settings

Please consult
the [GoDoc](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@v1.0.0-beta.9/api/rdb/v1#EngineVersion) to list all
available `settings` and `init_settings` on your `node_type` of your convenient.

## Logs Policy

The `logs_policy` is a list of attributes allowing to configure the logs of a Database Instance:

- `max_age_retention` - `max age (in days) of remote logs to keep on the Database Instance.

- `total_disk_retention` - max disk size of remote logs to keep on the Database Instance. It must be greater than or
  equal to 100000000

More about the logs policy, check
our [documentation](https://www.scaleway.com/en/developers/api/managed-database-postgre-mysql/#path-database-instances-list-available-logs-of-a-database-instance).

## Private Network

~> **Important:** Updates to `private_network` will recreate the attachment Instance.

~> **NOTE:** Please calculate your host IP.
using [cirhost](https://developer.hashicorp.com/terraform/language/functions/cidrhost). Otherwise, lets IPAM service
handle the host IP on the network.

- `ip_net` - (Optional) The IP network address within the private subnet. This must be an IPv4 address with a
  CIDR notation. The IP network address within the private subnet is determined by the IP Address Management (IPAM)
  service if not set.
- `pn_id` - (Required) The ID of the private network.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Database Instance.

~> **Important:** Database instances' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they
are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

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
    - `endpoint_id` - The ID of the endpoint.
    - `ip` - IPv4 address on the network.
    - `port` - Port in the Private Network.
    - `name` - Name of the endpoint.
    - `hostname` - Hostname of the endpoint.
- `certificate` - Certificate of the database instance.
- `organization_id` - The organization ID the Database Instance is associated with.

## Limitations

The Managed Database product is only compliant with the private network in the default availability zone (AZ).
i.e. `fr-par-1`, `nl-ams-1`, `pl-waw-1`. To learn more, read our
section [How to connect a PostgreSQL and MySQL Database Instance to a Private Network](https://www.scaleway.com/en/docs/managed-databases/postgresql-and-mysql/how-to/connect-database-private-network/)

## Import

Database Instance can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_rdb_instance.rdb01 fr-par/11111111-1111-1111-1111-111111111111
```
