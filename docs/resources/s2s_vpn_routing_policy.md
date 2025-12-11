---
subcategory: "S2S VPN"
page_title: "Scaleway: scaleway_s2s_vpn_routing_policy"
---

# Resource: scaleway_s2s_vpn_routing_policy

Creates and manages Scaleway Site-to-Site VPN Routing Policies.
A routing policy defines which routes are accepted from and advertised to the peer gateway via BGP.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/site-to-site-vpn/reference-content/understanding-s2svpn/).

## Example Usage

### Basic

```terraform
resource "scaleway_s2s_vpn_routing_policy" "policy" {
  name            = "my-routing-policy"
  prefix_filter_in  = ["10.0.2.0/24"]
  prefix_filter_out = ["10.0.1.0/24"]
}
```

### Multiple Prefixes

```terraform
resource "scaleway_s2s_vpn_routing_policy" "policy" {
  name            = "my-routing-policy"
  prefix_filter_in  = ["10.0.2.0/24", "10.0.3.0/24"]
  prefix_filter_out = ["10.0.1.0/24", "172.16.0.0/16"]
}
```

## Argument Reference

The following arguments are supported:

- `prefix_filter_in` - (Optional) List of IP prefixes (in CIDR notation) to accept from the peer gateway. These are the routes that the customer gateway can announce to Scaleway.
- `prefix_filter_out` - (Optional) List of IP prefixes (in CIDR notation) to advertise to the peer gateway. These are the routes that Scaleway will announce to the customer gateway.
- `name` - (Optional) The name of the routing policy. If not provided, it will be randomly generated.
- `tags` - (Optional) The list of tags to apply to the routing policy.
- `is_ipv6` - (Optional) Defines whether the routing policy is for IPv6 prefixes. Defaults to `false` (IPv4).
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the routing policy should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the routing policy is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the routing policy.
- `created_at` - The date and time of the creation of the routing policy (RFC 3339 format).
- `updated_at` - The date and time of the last update of the routing policy (RFC 3339 format).
- `organization_id` - The Organization ID the routing policy is associated with.

~> **Important:** Routing Policies' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

## Import

Routing Policies can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_s2s_vpn_routing_policy.main fr-par/11111111-1111-1111-1111-111111111111
```
