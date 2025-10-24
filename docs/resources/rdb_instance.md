---
subcategory: "Databases"
page_title: "Scaleway: scaleway_rdb_instance"
---

# Resource: scaleway_rdb_instance

Creates and manages Scaleway Database Instances.
For more information, see refer to the [API documentation](https://www.scaleway.com/en/developers/api/managed-database-postgre-mysql/).

## Example Usage

### Example Basic

```terraform
resource "scaleway_rdb_instance" "main" {
  name               = "test-rdb"
  node_type          = "DB-DEV-S"
  engine             = "PostgreSQL-15"
  is_ha_cluster      = true
  disable_backup     = true
  user_name          = "my_initial_user"
  password           = "thiZ_is_v&ry_s3cret"
  encryption_at_rest = true
}
```

### Example Block Storage Low Latency

```terraform
resource "scaleway_rdb_instance" "main" {
  name              = "test-rdb-sbs"
  node_type         = "db-play2-pico"
  engine            = "PostgreSQL-15"
  is_ha_cluster     = true
  disable_backup    = true
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_type       = "sbs_15k"
  volume_size_in_gb = 10
}
```

### Example with Settings

```terraform
resource "scaleway_rdb_instance" "main" {
  name           = "test-rdb"
  node_type      = "db-dev-s"
  disable_backup = true
  engine         = "MySQL-8"
  user_name      = "my_initial_user"
  password       = "thiZ_is_v&ry_s3cret"
  init_settings = {
    "lower_case_table_names" = 1
  }
  settings = {
    "max_connections" = "350"
  }
}
```

### Example with backup schedule

```terraform
resource "scaleway_rdb_instance" "main" {
  name          = "test-rdb"
  node_type     = "DB-DEV-S"
  engine        = "PostgreSQL-15"
  is_ha_cluster = true
  user_name     = "my_initial_user"
  password      = "thiZ_is_v&ry_s3cret"

  disable_backup            = false
  backup_schedule_frequency = 24 # every day
  backup_schedule_retention = 7  # keep it one week
}
```

### Example Engine Upgrade

```terraform
# Initial creation with PostgreSQL 14
resource "scaleway_rdb_instance" "main" {
  name           = "my-database"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-14"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_user"
  password       = "thiZ_is_v&ry_s3cret"
}

# Check available versions for upgrade
output "upgradable_versions" {
  value = scaleway_rdb_instance.main.upgradable_versions
}

# To upgrade to PostgreSQL 15, simply change the engine value
# This will trigger a blue/green upgrade with automatic endpoint migration
# resource "scaleway_rdb_instance" "main" {
#   name           = "my-database"
#   node_type      = "DB-DEV-S"
#   engine         = "PostgreSQL-15"  # Changed from PostgreSQL-14
#   is_ha_cluster  = false
#   disable_backup = true
#   user_name      = "my_user"
#   password       = "thiZ_is_v&ry_s3cret"
# }
```

### Examples of endpoint configuration

Database Instances can have a maximum of 1 public endpoint and 1 private endpoint. They can have both, or none.

#### 1 static Private Network endpoint

```terraform
resource "scaleway_vpc_private_network" "pn" {
  ipv4_subnet {
    subnet = "172.16.20.0/22"
  }
}

resource "scaleway_rdb_instance" "main" {
  node_type = "db-dev-s"
  engine    = "PostgreSQL-15"
  private_network {
    pn_id  = scaleway_vpc_private_network.pn.id
    ip_net = "172.16.20.4/22" # IP address within a given IP network
    # enable_ipam = false
  }
}
```

#### 1 IPAM Private Network endpoint + 1 public endpoint

```terraform
resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_rdb_instance" "main" {
  node_type = "DB-DEV-S"
  engine    = "PostgreSQL-15"
  private_network {
    pn_id       = scaleway_vpc_private_network.pn.id
    enable_ipam = true
  }
  load_balancer {}
}
```

#### Default: 1 public endpoint

```terraform
resource "scaleway_rdb_instance" "main" {
  node_type = "db-dev-s"
  engine    = "PostgreSQL-15"
}
```

-> **Note** If nothing is defined, your Database Instance will have a default public load-balancer endpoint.

-> **Note** Managed PostgreSQL and MySQL Database Instances are compatible with the [VPC routing](https://www.scaleway.com/en/docs/network/vpc/concepts/#routing) feature, which allows you to connect one or more Database Instances in a Private Network to resources in other Private Networks of the same VPC. This feature is automatically enabled when your Database Instance is connected to a Private Network within a VPC that has routing enabled. Refer to the [How to manage routing](https://www.scaleway.com/en/docs/network/vpc/how-to/manage-routing/) documentation page for more information about VPC routing.

## Argument Reference

The following arguments are supported:

- `node_type` - (Required) The type of Database Instance you want to create (e.g. `db-dev-s`).

~> **Important** Updates to `node_type` will upgrade the Database Instance to the desired `node_type` without any
interruption.

~> **Important** Once your Database Instance reaches `disk_full` status, if you are using `lssd` storage, you should upgrade the `node_type`, and if you are using `bssd` storage, you should increase the volume size before making any other changes to your Database Instance.

- `engine` - (Required) Database Instance's engine version name (e.g. `PostgreSQL-16`, `MySQL-8`).

~> **Important** Updates to `engine` will perform a blue/green upgrade using `MajorUpgradeWorkflow`. This creates a new instance from a snapshot, migrates endpoints automatically, and updates the Terraform state with the new instance ID. The upgrade ensures minimal downtime but **any writes between the snapshot and the endpoint migration will be lost**. Use the `upgradable_versions` computed attribute to check available versions for upgrade.

- `volume_type` - (Optional, default to `lssd`) Type of volume where data are stored (`lssd`, `sbs_5k` or `sbs_15k`).

- `volume_size_in_gb` - (Optional) Volume size (in GB). Cannot be used when `volume_type` is set to `lssd`.

~> **Important** Once your Database Instance reaches `disk_full` status, you should increase the volume size before making any other change to your Database Instance.

- `user_name` - (Optional) Identifier for the first user of the Database Instance.

~> **Important** Updates to `user_name` will recreate the Database Instance.

- `password` - (Optional) Password for the first user of the Database Instance.

- `is_ha_cluster` - (Optional) Enable or disable high availability for the Database Instance.

~> **Important** Updates to `is_ha_cluster` will recreate the Database Instance.

- `name` - (Optional) The name of the Database Instance.

- `tags` - (Optional) The tags associated with the Database Instance.

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions)
  in which the Database Instance should be created.

- `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the project the Database
  Instance is associated with.

- `encryption_at_rest` - (Optional) Enable or disable encryption at rest for the Database Instance.

- `snapshot_id` - (Optional) The ID of an existing snapshot to restore or create the Database Instance from. Conflicts with the `engine` parameter and backup settings.

### Backups

- `disable_backup` - (Optional) Disable automated backup for the Database Instance.

- `backup_schedule_frequency` - (Optional) Backup schedule frequency in hours.

- `backup_schedule_retention` - (Optional) Backup schedule retention in days.

- `backup_same_region` - (Optional) Boolean to store logical backups in the same region as the Database Instance.

### Settings

- `settings` - (Optional) Map of engine settings to be set. Using this option will override default config.

- `init_settings` - (Optional) Map of engine settings to be set at database initialisation.

~> **Important** Updates to `init_settings` will recreate the Database Instance.

-> **Note** Refer to the [GoDoc](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@v1.0.0-beta.9/api/rdb/v1#EngineVersion) to list all available `settings` and `init_settings` for the `node_type` of your convenience.

### Endpoints

- `private_network` - List of Private Networks endpoints of the Database Instance.

    - `pn_id` - (Required) The ID of the Private Network.
    - `ip_net` - (Optional) The IP network address within the private subnet. This must be an IPv4 address with a CIDR notation. If not set, The IP network address within the private subnet is determined by the IP Address Management (IPAM) service.
    - `enable_ipam` - (Optional) If true, the IP network address within the private subnet is determined by the IP Address Management (IPAM) service.

~> **Important** One of `ip_net` or `enable_ipam=true` must be set.

~> **Important** Updates to `private_network` will recreate the Instance's endpoint

~> **Note** You can calculate your host IP using [cidrhost](https://developer.hashicorp.com/terraform/language/functions/cidrhost). Otherwise, let IPAM service
handle the host IP on the network.

- `load_balancer` - (Optional) List of Load Balancer endpoints of the Database Instance. A load-balancer endpoint will be set by default if no Private Network is.
  This block must be defined if you want a public endpoint in addition to your private endpoint.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Database Instance.

~> **Important** Database Instances' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they
are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `endpoint_ip` - (Deprecated) The IP of the Database Instance. Please use the private_network or the load_balancer attribute.
- `endpoint_port` - (Deprecated) The port of the Database Instance. Please use the private_network or the load_balancer attribute.
- `read_replicas` - List of read replicas of the Database Instance.
    - `ip` - IP of the replica.
    - `port` - Port of the replica.
    - `name` - Name of the replica.
- `load_balancer` - List of Load Balancer endpoints of the Database Instance.
    - `endpoint_id` - The ID of the endpoint of the Load Balancer.
    - `ip` - IP of the endpoint.
    - `port` - Port of the endpoint.
    - `name` - Name of the endpoint.
    - `hostname` - Name of the endpoint.
- `private_network` - List of Private Networks endpoints of the Database Instance.
    - `endpoint_id` - The ID of the endpoint.
    - `ip` - IPv4 address on the network.
    - `port` - Port in the Private Network.
    - `name` - Name of the endpoint.
    - `hostname` - Hostname of the endpoint.
- `private_ip` - The private IPv4 address associated with the resource.
    - `id` - The ID of the IPv4 address resource.
    - `address` - The private IPv4 address.
- `certificate` - Certificate of the Database Instance.
- `organization_id` - The organization ID the Database Instance is associated with.
- `upgradable_versions` - List of available engine versions for upgrade. Each version contains:
    - `id` - Version ID to use in upgrade requests.
    - `name` - Engine version name (e.g., `PostgreSQL-15`).
    - `version` - Version string (e.g., `15.5`).
    - `minor_version` - Minor version string (e.g., `15.5.0`).

## Limitations

The Managed Database product is only compliant with the Private Network in the default availability zone (AZ).
i.e. `fr-par-1`, `nl-ams-1`, `pl-waw-1`. To learn more, read our
section [How to connect a PostgreSQL and MySQL Database Instance to a Private Network](https://www.scaleway.com/en/docs/managed-databases/postgresql-and-mysql/how-to/connect-database-private-network/)

## Import

Database Instance can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_rdb_instance.rdb01 fr-par/11111111-1111-1111-1111-111111111111
```
