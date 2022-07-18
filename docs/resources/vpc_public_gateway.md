---
page_title: "Scaleway: scaleway_vpc_public_gateway"
description: |-
  Manages Scaleway VPC Public Gateways.
---

# scaleway_vpc_public_gateway

Creates and manages Scaleway VPC Public Gateway.
For more information, see [the documentation](https://developers.scaleway.com/en/products/vpc-gw/api/v1).

## Example

```hcl
resource "scaleway_vpc_public_gateway" "main" {
    name = "public_gateway_demo"
    type = "VPC-GW-S"
    tags = ["demo", "terraform"]
}
```

## Arguments Reference

The following arguments are supported:

- `type` - (Required) The gateway type.
- `name` - (Optional) The name of the public gateway. If not provided it will be randomly generated.
- `tags` - (Optional) The tags associated with the public gateway.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the public gateway should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the public gateway is associated with.
- `upstream_dns_servers` - (Optional) override the gateway's default recursive DNS servers, if DNS features are enabled.
- `ip_id` - (Optional) attach an existing flexible IP to the gateway
- `bastion_enabled` - (Optional) Enable SSH bastion on the gateway
- `bastion_port` - (Optional) The port on which the SSH bastion will listen.
- `enable_smtp` - (Optional) Enable SMTP on the gateway

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the public gateway.
- `organization_id` - The organization ID the public gateway is associated with.
- `created_at` - The date and time of the creation of the public gateway.
- `updated_at` - The date and time of the last update of the public gateway.

## Import

Public gateway can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_vpc_public_gateway.main fr-par-1/11111111-1111-1111-1111-111111111111
```
