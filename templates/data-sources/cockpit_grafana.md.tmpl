---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_grafana"
---

# Data Source: scaleway_cockpit_grafana

Gets information about Scaleway Cockpit's Grafana instance for a specific project.

This data source provides the Grafana URL and project details. Authentication is managed through [Scaleway IAM (Identity and Access Management)](https://www.scaleway.com/en/docs/identity-and-access-management/iam/).

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### Basic usage

```terraform
data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_account_project.project.id
}

output "grafana_url" {
  value       = data.scaleway_cockpit_grafana.main.grafana_url
  description = "Access Grafana using your Scaleway IAM credentials"
}
```

### Using with default project

```terraform
# Uses the default project from provider configuration
data "scaleway_cockpit_grafana" "main" {}

output "grafana_url" {
  value = data.scaleway_cockpit_grafana.main.grafana_url
}
```

### Complete example with Cockpit setup

```terraform
resource "scaleway_account_project" "project" {
  name = "my-observability-project"
}

resource "scaleway_cockpit" "main" {
  project_id = scaleway_account_project.project.id
}

data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_cockpit.main.project_id
}

output "grafana_connection_info" {
  value = {
    url        = data.scaleway_cockpit_grafana.main.grafana_url
    project_id = data.scaleway_cockpit_grafana.main.project_id
  }
  description = "Use your Scaleway IAM credentials to authenticate"
}
```

## Argument Reference

- `project_id` - (Optional) The ID of the project the Grafana instance is associated with. If not provided, the default project configured in the provider is used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the project (same as `project_id`).
- `grafana_url` - The URL to access the Grafana dashboard. Use your Scaleway IAM credentials to authenticate.

## Authentication

To access Grafana, use your Scaleway IAM credentials:

1. Navigate to the `grafana_url` provided by this data source
2. Sign in using your Scaleway account (IAM authentication)
3. Your access level is determined by your IAM permissions on the project

For more information about IAM authentication, see the [Scaleway IAM documentation](https://www.scaleway.com/en/docs/identity-and-access-management/iam/).

