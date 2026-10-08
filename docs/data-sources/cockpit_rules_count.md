---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_rules_count"
---

# Data Source: scaleway_cockpit_rules_count

Gets a detailed count of enabled alerting and recording rules in a Cockpit project (preconfigured vs custom totals, and counts by data source).

Use this data source for governance, quotas, and drift checks alongside [`scaleway_cockpit_alert_manager`](../resources/cockpit_alert_manager.md).

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

```terraform
data "scaleway_cockpit_rules_count" "main" {
  region     = "fr-par"
  project_id = "11111111-1111-1111-1111-111111111111"
}

output "custom_rules" {
  value = data.scaleway_cockpit_rules_count.main.custom_rules_count
}
```

## Argument Reference

- `project_id` - (Optional) The ID of the project to retrieve the rule count for. If not provided, the default project configured in the provider is used.
- `region` - (Optional, Computed, Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which the rules exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the data source (region and project ID).
- `preconfigured_rules_count` - Total count of preconfigured rules.
- `custom_rules_count` - Total count of custom rules.
- `rules_count_by_datasource` - Total count of rules grouped by data source. (see [below](#nestedatt--rules_count_by_datasource))

<a id="nestedatt--rules_count_by_datasource"></a>
### Nested Schema for `rules_count_by_datasource`

- `data_source_id` - ID of the data source.
- `data_source_name` - Name of the data source.
- `rules_count` - Total count of rules associated with this data source.
