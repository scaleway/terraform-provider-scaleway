---
subcategory: "Interlink"
page_title: "Scaleway: scaleway_interlink_routing_policy"
---

# scaleway_interlink_routing_policy

Creates and manages Scaleway Interlink Routing Policies.

A routing policy defines IP prefix filters that control which routes are accepted from and advertised to a peer via BGP on an Interlink connection. All routes across a link are blocked by default, so you must attach a routing policy to enable traffic flow.


For more information, see [the Interlink documentation](https://www.scaleway.com/en/docs/network/interlink/) and [API documentation](https://www.scaleway.com/en/developers/api/interlink/).


## Example Usage

```terraform
resource "scaleway_interlink_routing_policy" "main" {
  name              = "my-routing-policy"
  prefix_filter_in  = ["10.0.2.0/24"]
  prefix_filter_out = ["10.0.1.0/24"]
}
```

```terraform
resource "scaleway_interlink_routing_policy" "main" {
  name              = "my-routing-policy-v6"
  is_ipv6           = true
  prefix_filter_in  = ["2001:db8:1::/48"]
  prefix_filter_out = ["2001:db8:2::/48"]
}
```

```terraform
resource "scaleway_interlink_routing_policy" "main" {
  name              = "my-routing-policy"
  prefix_filter_in  = ["10.0.2.0/24", "10.0.3.0/24"]
  prefix_filter_out = ["10.0.1.0/24", "172.16.0.0/16"]
}
```




## Argument Reference

The following arguments are supported:

- `prefix_filter_in` - (Optional) List of IP prefixes (in CIDR notation) to accept from the peer. These are the ranges of route announcements to accept.
- `prefix_filter_out` - (Optional) List of IP prefixes (in CIDR notation) to advertise to the peer. These are the ranges of routes to advertise.
- `name` - (Optional) The name of the routing policy. If not provided, a name will be randomly generated.
- `tags` - (Optional) The list of tags to apply to the routing policy.
- `is_ipv6` - (Optional) Defines whether the routing policy uses IPv6 prefixes. Defaults to `false` (IPv4).
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the routing policy should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the routing policy is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the routing policy.
- `created_at` - The date and time of the creation of the routing policy (RFC 3339 format).
- `updated_at` - The date and time of the last update of the routing policy (RFC 3339 format).
- `organization_id` - The Organization ID the routing policy is associated with.

~> **Important:** Interlink Routing Policies' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

## Import

Interlink Routing Policies can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_interlink_routing_policy.main fr-par/11111111-1111-1111-1111-111111111111
```
