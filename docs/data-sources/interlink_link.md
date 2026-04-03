---
subcategory: "Interlink"
page_title: "Scaleway: <no value>"
---

# <no value> (Data Source)

<no value>

## Argument Reference

- `link_id` - (Optional) The ID of the link. Conflicts with all filter arguments below.

The following arguments can be used to look up a link via the list API. They all conflict with `link_id`:

- `name` - (Optional) The name to filter for.
- `tags` - (Optional) List of tags to filter for.
- `pop_id` - (Optional) The PoP ID to filter for.
- `partner_id` - (Optional) The partner ID to filter for.
- `vpc_id` - (Optional) The VPC ID to filter for.
- `connection_id` - (Optional) The connection ID to filter for.
- `project_id` - (Optional) The ID of the Project to filter for.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the link exists.

## Attributes Reference

Exported attributes are the ones from `scaleway_interlink_link` [resource](../resources/interlink_link.md).
