---
page_title: "Scaleway: scaleway_vpc_public_gateway_pat_rule"
description: |- Get information about Scaleway VPC Public Gateway PAT rule.
---

# scaleway_vpc_public_gateway_pat_rule

Gets information about a public gateway PAT rule. For further information please check the
API [documentation](https://developers.scaleway.com/en/products/vpc-gw/api/v1/#get-8faeea)

## Example Usage

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
	depends_on = [scaleway_vpc_private_network.pn01]
	cleanup_dhcp = true
	enable_masquerade = true
}

resource scaleway_vpc_public_gateway_pat_rule main {
	gateway_id = scaleway_vpc_public_gateway.pg01.id
	private_ip = scaleway_vpc_public_gateway_dhcp.dhcp01.address
	private_port = 42
	public_port = 42
	protocol = "both"
	depends_on = [scaleway_vpc_gateway_network.gn01, scaleway_vpc_private_network.pn01]
}

data "scaleway_vpc_public_gateway_pat_rule" "main" {
	pat_rule_id = "${scaleway_vpc_public_gateway_pat_rule.main.id}"
}
```

## Argument Reference

- `pat_rule_id`  (Required) The ID of the PAT rule to retrieve
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which
  the image exists.

## Attributes Reference

`id` is set to the ID of the found public gateway PAT RULE.

The following arguments are exported:

- `gateway_id` - The ID of the public gateway.
- `private_ip` - The Private IP to forward data to (IP address).
- `public_port` - The Public port to listen on.
- `private_port` - The Private port to translate to.
- `protocol` - The Protocol the rule should apply to. Possible values are both, tcp and udp.
