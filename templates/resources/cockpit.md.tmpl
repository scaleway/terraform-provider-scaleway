---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit"
---

# Resource: scaleway_cockpit

~> **Important:** The resource `scaleway_cockpit` has been deprecated and will no longer be supported after January 1st, 2025. Instead, use the new specialized resources:

- `scaleway_cockpit_source` for managing data sources (metrics, logs, traces)
- `scaleway_cockpit_alert_manager` for managing alert manager
- `scaleway_cockpit_grafana_user` for managing Grafana users

For detailed migration instructions, see the [Cockpit Migration Guide](../guides/migration_guide_cockpit_plan.md).

-> **Note:**
As of September 2024, Cockpit has introduced [regionalization](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#region) to offer more flexibility and resilience.
If you have created customized dashboards with data for your Scaleway resources before April 2024, you will need to update your queries in Grafana, with the new regionalized [data sources](../resources/cockpit_source.md).

-> **Note:**
From January 1st 2025, Cockpit plans have been deprecated. You can now edit the retention period for all your datasources (metrics, logs, and traces) separately. Refer to our product documentation for more information on [possible retention values](https://www.scaleway.com/en/docs/cockpit/concepts/#retention) and [pricing](https://www.scaleway.com/en/docs/cockpit/faq/#how-am-i-billed-for-increasing-data-retention-period).

Please note that even if you provide the grafana_url, it will only be active if a [Grafana user](../resources/cockpit_grafana_user.md) is created first. Make sure to create a Grafana user in your Cockpit instance to enable full access to Grafana.

The `scaleway_cockpit` resource allows you to create and manage Scaleway Cockpit instances.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

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

resource "scaleway_cockpit_grafana_user" "main" {
  project_id = scaleway_cockpit.main.project_id
  login      = "example"
  role       = "editor"
}

resource "scaleway_cockpit" "main" {}

provider "grafana" {
  url  = scaleway_cockpit.main.endpoints.0.grafana_url
  auth = "${scaleway_cockpit_grafana_user.main.login}:${scaleway_cockpit_grafana_user.main.password}"
}

resource "grafana_folder" "test_folder" {
  title = "Test Folder"
}
```

## Argument Reference

- `project_id` - (Defaults to the Project specified in the [provider's configuration](../index.md#project_id)) The ID of the Project the Cockpit is associated with.
- `plan` - (Deprecated) Name of the plan to use. Available plans are: free, premium, and custom.
~> **Important:** The plan field is deprecated. Any modification or selection will have no effect.

~> **Warning:** This entire resource is deprecated and will be removed after January 1st, 2025. Please migrate to the new specialized resources as described in the [Cockpit Migration Guide](../guides/migration_guide_cockpit_plan.md).

## Attributes Reference

~> **Warning:** This entire resource is deprecated and will be removed after January 1st, 2025. All attributes below are deprecated and will no longer be available.

In addition to all arguments above, the following attributes are exported:

- `plan_id` - (Deprecated) The ID of the current pricing plan.
- `endpoints` - (Deprecated) A list of [endpoints](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#endpoints) related to Cockpit, each with specific URLs:
    - `metrics_url` - (Deprecated) URL for [metrics](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#metric) to retrieve in the [Data sources tab](https://console.scaleway.com/cockpit/dataSource) of the Scaleway console.
    - `logs_url` - (Deprecated) URL for [logs](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#logs) to retrieve in the [Data sources tab](https://console.scaleway.com/cockpit/dataSource) of the Scaleway console.
    - `alertmanager_url` - (Deprecated) URL for the [Alert manager](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#alert-manager).
    - `grafana_url` - (Deprecated) URL for Grafana.
    - `traces_url` - (Deprecated) URL for [traces](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#traces) to retrieve in the [Data sources tab](https://console.scaleway.com/cockpit/dataSource) of the Scaleway console.

## Migration

This resource is deprecated and will be removed after January 1st, 2025. To migrate to the new infrastructure, please refer to the [Cockpit Migration Guide](../guides/migration_guide_cockpit_plan.md) which provides step-by-step instructions for transitioning to:

- `scaleway_cockpit_source` for managing data sources (metrics, logs, traces)
- `scaleway_cockpit_alert_manager` for managing alert manager  
- `scaleway_cockpit_grafana_user` for managing Grafana users

## Import

This section explains how to import a Cockpit using its `{project_id}`.

```bash
terraform import scaleway_cockpit.main 11111111-1111-1111-1111-111111111111
```

~> **Note:** Import functionality will be removed when this resource is deprecated.
