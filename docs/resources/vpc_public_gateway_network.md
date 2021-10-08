---
page_title: "Scaleway: scaleway_vpc_public_gateway_network"
description: |-
  Manages Scaleway VPC Public Gateway Networks.
---

# scaleway_vpc_public_gateway_network

Creates and manages Scaleway VPC Public Gateway Network.
It allows attaching Private Networks to the VPC Public Gateway and your DHCP config
For more information, see [the documentation](https://developers.scaleway.com/en/products/vpc-gw/api/#step-3-attach-private-networks-to-the-vpc-public-gateway).

## Example

```hcl
resource scaleway_vpc_private_network pn01 {
  name = "pn_test_network"
}

resource scaleway_vpc_public_gateway_ip gw01 {
}

resource scaleway_vpc_public_gateway_dhcp dhcp01 {
  subnet = "192.168.1.0/24"
}

resource scaleway_vpc_public_gateway pg01 {
  name = "foobar"
  type = "VPC-GW-S"
  ip_id = scaleway_vpc_public_gateway_ip.gw01.id
}

resource scaleway_vpc_public_gateway_network main {
  gateway_id = scaleway_vpc_public_gateway.pg01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
}
```

## Arguments Reference

The following arguments are supported:

- `gateway_id` - (Required) The ID of the public gateway.
- `private_network_id` - (Required) The ID of the private network.
- `dhcp_id` - (Required) The ID of the public gateway DHCP config.
- `enable_masquerade` - (Defaults to false) Enable masquerade on this network
- `enable_dhcp` - (Defaults to true) Enable DHCP config on this network. It requires DHCP id.
- `static_address` - Enable DHCP config on this network
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the public gateway should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the public gateway is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the public gateway network.
- `mac_address` - The mac address of the creation of the gateway network.
- `created_at` - The date and time of the creation of the gateway network.
- `updated_at` - The date and time of the last update of the gateway network.

## Import

Public gateway network can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_vpc_public_gateway_network.main fr-par-1/11111111-1111-1111-1111-111111111111
```

