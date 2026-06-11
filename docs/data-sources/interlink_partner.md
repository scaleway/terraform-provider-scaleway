---
subcategory: "Interlink"
page_title: "Scaleway: scaleway_interlink_partner"
---

# scaleway_interlink_partner (Data Source)

Gets information about an Interlink Partner.

A partner is an organization that provides shared connections at PoPs, allowing you to create hosted links without owning physical infrastructure.

## Example Usage

```terraform
# Retrieve a partner by its ID
data "scaleway_interlink_partner" "by_id" {
  partner_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
```

```terraform
# Retrieve a partner by name
data "scaleway_interlink_partner" "by_name" {
  name = "FreePro"
}
```



## Argument Reference

- `partner_id` - (Optional) The ID of the partner. Can be a plain UUID or a regional ID. Conflicts with `name`.
- `name` - (Optional) The name of the partner to filter for. Conflicts with `partner_id`.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the partner operates.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the partner.
- `contact_email` - Contact email address of the partner.
- `logo_url` - URL of the partner's logo.
- `portal_url` - URL of the partner's portal.
- `created_at` - Creation date of the partner.
- `updated_at` - Last update date of the partner.
