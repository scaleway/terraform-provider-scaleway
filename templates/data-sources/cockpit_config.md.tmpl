---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_config"
---

# Data Source: scaleway_cockpit_config

Gets regional Cockpit configuration, including retention limits and defaults for custom and product data sources.

Use this data source to validate `retention_days` values before creating or updating [`scaleway_cockpit_source`](../resources/cockpit_source.md) resources.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

```terraform
data "scaleway_cockpit_config" "main" {
  region = "fr-par"
}

resource "scaleway_cockpit_source" "metrics" {
  name           = "my-metrics"
  type           = "metrics"
  retention_days = data.scaleway_cockpit_config.main.custom_metrics_retention[0].default_days
}

output "custom_metrics_retention_bounds" {
  value = data.scaleway_cockpit_config.main.custom_metrics_retention[0]
}
```

## Argument Reference

- `region` - (Optional) The [region](../guides/regions_and_zones.md#regions) to query. Defaults to the region configured in the provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The region ID (same as `region`).
- `custom_metrics_retention` - Retention limits and default for custom metrics data sources. (see [below](#nestedatt--custom_metrics_retention))
- `custom_logs_retention` - Retention limits and default for custom logs data sources. (see [below](#nestedatt--custom_logs_retention))
- `custom_traces_retention` - Retention limits and default for custom traces data sources. (see [below](#nestedatt--custom_traces_retention))
- `product_metrics_retention` - Retention limits and default for Scaleway product metrics data sources. (see [below](#nestedatt--product_metrics_retention))
- `product_logs_retention` - Retention limits and default for Scaleway product logs data sources. (see [below](#nestedatt--product_logs_retention))

<a id="nestedatt--custom_metrics_retention"></a>
### Nested Schema for `custom_metrics_retention`, `custom_logs_retention`, `custom_traces_retention`, `product_metrics_retention`, `product_logs_retention`

Read-Only:

- `min_days` - Minimum retention in days.
- `max_days` - Maximum retention in days.
- `default_days` - Default retention in days.
