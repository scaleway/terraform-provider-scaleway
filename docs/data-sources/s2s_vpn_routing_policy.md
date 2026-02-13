---
subcategory: "S2S VPN"
page_title: "Scaleway: scaleway_s2s_vpn_routing_policy"
---

# scaleway_s2s_vpn_routing_policy



For further information refer to the Site-to-Site VPN [API documentation](https://www.scaleway.com/en/developers/api/site-to-site-vpn/).


## Example Usage

```terraform
# Get info by routing policy ID
data "scaleway_s2s_vpn_routing_policy" "my_policy" {
  routing_policy_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Get info by name
data "scaleway_s2s_vpn_routing_policy" "my_policy" {
  name = "foobar"
}
```




## Argument Reference

- `name` - (Optional) The name of the routing policy.

- `routing_policy_id` - (Optional) The routing policy ID.

  -> **Note** You must specify at least one: `name` and/or `routing_policy_id`.

- `region` - (Defaults to [provider](../index.md) `region`) The [region](../guides/regions_and_zones.md#regions) in which the routing policy exists.

- `project_id` - (Optional) The ID of the project the routing policy is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the routing policy.
- `is_ipv6` - Whether the routing policy is for IPv6 prefixes.
- `prefix_filter_in` - List of IP prefixes accepted from the peer gateway.
- `prefix_filter_out` - List of IP prefixes advertised to the peer gateway.
- `tags` - The tags associated with the routing policy.
- `created_at` - The date and time of creation of the routing policy.
- `updated_at` - The date and time of the last update of the routing policy.
- `organization_id` - The Organization ID the routing policy is associated with.

~> **Important:** Routing Policies IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`
