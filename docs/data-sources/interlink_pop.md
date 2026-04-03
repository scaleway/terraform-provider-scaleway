---
subcategory: "Interlink"
page_title: "Scaleway: scaleway_interlink_pop"
---

# scaleway_interlink_pop (Data Source)

Gets information about an Interlink PoP (Point of Presence).

A PoP is a physical location where Scaleway infrastructure connects to external networks. PoPs host connections that can be used to create links between your Scaleway VPC and external networks.



## Example Usage

```terraform
# Retrieve a PoP by its ID
data "scaleway_interlink_pop" "by_id" {
  pop_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve a PoP by name
data "scaleway_interlink_pop" "by_name" {
  name = "DC2"
}
```




## Argument Reference

- `pop_id` - (Optional) The ID of the PoP. Conflicts with `name`.
- `name` - (Optional) The name of the PoP to filter for. Conflicts with `pop_id`.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the PoP exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the PoP.
- `hosting_provider_name` - Name of the PoP's hosting provider (e.g., Telehouse, OpCore).
- `address` - Physical address of the PoP.
- `city` - City where the PoP is located.
- `logo_url` - URL of the PoP's logo.
- `available_link_bandwidths_mbps` - List of available bandwidth options in Mbps for hosted links.
- `display_name` - Human-readable display name including location information.
