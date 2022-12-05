---
page_title: "Scaleway: scaleway_baremetal_private_network"
description: |-
Manages Scaleway Compute Baremetal private networks.
---

# scaleway_baremetal_private_network

Creates and manages Scaleway Compute Baremetal private networks. For more information, see [the documentation](https://developers.scaleway.com/en/products/vpc-elasticmetal/api/).

## Examples

### Basic

```hcl
data "scaleway_account_ssh_key" "main" {
  name = "main"
}

data "scaleway_baremetal_os" "my_os" {
  zone    = "fr-par-2"
  name    = "Ubuntu"
  version = "22.04 LTS (Jammy Jellyfish)"
}

data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-2"
  name = "EM-B112X-SSD"
}

data "scaleway_baremetal_option" "private_network" {
  zone = "fr-par-2"
  name = "Private Network"
}

resource "scaleway_baremetal_server" "base" {
  zone        = "fr-par-2"
  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
  os          = data.scaleway_baremetal_os.my_os.os_id
  ssh_key_ids = [data.scaleway_account_ssh_key.main.id]

  options {
    id = data.scaleway_baremetal_option.private_network.option_id
  }
}

resource "scaleway_vpc_private_network" "pn" {
  zone = "fr-par-2"
  name = "private_network_baremetal"
  tags = ["pn", "baremetal"]
}

resource "scaleway_baremetal_private_network" "main" {
  zone = "fr-par-2"
  server_id = scaleway_baremetal_server.base.id
  
  private_networks {
    id = scaleway_vpc_private_network.pn.id
  }
}
```

## Arguments Reference

The following arguments are required:

- `server_id` - (Required) The ID of the server associated with.
- `private_networks` - (Required) The private networks to attach to the server.
  ~> The `private_networks` block supports:
    - `id` - (Required) The id of the private network to attach.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the private network must be created.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the server private network.
- `server_id` - The ID of the server.
- `private_networks` - The private networks attached to the server.
    -> The `private_networks` block contains :
    - `id` - The ID of the private network.
    - `project_id` - The private network project ID.
    - `vlan` - The VLAN ID associated to the private network.
    - `status` - The private network status.
    - `created_at` - The date and time of the creation of the private network.
    - `updated_at` - The date and time of the last update of the private network.

## Import

Baremetal private networks can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_baremetal_private_network.pn01 fr-par-2/11111111-1111-1111-1111-111111111111
```
