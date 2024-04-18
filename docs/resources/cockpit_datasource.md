---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_datasource"
---

# Resource: scaleway_cockpit_datasource

Creates and manages Scaleway Cockpit Datasources.

For more information consult the [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#data-sources).

## Example Usage

```terraform
resource "scaleway_account_project" "project" {
    name = "test project datasource"
}

resource "scaleway_cockpit_datasource" "main" {
    project_id = scaleway_account_project.project.id
    name       = "my-datasource"
    type       = "metrics"
}
```

## Argument Reference

- `name` - (Required) The name of the cockpit datasource.
- `type` - (Required) The type of the cockpit datasource.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) of the cockpit datasource.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the cockpit datasource is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the cockpit datasource.

~> **Important:** cockpit datasources' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111

- `url` - The URL of the cockpit datasource.
- `origin` - The origin of the cockpit datasource.
- `synchronized_with_grafana` - Indicates whether the data source is synchronized with Grafana.
- `created_at` - Date and time of the cockpit datasource's creation (RFC 3339 format).
- `updated_at` - Date and time of the cockpit datasource's last update (RFC 3339 format).

## Import

Cockpits Datasources can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_cockpit_datasource.main fr-par/11111111-1111-1111-1111-111111111111
```
