---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit"
---

# Resource: scaleway_cockpit

-> **Note:**
As of April 2024, Cockpit has introduced [regionalization](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#region) to offer more flexibility and resilience.
If you have created customized dashboards with data for your Scaleway resources before April 2024, you will need to update your queries in Grafana, with the new regionalized [data sources](../resources/cockpit_source.md).

The `scaleway_cockpit` resource allows you to create and manage Scaleway Cockpit instances.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

For more information consult the [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/).

## Use Cockpit

### Manage Cockpit in the Scaleway default Project

```terraform
// Activate Cockpit in the default Project
resource "scaleway_cockpit" "main" {}
```

### Manage Cockpit in a specific Project

```terraform
// Activate Cockpit in a specific Project
resource "scaleway_cockpit" "main" {
  project_id = "11111111-1111-1111-1111-111111111111"
}
```

### Choose a specific pricing plan for Cockpit

```terraform
resource "scaleway_cockpit" "main" {
  project_id = "11111111-1111-1111-1111-111111111111"
  plan       = "premium"
}
```

### Use the Grafana Terraform provider

```terraform
// Use the Grafana Terraform provider to create a Grafana user and a Grafana folder in the default Project's Cockpit
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

## Arguments reference

- `project_id` - - `project_id` - (Defaults to the Project specified in the [provider's configuration](../index.md#project_id)) The ID of the Project the Cockpit is associated with.
- `plan` - (Optional) Name of the plan to use. Available plans are: free, premium, and custom.

## Attributes reference

This section lists the attributes that are automatically exported when the `scaleway_cockpit` resource is created:

- `plan_id` - (Deprecated) The ID of the current pricing plan.
- `endpoints` - (Deprecated) A list of [endpoints](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#endpoints) related to Cockpit, each with specific URLs:
    - `metrics_url` - (Deprecated) URL for [metrics](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#metric) to retrieve in the [Data sources tab](https://console.scaleway.com/cockpit/dataSource) of the Scaleway console.
    - `logs_url` - (Deprecated) URL for [logs](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#logs) to retrieve in the [Data sources tab](https://console.scaleway.com/cockpit/dataSource) of the Scaleway console.
    - `alertmanager_url` - (Deprecated) URL for the [Alert manager](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#alert-manager).
    - `grafana_url` - (Deprecated) URL for Grafana.
    - `traces_url` - (Deprecated) URL for [traces](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#traces) to retrieve in the [Data sources tab](https://console.scaleway.com/cockpit/dataSource) of the Scaleway console.

## Import

This section explains how to import Cockpits using its `{project_id}`:

```bash
terraform import scaleway_cockpit.main 11111111-1111-1111-1111-111111111111
```