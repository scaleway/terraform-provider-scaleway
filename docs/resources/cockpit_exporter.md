---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_exporter"
---

# Resource: scaleway_cockpit_exporter

The `scaleway_cockpit_exporter` resource allows you to create and manage [data exports](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#data-exports) in Scaleway's Cockpit. Data exports send metrics and logs from Scaleway products to external destinations like Datadog or OTLP-compatible endpoints.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### Datadog destination

```terraform
data "scaleway_account_project" "project" {
  name = "default"
}

data "scaleway_cockpit_sources" "scaleway_metrics" {
  project_id = data.scaleway_account_project.project.id
  origin     = "scaleway"
  type       = "metrics"
}

resource "scaleway_cockpit_exporter" "main" {
  project_id        = data.scaleway_account_project.project.id
  datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
  name              = "my-datadog-exporter"
  exported_products = ["all"]

  datadog_destination {
    api_key  = var.datadog_api_key
    endpoint = "https://api.datadoghq.com"
  }
}
```

### OTLP destination

```terraform
data "scaleway_account_project" "project" {
  name = "default"
}

data "scaleway_cockpit_sources" "scaleway_metrics" {
  project_id = data.scaleway_account_project.project.id
  origin     = "scaleway"
  type       = "metrics"
}

resource "scaleway_cockpit_source" "otlp_target" {
  project_id     = data.scaleway_account_project.project.id
  name           = "otlp-target"
  type           = "metrics"
  retention_days = 31
}

resource "scaleway_cockpit_exporter" "main" {
  project_id        = data.scaleway_account_project.project.id
  datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
  name              = "my-otlp-exporter"
  exported_products = ["lb", "object-storage", "rdb"]

  otlp_destination {
    endpoint = scaleway_cockpit_source.otlp_target.push_url
  }
}
```

## Argument Reference

- `name` - (Required) Name of the data export.
- `datasource_id` - (Required) ID of the data source linked to the data export. Use [`scaleway_cockpit_sources`](../data-sources/cockpit_sources.md) to find available data sources.
- `datadog_destination` - (Optional) Datadog destination configuration. Cannot be used with `otlp_destination`.
- `otlp_destination` - (Optional) OTLP destination configuration. Cannot be used with `datadog_destination`.
- `exported_products` - (Optional) List of Scaleway products to export. Use `["all"]` to export all products. Use [`scaleway_cockpit_products`](../data-sources/cockpit_products.md) for valid product names. Defaults to `["all"]`.
- `description` - (Optional) Description of the data export.
- `project_id` - (Defaults to the Project ID specified in the [provider configuration](../index.md#project_id)) The ID of the Project.
- `region` - (Defaults to the region specified in the [provider configuration](../index.md#arguments-reference)) The [region](../guides/regions_and_zones.md#regions) where the exporter is located.

### datadog_destination block

- `api_key` - (Required) Datadog API key. Sensitive.
- `endpoint` - (Optional) Datadog endpoint URL. Defaults to `https://api.datadoghq.com`.

### otlp_destination block

- `endpoint` - (Required) OTLP endpoint URL.
- `headers` - (Optional) Headers to include in requests.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the exporter (regional format: `{region}/{id}`).
- `status` - Status of the data export (`creating`, `ready`, `error`).
- `created_at` - Date and time of creation (RFC 3339 format).
- `updated_at` - Date and time of last update (RFC 3339 format).

## Import

Import an exporter using the regional ID:

```bash
terraform import scaleway_cockpit_exporter.main fr-par/11111111-1111-1111-1111-111111111111
```
