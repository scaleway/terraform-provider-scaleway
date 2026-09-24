---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_grafana_product_dashboard"
---

# Data Source: scaleway_cockpit_grafana_product_dashboard

Gets information about a single Scaleway product dashboard in Cockpit Grafana.

Use this data source when you know the dashboard name (from the console or from [`scaleway_cockpit_grafana_product_dashboards`](cockpit_grafana_product_dashboards.md)) and need its URL, title, tags, or variables.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### Basic usage

```terraform
data "scaleway_cockpit_grafana_product_dashboard" "rdb" {
  project_id     = var.project_id
  dashboard_name = "rdb"
}

output "rdb_dashboard_url" {
  value = data.scaleway_cockpit_grafana_product_dashboard.rdb.url
}
```

### From the list data source

```terraform
data "scaleway_cockpit_grafana_product_dashboards" "all" {
  project_id = var.project_id
  tags       = ["rdb"]
}

data "scaleway_cockpit_grafana_product_dashboard" "first" {
  project_id     = var.project_id
  dashboard_name = data.scaleway_cockpit_grafana_product_dashboards.all.names[0]
}
```

## Argument Reference

- `project_id` - (Optional) The ID of the project the dashboard belongs to. Defaults to the project configured in the provider.
- `dashboard_name` - (Required) Name of the Grafana product dashboard to retrieve.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the data source (`{project_id}/{name}`).
- `name` - Dashboard name.
- `title` - Human-readable dashboard title.
- `url` - URL to open the dashboard in Grafana.
- `tags` - Dashboard tags.
- `variables` - Dashboard variables.
