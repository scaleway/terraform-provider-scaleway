---
subcategory: "MongoDB®"
page_title: "Scaleway: scaleway_mongodb_instance"
---

# Resource: scaleway_mongodb_instance

Creates and manages Scaleway MongoDB® instance.
For more information refer to the [product documentation](https://www.scaleway.com/en/docs/managed-mongodb-databases/).



## Example Usage

```terraform
### Basic MongoDB instance creation

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-basic1"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5
}
```

```terraform
### Create and instance with a Write Only password (not stored in state), update and rollback the password while ensuring the password is not stored in the state

# Generate an ephemeral password (not stored in the state)
ephemeral "random_password" "main" {
  length      = 20
  special     = true
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  min_special = 1
  # Exclude characters that might cause issues in some contexts
  override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
}

resource "scaleway_secret" "main" {
  name        = "mongodb-instance-password"
  description = "Password for MongoDB instance"
}

# Store the generated password in a Write Only data (not stored in the state)
resource "scaleway_secret_version" "main" {
  secret_id       = scaleway_secret.main.id
  data_wo         = ephemeral.random_password.main.result
  data_wo_version = 1
}

# Create an instance, using the ephemeral password in the Write Only password attribute (not stored in the state)
resource "scaleway_mongodb_instance" "password_wo_instance" {
  name                = "test-mongodb-password-wo-rollback"
  version             = "7.0.12"
  node_type           = "MGDB-PLAY2-NANO"
  node_number         = 1
  user_name           = "my_initial_user"
  password_wo         = ephemeral.random_password.main.result
  password_wo_version = scaleway_secret_version.main.data_wo_version
}

## Generate a new ephemeral password (not stored in the state)
ephemeral "random_password" "renewed" {
  length      = 20
  special     = true
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  min_special = 1
  # Exclude characters that might cause issues in some contexts
  override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
}

# Store the renewed generated password in a Write Only data (not stored in the state)
resource "scaleway_secret_version" "renewed" {
  secret_id       = scaleway_secret.main.id
  data_wo         = ephemeral.random_password.renewed.result
  data_wo_version = 2
}

# Renew the instance password
# resource "scaleway_mongodb_instance" "password_wo_instance" {
#   name                = "test-mongodb-password-wo-rollback"
#   version             = "7.0.12"
#   node_type           = "MGDB-PLAY2-NANO"
#   node_number         = 1
#   user_name           = "my_initial_user"
#   password_wo         = ephemeral.random_password.renewed.result
#   password_wo_version = scaleway_secret_version.renewed.data_wo_version
# }

# Query the first password version as an Ephemeral Resource (not stored in the state)
# ephemeral "scaleway_secret_version" "main" {
#   secret_id = scaleway_secret.main.id
#   revision  = 1
# }

# resource "scaleway_mongodb_instance" "password_wo_instance" {
#   name                = "test-mongodb-password-wo-rollback"
#   version             = "7.0.12"
#   node_type           = "MGDB-PLAY2-NANO"
#   node_number         = 1
#   user_name           = "my_initial_user"
#   password_wo         = ephemeral.scaleway_secret_version.main.data
#   password_wo_version = 1
# }
```

```terraform
### Creating a MongoDB instance using a Write Only password (not stored in state)

## Generate an ephemeral password (not stored in the state)
ephemeral "random_password" "db_password" {
  length      = 20
  special     = true
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  min_special = 1
  # Exclude characters that might cause issues in some contexts
  override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
}

resource "scaleway_mongodb_instance" "password_wo_instance" {
  name                = "test-mongodb-password-wo"
  version             = "7.0.12"
  node_type           = "MGDB-PLAY2-NANO"
  node_number         = 1
  user_name           = "my_initial_user"
  password_wo         = ephemeral.random_password.db_password.result
  password_wo_version = 1
}
```

```terraform
### MongoDB instance with Private Network

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my_private_network"
  region = "fr-par"
}

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-basic1"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5

  private_network {
    pn_id = scaleway_vpc_private_network.pn01.id
  }
}
```

```terraform
### MongoDB instance with Private Network and Public Network

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my_private_network"
  region = "fr-par"
}

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-basic1"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5

  private_network {
    pn_id = scaleway_vpc_private_network.pn01.id
  }

  public_network {}
}
```

```terraform
### MongoDB instance restored from Snapshot

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-basic1"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5
}

resource "scaleway_mongodb_snapshot" "main_snapshot" {
  instance_id = scaleway_mongodb_instance.main.id
  name        = "my-mongodb-snapshot"
}

resource "scaleway_mongodb_instance" "restored_instance" {
  snapshot_id = scaleway_mongodb_snapshot.main_snapshot.id
  name        = "restored-mongodb-from-snapshot"
  node_type   = "MGDB-PLAY2-NANO"
  node_number = 1
}
```

```terraform
### MongoDB instance with Snapshot Scheduling

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-with-snapshots"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5

  # Snapshot scheduling configuration
  snapshot_schedule_frequency_hours = 24
  snapshot_schedule_retention_days  = 7
  is_snapshot_schedule_enabled      = true
}
```





## Argument Reference

The following arguments are supported:

- `version` - (Optional) MongoDB® version of the instance.
- `node_type` - (Required) The type of MongoDB® instance to create.
- `user_name` - (Optional) Name of the user created when the instance is created.
- `password` - (Optional) Password of the user.
- `name` - (Optional) Name of the MongoDB® instance.
- `tags` - (Optional) List of tags attached to the MongoDB® instance.
- `volume_type` - (Optional) Volume type of the instance.
- `volume_size_in_gb` - (Optional) Volume size in GB.
- `snapshot_id` - (Optional) Snapshot ID to restore the MongoDB® instance from.
- `private_network` - (Optional) Private Network endpoints of the Database Instance.
    - `pn_id` - (Required) The ID of the Private Network.
- `public_network` - (Optional) Public network endpoint configuration (no arguments).
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the MongoDB® instance should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the MongoDB® instance is associated with.

### Snapshot Scheduling

- `snapshot_schedule_frequency_hours` - (Optional) Snapshot schedule frequency in hours.
- `snapshot_schedule_retention_days` - (Optional) Snapshot schedule retention in days.
- `is_snapshot_schedule_enabled` - (Optional) Enable or disable automatic snapshot scheduling.

~> **Important** If neither private_network nor public_network is specified, a public network endpoint is created by default.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the MongoDB® instance.
- `created_at` - The date and time of the creation of the MongoDB® instance.
- `updated_at` - The date and time of the last update of the MongoDB® instance.
- `region` - The region of the MongoDB® instance.
- `private_network` - Private Network endpoints of the Database Instance.
    - `id` - The ID of the endpoint.
    - `ips` - List of IP addresses for your endpoint.
    - `port` - TCP port of the endpoint.
    - `dns_records` - List of DNS records for your endpoint.
- `private_ip` - The private IPv4 address associated with the instance.
    - `id` - The ID of the IPv4 address resource.
    - `address` - The private IPv4 address.
- `public_network` - Private Network endpoints of the Database Instance.
    - `id` - The ID of the endpoint.
    - `port` - TCP port of the endpoint.
    - `dns_records` - List of DNS records for your endpoint.
- `snapshot_schedule_frequency_hours` - Snapshot schedule frequency in hours.
- `snapshot_schedule_retention_days` - Snapshot schedule retention in days.
- `is_snapshot_schedule_enabled` - Whether automatic snapshot scheduling is enabled.
- `tls_certificate` - The PEM-encoded TLS certificate for the MongoDB® instance, if available.

## Import

MongoDB® instance can be imported using the `id`, e.g.

```bash
terraform import scaleway_mongodb_instance.main fr-par/11111111-1111-1111-1111-111111111111
```
