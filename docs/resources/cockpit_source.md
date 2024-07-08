---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_source"
---

# Resource: scaleway_cockpit_source

Creates and manages Scaleway Cockpit Data Sources.

For more information consult the [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#data-sources).

## Example Usage

```terraform
resource "scaleway_account_project" "project" {
    name = "test project data source"
}

resource "scaleway_cockpit_source" "main" {
    project_id = scaleway_account_project.project.id
    name       = "my-data-source"
    type       = "metrics"
}
```

## Argument Reference

- `name` - (Required) The name of the cockpit data source.
- `type` - (Required) The type of the cockpit data source. Possible values are: `metrics`, `logs` or `traces`.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) of the cockpit datasource.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the cockpit data source is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the cockpit data source.

~> **Important:** cockpit data sources' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111

- `url` - The URL of the cockpit data source.
- `origin` - The origin of the cockpit data source.
- `synchronized_with_grafana` - Indicates whether the data source is synchronized with Grafana.
- `created_at` - Date and time of the cockpit data source's creation (RFC 3339 format).
- `updated_at` - Date and time of the cockpit datas ource's last update (RFC 3339 format).

## Import

Cockpits Data Sources can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_cockpit_source.main fr-par/11111111-1111-1111-1111-111111111111
```
