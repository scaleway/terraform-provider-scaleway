---
page_title: "Scaleway: scaleway_instance_private_nic"
description: |-
  Manages Scaleway Compute Instance Private NICs.
---

# scaleway_instance_private_nic

Creates and manages Scaleway Instance Private NICs. For more information, see
[the documentation](https://developers.scaleway.com/en/products/instance/api/#private-nics-a42eea).

## Examples

### Basic

```hcl
resource "scaleway_instance_private_nic" "pnic01" {
    server_id          = "fr-par-1/11111111-1111-1111-1111-111111111111"
    private_network_id = "fr-par-1/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}
```

### With zone

```hcl
resource scaleway_vpc_private_network pn01 {
  name = "private_network_instance"
  zone = "fr-par-2"
}

resource "scaleway_instance_server" "base" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"
  zone = scaleway_vpc_private_network.pn01.zone
}

resource "scaleway_instance_private_nic" "pnic01" {
  server_id = scaleway_instance_server.base.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  zone = scaleway_vpc_private_network.pn01.zone
}
```

## Arguments Reference

The following arguments are required:

- `server_id` - (Required) The ID of the server associated with.
- `private_network_id` - (Required) The ID of the private network attached to.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the server must be created.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the private NIC.
- `mac_address` - The MAC address of the private NIC.

## Import

Private NICs can be imported using the `{zone}/{server_id}/{private_nic_id}`, e.g.

```bash
$ terraform import scaleway_instance_private_nic.pnic01 fr-par-1/11111111-1111-1111-1111-111111111111/22222222-2222-2222-2222-222222222222
```
