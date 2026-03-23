---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_exporter"
---

# Data Source: scaleway_cockpit_exporter

Gets information about a [data export](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#data-exports) in Scaleway's Cockpit.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### By ID

```terraform
data "scaleway_cockpit_exporter" "main" {
  id = "fr-par/11111111-1111-1111-1111-111111111111"
}
```

### By name and project

```terraform
data "scaleway_cockpit_exporter" "main" {
  project_id = "11111111-1111-1111-1111-111111111111"
  name       = "my-datadog-exporter"
}
```

## Argument Reference

- `id` - (Optional) The regional ID of the exporter (`{region}/{id}`). If set, other filters are ignored.
- `project_id` - (Required unless `id` is set) The project ID.
- `name` - (Optional) The name of the exporter.
- `region` - (Optional) The [region](../guides/regions_and_zones.md#regions). Defaults to the provider region.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The regional ID of the exporter.
- `name` - Name of the data export.
- `description` - Description of the data export.
- `datasource_id` - ID of the linked data source.
- `status` - Status of the data export (`creating`, `ready`, `error`).
- `exported_products` - List of exported products.
- `datadog_destination` - Datadog destination configuration (endpoint only, API key is write-only).
- `otlp_destination` - OTLP destination configuration.
- `created_at` - Creation date (RFC 3339).
- `updated_at` - Last update date (RFC 3339).
