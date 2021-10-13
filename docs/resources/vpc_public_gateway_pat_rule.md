---
page_title: "Scaleway: scaleway_vpc_public_gateway_pat_rule"
description: |-
Manages Scaleway VPC Public Gateways PAT rules.
---

# scaleway_vpc_public_gateway_pat_rule

Creates and manages Scaleway VPC Public Gateway PAT (Port Address Translation).
For more information, see [the documentation](https://developers.scaleway.com/en/products/vpc-gw/api/#pat-rules-e75d10).

## Example

```hcl
resource scaleway_vpc_public_gateway pg01 {
  type = "VPC-GW-S"
}

resource scaleway_vpc_public_gateway_dhcp dhcp01 {
  subnet = "192.168.1.0/24"
}

resource scaleway_vpc_private_network pn01 {
  name = "pn_test_network"
}

resource scaleway_vpc_gateway_network gn01 {
  gateway_id = scaleway_vpc_public_gateway.pg01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  dhcp_id = scaleway_vpc_public_gateway_dhcp.dhcp01.id
  cleanup_dhcp = true
}

resource scaleway_vpc_public_gateway_pat_rule main {
  gateway_id = scaleway_vpc_public_gateway.pg01.id
  private_ip = scaleway_vpc_public_gateway_dhcp.dhcp01.address
  private_port = 42
  public_port = 42
  protocol = "both"
  depends_on = [scaleway_vpc_gateway_network.gn01, scaleway_vpc_private_network.pn01]
}
```

## Arguments Reference

The following arguments are supported:

- `gateway_id` - (Required) The ID of the public gateway.
- `private_ip` - (Required) The Private IP to forward data to (IP address).
- `public_port` - (Required) The Public port to listen on.
- `private_port` - (Required) The Private port to translate to.
- `protocol` - (Defaults to both) The Protocol the rule should apply to. Possible values are both, tcp and udp.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the public gateway DHCP config should be created.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the public gateway DHCP config.
- `organization_id` - The organization ID the pat rule config is associated with.
- `created_at` - The date and time of the creation of the pat rule config.
- `updated_at` - The date and time of the last update of the pat rule config.

## Import

Public gateway PAT rules config can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_vpc_public_gateway_pat_rule.main fr-par-1/11111111-1111-1111-1111-111111111111
```
