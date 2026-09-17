---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_grafana_product_dashboards"
---

# Data Source: scaleway_cockpit_grafana_product_dashboards

Gets the list of Scaleway product dashboards available in Cockpit Grafana for a project.

Use this data source to discover dashboard names, titles, URLs, tags, and variables for automation (runbooks, portals). Filter with `tags` when you only need dashboards for a given product family.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### List all product dashboards

```terraform
data "scaleway_cockpit_grafana_product_dashboards" "all" {
  project_id = var.project_id
}

output "dashboard_urls" {
  value = {
    for d in data.scaleway_cockpit_grafana_product_dashboards.all.dashboards
    : d.name => d.url
  }
}
```

### Filter by tags

```terraform
data "scaleway_cockpit_grafana_product_dashboards" "rdb" {
  project_id = var.project_id
  tags       = ["rdb"]
}
```

## Argument Reference

- `project_id` - (Optional) The ID of the project to list dashboards for. Defaults to the project configured in the provider.
- `tags` - (Optional) Filter dashboards by tags (for example `rdb`, `lb`).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The project ID.
- `dashboards` - List of Grafana product dashboards. (see [below](#nestedatt--dashboards))
- `names` - List of dashboard names (useful with `scaleway_cockpit_grafana_product_dashboard`).

<a id="nestedatt--dashboards"></a>
### Nested Schema for `dashboards`

- `name` - Dashboard name.
- `title` - Human-readable dashboard title.
- `url` - URL to open the dashboard in Grafana.
- `tags` - Dashboard tags.
- `variables` - Dashboard variables.
