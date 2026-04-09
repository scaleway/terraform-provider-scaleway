---
subcategory: "Interlink"
page_title: "Scaleway: scaleway_interlink_partners"
---

# scaleway_interlink_partners (Data Source)

Gets information about multiple Interlink Partners.

A partner is an organization that provides shared connections at PoPs. Use this data source to list and filter available partners for creating hosted links.

## Example Usage

```terraform
# List all partners in a region
data "scaleway_interlink_partners" "all" {
  region = "fr-par"
}

# List partners available at specific PoPs
data "scaleway_interlink_partners" "at_pops" {
  pop_ids = [
    data.scaleway_interlink_pop.main.id,
  ]
}
```



## Argument Reference

- `pop_ids` - (Optional) Filter for partners present (offering a connection) in one of these PoPs.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) to list partners from.

## Attributes Reference

- `partners` - List of partners matching the filters. Each entry contains:
    - `id` - ID of the partner.
    - `name` - Name of the partner.
    - `contact_email` - Contact email address.
    - `logo_url` - URL of the partner's logo.
    - `portal_url` - URL of the partner's portal.
    - `created_at` - Creation date.
    - `updated_at` - Last update date.
