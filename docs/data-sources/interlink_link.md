---
subcategory: "Interlink"
page_title: "Scaleway: scaleway_interlink_link"
---

# scaleway_interlink_link

Gets information about an Interlink Link.

A link is a logical Interlink session created within a PoP, representing the connection between your infrastructure and Scaleway.


For more information, see [the Interlink documentation](https://www.scaleway.com/en/docs/network/interlink/) and [API documentation](https://www.scaleway.com/en/developers/api/interlink/).


## Example Usage

```terraform
# Get link info by ID
data "scaleway_interlink_link" "my_link" {
  link_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Get link info by name
data "scaleway_interlink_link" "my_link" {
  name = "my-link"
}
```




## Argument Reference

- `name` - (Optional) Name of the link. Conflicts with `link_id`.

- `link_id` - (Optional) Unique identifier of the link. Conflicts with `name`.

  -> **Note** You must specify at least one: `name` and/or `link_id`.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the link exists.

- `project_id` - (Optional) Project ID.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - Unique identifier of the link.
- `pop_id` - ID of the PoP where the link's corresponding connection is located.
- `bandwidth_mbps` - Rate limited bandwidth of the link.
- `status` - Status of the link.
- `bgp_v4_status` - Status of the link's BGP IPv4 session.
- `bgp_v6_status` - Status of the link's BGP IPv6 session.
- `vpc_id` - ID of the Scaleway VPC attached to the link.
- `enable_route_propagation` - Defines whether route propagation is enabled or not.
- `partner_id` - ID of the partner facilitating the link.
- `connection_id` - Dedicated physical connection supporting the link.
- `pairing_key` - Used to identify a link from a user or partner's point of view.
- `vlan` - VLAN of the link.
- `peer_asn` - For self-hosted links, the peer AS Number to establish BGP session.
- `routing_policy_v4_id` - ID of the routing policy IPv4 attached to the link.
- `routing_policy_v6_id` - ID of the routing policy IPv6 attached to the link.
- `scw_bgp_config` - BGP configuration on Scaleway's side.
- `peer_bgp_config` - BGP configuration on peer's side (on-premises or other hosting provider).
- `tags` - List of tags associated with the link.
- `created_at` - Creation date of the link.
- `updated_at` - Last modification date of the link.
- `organization_id` - Organization ID.

~> **Important:** Interlink Links IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`
