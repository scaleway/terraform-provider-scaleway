---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_gateway_network"
---

# scaleway_vpc_gateway_network

Gets information about a GatewayNetwork (a connection between a Public Gateway and a Private Network)/

## Example Usage

```hcl
resource "scaleway_vpc_gateway_network" "main" {
  gateway_id = scaleway_vpc_public_gateway.pg01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
  cleanup_dhcp       = true
  enable_masquerade  = true
}

data scaleway_vpc_gateway_network by_id {
    gateway_network_id = scaleway_vpc_gateway_network.main.id
}

data scaleway_vpc_gateway_network by_gateway_and_pn {
    gateway_id = scaleway_vpc_public_gateway.pg01.id
    private_network_id = scaleway_vpc_private_network.pn01.id
}
```

## Argument Reference

* `gateway_network_id` - (Optional) ID of the GatewayNetwork.

~> Only one of `gateway_network_id` or usage of the following filters should be specified. If using filters, you can use as many as you want.

* `gateway_id` - (Optional) ID of the Public Gateway the GatewayNetwork is linked to
* `private_network_id` - (Optional) ID of the Private Network the GatewayNetwork is linked to
* `enable_masquerade` - (Optional) Whether masquerade (dynamic NAT) is enabled on requested GatewayNetwork
* `dhcp_id` - (Optional) ID of the Public Gateway's DHCP configuration

## Attributes Reference

Exported attributes are the attributes of the [resource](../resources/vpc_gateway_network.md)
