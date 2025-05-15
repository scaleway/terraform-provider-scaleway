---
subcategory: "MongoDB®"
page_title: "Scaleway: scaleway_mongodb_instance"
---

# Resource: scaleway_mongodb_instance

Creates and manages Scaleway MongoDB® instance.
For more information refer to the [product documentation](https://www.scaleway.com/en/docs/managed-mongodb-databases/).

## Example Usage

### Basic

```terraform
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

### Private Network

```terraform
resource scaleway_vpc_private_network pn01 {
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
    pn_id = "${scaleway_vpc_private_network.pn02.id}"
  }

}
```


### Restore From Snapshot

```terraform

resource "scaleway_mongodb_instance" "restored_instance" {
  snapshot_id = "${scaleway_vpc_private_network.pn.idscaleway_mongodb_snapshot.main_snapshot.id}"
  name        = "restored-mongodb-from-snapshot"
  node_type   = "MGDB-PLAY2-NANO"
  node_number = 1
}
```

## Argument Reference

The following arguments are supported:

- `version` - (Optional) MongoDB® version of the instance.
- `node_type` - (Required) The type of MongoDB® intance to create.
- `user_name` - (Optional) Name of the user created when the intance is created.
- `password` - (Optional) Password of the user.
- `name` - (Optional) Name of the MongoDB® instance.
- `tags` - (Optional) List of tags attached to the MongoDB® instance.
- `volume_type` - (Optional) Volume type of the instance.
- `volume_size_in_gb` - (Optional) Volume size in GB.
- `snapshot_id` - (Optional) Snapshot ID to restore the MongoDB® instance from.
- `private_network` - (Optional) Private Network endpoints of the Database Instance.
    - `pn_id` - (Required) The ID of the Private Network.
- `public_network` - (Optional) Public network specs details.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the MongoDB® instance.
- `created_at` - The date and time of the creation of the MongoDB® instance.
- `updated_at` - The date and time of the last update of the MongoDB® instance.
- `private_network` - Private Network endpoints of the Database Instance.
    - `id` - The ID of the endpoint.
    - `ips` - List of IP addresses for your endpoint.
    - `port` - TCP port of the endpoint.
    - `dns_records` - List of DNS records for your endpoint.

## Import

MongoDB® instance can be imported using the `id`, e.g.

```bash
terraform import scaleway_mongodb_instance.main fr-par-1/11111111-1111-1111-1111-111111111111
```
