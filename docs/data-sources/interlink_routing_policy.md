---
subcategory: "Interlink"
page_title: "Scaleway: scaleway_interlink_routing_policy"
---

# scaleway_interlink_routing_policy

Gets information about an Interlink Routing Policy.

A routing policy defines IP prefix filters that control which routes are accepted from and advertised to a peer via BGP on an Interlink connection.


For more information, see [the Interlink documentation](https://www.scaleway.com/en/docs/network/interlink/) and [API documentation](https://www.scaleway.com/en/developers/api/interlink/).


## Example Usage

```terraform
# Get routing policy info by ID
data "scaleway_interlink_routing_policy" "my_policy" {
  routing_policy_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Get routing policy info by name
data "scaleway_interlink_routing_policy" "my_policy" {
  name = "my-routing-policy"
}
```




## Argument Reference

- `name` - (Optional) The name of the routing policy. Conflicts with `routing_policy_id`.

- `routing_policy_id` - (Optional) The ID of the routing policy. Conflicts with `name`.

  -> **Note** You must specify at least one: `name` and/or `routing_policy_id`.

- `region` - (Defaults to [provider](../index.md) `region`) The [region](../guides/regions_and_zones.md#regions) in which the routing policy exists.

- `project_id` - (Optional) The ID of the project the routing policy is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the routing policy.
- `is_ipv6` - Whether the routing policy uses IPv6 prefixes.
- `prefix_filter_in` - List of IP prefixes accepted from the peer.
- `prefix_filter_out` - List of IP prefixes advertised to the peer.
- `tags` - The tags associated with the routing policy.
- `created_at` - The date and time of creation of the routing policy.
- `updated_at` - The date and time of the last update of the routing policy.
- `organization_id` - The Organization ID the routing policy is associated with.

~> **Important:** Interlink Routing Policies IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`
