---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit"
---

# Resource: scaleway_cockpit

-> **Note:**
As of April 2024, Cockpit has introduced regionalization to offer more flexibility and resilience.
If you have customized dashboards in Grafana for monitoring Scaleway resources, please update your queries to accommodate the new regionalized [data sources](./cockpit_source.md).

Creates and manages Scaleway Cockpit.

For more information consult the [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/).

## Example Usage

### Manage Cockpit in the default project

```terraform
resource "scaleway_cockpit" "main" {}
```

### Manage Cockpit in a specific project

```terraform
resource "scaleway_cockpit" "main" {
  project_id = "11111111-1111-1111-1111-111111111111"
}
```

### Choose a specific plan for Cockpit

```terraform
resource "scaleway_cockpit" "main" {
project_id = "11111111-1111-1111-1111-111111111111"
plan       = "premium"
}
```

### Use the Grafana Terraform provider

```terraform
resource "scaleway_cockpit" "main" {}

resource "scaleway_cockpit_grafana_user" "main" {
  project_id = scaleway_cockpit.main.project_id
  login      = "example"
  role       = "editor"
}

provider "grafana" {
  url  = scaleway_cockpit.main.endpoints.0.grafana_url
  auth = "${scaleway_cockpit_grafana_user.main.login}:${scaleway_cockpit_grafana_user.main.password}"
}

resource "grafana_folder" "test_folder" {
  title = "Test Folder"
}
```

## Argument Reference

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the cockpit is associated with.
- `plan` - (Optional) Name of the plan to use. Available plans are free, premium, and custom.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `plan_id` - (Deprecated) The ID of the current plan. Please use plan instead.
- `endpoints` - (Deprecated) Endpoints. Please use scaleway_cockpit_source instead.
    - `metrics_url` - (Deprecated) The metrics URL.
    - `logs_url` - (Deprecated) The logs URL.
    - `alertmanager_url` - (Deprecated) The alertmanager URL.
    - `grafana_url` - (Deprecated) The grafana URL.
    - `traces_url` - (Deprecated) The traces URL.

## Import

Cockpits can be imported using the `{project_id}`, e.g.

```bash
$ terraform import scaleway_cockpit.main 11111111-1111-1111-1111-111111111111
```
