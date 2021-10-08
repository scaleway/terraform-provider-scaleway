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
resource scaleway_vpc_public_gateway main {
  type = "VPC-GW-S"
}

resource scaleway_vpc_public_gateway_pat_rule main {
  gateway_id = scaleway_vpc_public_gateway.main.id
  private_ip = "192.168.0.1"
  private_port = 8080
  public_port = 8080
  protocol = "both"
}
```

## Arguments Reference

The following arguments are supported:

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the public gateway DHCP config should be created.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the public gateway DHCP config.
- `organization_id` - The organization ID the public gateway DHCP config is associated with.
- `created_at` - The date and time of the creation of the public gateway DHCP config.
- `updated_at` - The date and time of the last update of the public gateway DHCP config.

## Import

Public gateway PAT rules config can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_vpc_public_gateway_pat_rule.main fr-par-1/11111111-1111-1111-1111-111111111111
```
