---
page_title: "Scaleway: scaleway_instance_server"
description: |-
  Manages Scaleway Compute Instance servers.
---

# scaleway_instance_server

Creates and manages Scaleway Compute Instance servers. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#servers-8bf7d7).

## Examples

### Basic

```hcl
resource "scaleway_instance_ip" "public_ip" {}

resource "scaleway_instance_server" "web" {
  type = "DEV1-S"
  image = "ubuntu_focal"
  ip_id = scaleway_instance_ip.public_ip.id
}
```

### With additional volumes and tags

```hcl
resource "scaleway_instance_volume" "data" {
  size_in_gb = 100
  type = "b_ssd"
}

resource "scaleway_instance_server" "web" {
  type = "DEV1-S"
  image = "ubuntu_focal"

  tags = [ "hello", "public" ]

  root_volume {
    delete_on_termination = false
  }

  additional_volume_ids = [ scaleway_instance_volume.data.id ]
}
```

### With a reserved IP

```hcl
resource "scaleway_instance_ip" "ip" {}

resource "scaleway_instance_server" "web" {
  type = "DEV1-S"
  image = "f974feac-abae-4365-b988-8ec7d1cec10d"

  tags = [ "hello", "public" ]

  ip_id = scaleway_instance_ip.ip.id
}
```

### With security group

```hcl
resource "scaleway_instance_security_group" "www" {
  inbound_default_policy = "drop"
  outbound_default_policy = "accept"

  inbound_rule {
    action = "accept"
    port = "22"
    ip = "212.47.225.64"
  }

  inbound_rule {
    action = "accept"
    port = "80"
  }

  inbound_rule {
    action = "accept"
    port = "443"
  }

  outbound_rule {
    action = "drop"
    ip_range = "10.20.0.0/24"
  }
}

resource "scaleway_instance_server" "web" {
  type = "DEV1-S"
  image = "ubuntu_focal"

  security_group_id= scaleway_instance_security_group.www.id
}
```

### With user data and cloud-init

```hcl
resource "scaleway_instance_server" "web" {
  type = "DEV1-S"
  image = "ubuntu_focal"

  tags = [ "web", "public" ]

  user_data {
    key = "plop"
    value = "world"
  }

  user_data {
    key = "xavier"
    value = "niel"
  }

  cloud_init = file("${path.module}/cloud-init.yml")
}
```

## Arguments Reference

The following arguments are supported:

- `type` - (Required) The commercial type of the server.
You find all the available types on the [pricing page](https://www.scaleway.com/en/pricing/).
Updates to this field will recreate a new resource.

[//]: # (TODO: Improve me)

- `image` - (Required) The UUID or the label of the base image used by the server. You can use [this endpoint](https://api-marketplace.scaleway.com/images?page=1&per_page=100)
to find either the right `label` or the right local image `ID` for a given `type`.

[//]: # (TODO: Improve me)

- `name` - (Optional) The name of the server.

- `tags` - (Optional) The tags associated with the server.

- `security_group_id` - (Optional) The [security group](https://developers.scaleway.com/en/products/instance/api/#security-groups-8d7f89) the server is attached to.

- `placement_group_id` - (Optional) The [placement group](https://developers.scaleway.com/en/products/instance/api/#placement-groups-d8f653) the server is attached to.


~> **Important:** When updating `placement_group_id` the `state` must be set to `stopped`, otherwise it will fail.

- `root_volume` - (Optional) Root [volume](https://developers.scaleway.com/en/products/instance/api/#volumes-7e8a39) attached to the server on creation.
    - `size_in_gb` - (Required) Size of the root volume in gigabytes.
    To find the right size use [this endpoint](https://api.scaleway.com/instance/v1/zones/fr-par-1/products/servers) and
    check the `volumes_constraint.{min|max}_size` (in bytes) for your `commercial_type`.
    Updates to this field will recreate a new resource.
    - `delete_on_termination` - (Defaults to `true`) Forces deletion of the root volume on instance termination.

~> **Important:** Updates to `root_volume.size_in_gb` will be ignored after the creation of the server.

- `additional_volume_ids` - (Optional) The [additional volumes](https://developers.scaleway.com/en/products/instance/api/#volumes-7e8a39)
attached to the server. Updates to this field will trigger a stop/start of the server.

~> **Important:** If this field contains local volumes, the `state` must be set to `stopped`, otherwise it will fail.

~> **Important:** If this field contains local volumes, you have to first detach them, in one apply, and then delete the volume in another apply.

- `enable_ipv6` - (Defaults to `false`) Determines if IPv6 is enabled for the server.

- `ip_id` = (Optional) The ID of the reserved IP that is attached to the server.

- `enable_dynamic_ip` - (Defaults to `false`) If true a dynamic IP will be attached to the server.

- `state` - (Defaults to `started`) The state of the server. Possible values are: `started`, `stopped` or `standby`.

- `cloud_init` - (Optional) The cloud init script associated with this server. Updates to this field will trigger a stop/start of the server.

- `user_data` - (Optional) The user data associated with the server.

    - `key` - (Required) The user data key. The `cloud-init` key is reserved, please use `cloud_init` attribute instead.

    - `value` - (Required) The user data content. It could be a string or a file content using [file](https://www.terraform.io/docs/configuration/functions/file.html) or [filebase64](https://www.terraform.io/docs/configuration/functions/filebase64.html) for example.

- `boot_type` - The boot Type of the server. Possible values are: `local`, `bootscript` or `rescue`.

- `bootscript_id` - The ID of the bootscript to use  (set boot_type to `bootscript`).

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the server should be created.

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the server is associated with.
  If you intend to deploy your instance in another project than the default one use your `project_id` instead of the organization id.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the server is associated with.


## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the server.
- `placement_group_policy_respected` - True when the placement group policy is respected.
- `root_volume`
    - `volume_id` - The volume ID of the root volume of the server.
- `private_ip` - The Scaleway internal IP address of the server.
- `public_ip` - The public IPv4 address of the server.
- `ipv6_address` - The default ipv6 address routed to the server. ( Only set when enable_ipv6 is set to true )
- `ipv6_gateway` - The ipv6 gateway address. ( Only set when enable_ipv6 is set to true )
- `ipv6_prefix_length` - The prefix length of the ipv6 subnet routed to the server. ( Only set when enable_ipv6 is set to true )
- `boot_type` - The boot Type of the server. Possible values are: `local`, `bootscript` or `rescue`.

## Import

Instance servers can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_instance_server.web fr-par-1/11111111-1111-1111-1111-111111111111
```
