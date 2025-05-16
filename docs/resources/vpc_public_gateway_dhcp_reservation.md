---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_public_gateway_dhcp_reservation"
---

# Resource: scaleway_vpc_public_gateway_dhcp_reservation

~> **Important:**  The resource `scaleway_vpc_public_gateway_dhcp_reservation` has been deprecated and will no longer be supported.
In 2023, DHCP functionality was moved from Public Gateways to Private Networks, DHCP resources are now no longer needed.
You can use IPAM to manage your IPs. For more information, please refer to the [dedicated guide](../guides/migration_guide_vpcgw_v2.md).

Creates and manages [Scaleway DHCP Reservations](https://www.scaleway.com/en/docs/vpc/concepts/#dhcp).

These static associations are used to assign IP addresses based on the MAC addresses of the resource.

Statically assigned IP addresses should fall within the configured subnet, but be outside of the dynamic range.

For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/public-gateway/#dhcp-c05544).

[DHCP reservations](https://www.scaleway.com/en/developers/api/public-gateway/#dhcp-entries-e40fb6) hold both dynamic DHCP leases (IP addresses dynamically assigned by the gateway to resources) and static user-created DHCP reservations.

## Example Usage

```terraform
resource scaleway_vpc_private_network main {
  name = "your_private_network"
}

resource "scaleway_instance_server" "main" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"
  zone  = "fr-par-1"

  private_network {
    pn_id = scaleway_vpc_private_network.main.id
  }
}

resource scaleway_vpc_public_gateway_ip main {
}

resource scaleway_vpc_public_gateway_dhcp main {
  subnet = "192.168.1.0/24"
}

resource scaleway_vpc_public_gateway main {
  name  = "foobar"
  type  = "VPC-GW-S"
  ip_id = scaleway_vpc_public_gateway_ip.main.id
}

resource scaleway_vpc_gateway_network main {
  gateway_id         = scaleway_vpc_public_gateway.main.id
  private_network_id = scaleway_vpc_private_network.main.id
  dhcp_id            = scaleway_vpc_public_gateway_dhcp.main.id
  cleanup_dhcp       = true
  enable_masquerade  = true
  depends_on         = [scaleway_vpc_public_gateway_ip.main, scaleway_vpc_private_network.main]
}

resource scaleway_vpc_public_gateway_dhcp_reservation main {
  gateway_network_id = scaleway_vpc_gateway_network.main.id
  mac_address        = scaleway_instance_server.main.private_network.0.mac_address
  ip_address         = "192.168.1.1"
}
```

## Argument Reference

The following arguments are supported:

- `gateway_network_id` - (Required) The ID of the owning GatewayNetwork.
- `ip_address` - (Required) The IP address to give to the machine.
- `mac_address` - (Required) The MAC address for the static entry.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the public gateway DHCP config should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Public Gateway DHCP reservation configuration.

~> **Important:** Public Gateway DHCP reservations configurations IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `hostname` - The hostname of the client machine.
- `type` - The reservation type, either static (DHCP reservation) or dynamic (DHCP lease). Possible values are `reservation` and `lease`.
- `created_at` - The date and time of the creation of the Public Gateway DHCP configuration.
- `updated_at` - The date and time of the last update of the Public Gateway DHCP configuration.

## Import

Public Gateway DHCP reservation configurations can be imported using `{zone}/{id}`, e.g.

```bash
terraform import scaleway_vpc_public_gateway_dhcp_reservation.main fr-par-1/11111111-1111-1111-1111-111111111111
```
