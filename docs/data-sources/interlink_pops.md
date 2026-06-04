---
subcategory: "Interlink"
page_title: "Scaleway: scaleway_interlink_pops"
---

# scaleway_interlink_pops (Data Source)

Gets information about multiple Interlink PoPs (Points of Presence).

A PoP is a physical location where Scaleway infrastructure connects to external networks. Use this data source to list and filter available PoPs for creating interlink connections.



## Example Usage

```terraform
# List all PoPs in a region
data "scaleway_interlink_pops" "all" {
  region = "fr-par"
}

# List PoPs with a specific hosting provider name
data "scaleway_interlink_pops" "by_hosting_provider_name" {
  hosting_provider_name = "OpCore"
}

# List PoPs with dedicated connections available
data "scaleway_interlink_pops" "dedicated" {
  dedicated_available = true
}
```




## Argument Reference

- `name` - (Optional) PoP name to filter for.
- `hosting_provider_name` - (Optional) Hosting provider name to filter for.
- `partner_id` - (Optional) Filter for PoPs hosting an available shared connection from this partner.
- `link_bandwidth_mbps` - (Optional) Filter for PoPs with a shared connection allowing this bandwidth size.
- `dedicated_available` - (Optional) Filter for PoPs with a dedicated connection available for self-hosted links.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) to list PoPs from.

## Attributes Reference

- `pops` - List of PoPs matching the filters. Each entry contains:
    - `id` - ID of the PoP.
    - `name` - Name of the PoP.
    - `hosting_provider_name` - Name of the PoP's hosting provider.
    - `address` - Physical address of the PoP.
    - `city` - City where the PoP is located.
    - `logo_url` - URL of the PoP's logo.
    - `available_link_bandwidths_mbps` - List of available bandwidth options in Mbps.
    - `display_name` - Human-readable display name.
    - `region` - Region of the PoP.
