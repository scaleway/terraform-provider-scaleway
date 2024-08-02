---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit"
---
# scaleway_cockpit


~> **Important:**  The data source `scaleway_cockpit` has been deprecated and will no longer be supported. Instead, use resource `scaleway_cockpit`.

-> **Note:**
As of April 2024, Cockpit has introduced regionalization to offer more flexibility and resilience.
If you have customized dashboards in Grafana for monitoring Scaleway resources, please update your queries to accommodate the new regionalized [data sources](../resources/cockpit_source.md).

Gets information about the Scaleway Cockpit.

For more information consult the [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/).

## Example Usage

```hcl
// Get default project's cockpit
data "scaleway_cockpit" "main" {}
```

```hcl
// Get a specific project's cockpit
data "scaleway_cockpit" "main" {
project_id = "11111111-1111-1111-1111-111111111111"
}
```

## Arguments Reference

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the cockpit is associated with.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `plan_id` - (Deprecated) The ID of the current plan
- `endpoints` - (Deprecated) Endpoints
- `metrics_url` - (Deprecated) The metrics URL
- `logs_url` - (Deprecated) The logs URL
- `alertmanager_url` - (Deprecated) The alertmanager URL
- `grafana_url` - (Deprecated) The grafana URL