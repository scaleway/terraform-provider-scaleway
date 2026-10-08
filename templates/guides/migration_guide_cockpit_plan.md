---
page_title: "Migrating from Scaleway Cockpit to the New Infrastructure"
---

# How to Migrate from the Removed `scaleway_cockpit` Resource to the Specialized Cockpit Resources

## Overview

The `scaleway_cockpit` resource has been **removed** from the Scaleway Terraform provider. It was deprecated on the announced date (January 1st, 2025) and has now been purged. This guide provides a step-by-step process to migrate your Terraform configurations from the removed `scaleway_cockpit` resource and transition to the new specialized resources (`scaleway_cockpit_source`, `scaleway_cockpit_alert_manager`, and the `scaleway_cockpit_grafana` data source).

> **Note:**
> The `scaleway_cockpit` resource, as well as its companion `scaleway_cockpit` data source, have been removed. Configurations still using them must be updated before applying the new provider version.

## Prerequisites

### Ensure the Latest Provider Version

Ensure your Scaleway provider is updated to at least version `2.49.0`.

```terraform
terraform {
  required_providers {
    scaleway = {
      source  = "scaleway/scaleway"
      version = "~> 2.49.0"
    }
  }
}

provider "scaleway" {
  # Configuration details
}
```

Run the following command to initialize the updated provider:

```bash
terraform init
```

## Migrating Resources

### Transitioning from `scaleway_cockpit`

The `scaleway_cockpit` resource has been removed. Its functionalities, including endpoint management, are now divided across multiple specialized resources. Below are the steps to migrate:

#### Removed Resource: `scaleway_cockpit`

The following resource is no longer supported and must be replaced in your configurations:

```terraform
resource "scaleway_cockpit" "main" {
  project_id = "11111111-1111-1111-1111-111111111111"
  plan       = "premium"
}
```

#### New Resources

To handle specific functionalities previously managed by `scaleway_cockpit`, you need to use the following resources:

**Data Source Management:**

In the deprecated `scaleway_cockpit` resource, the `plan` argument determined the retention period for logs, metrics, and traces. Now, retention periods are set individually for each data source using the `retention_days` argument in `scaleway_cockpit_source` resources.

```terraform
resource "scaleway_account_project" "project" {
  name = "test project data source"
}

resource "scaleway_cockpit_source" "metrics" {
  project_id     = scaleway_account_project.project.id
  name           = "metrics-source"
  type           = "metrics"
  retention_days = 6 # Customize retention period (1-365 days)
}

resource "scaleway_cockpit_source" "logs" {
  project_id     = scaleway_account_project.project.id
  name           = "logs-source"
  type           = "logs"
  retention_days = 30
}

resource "scaleway_cockpit_source" "traces" {
  project_id     = scaleway_account_project.project.id
  name           = "traces-source"
  type           = "traces"
  retention_days = 15
}
```

**Alert Manager:**

To retrieve the deprecated `alertmanager_url`, you must now explicitly create an Alert Manager using the `scaleway_cockpit_alert_manager` resource:

```terraform
resource "scaleway_cockpit_alert_manager" "alert_manager" {
  project_id            = scaleway_account_project.project.id
  enable_managed_alerts = true

  contact_points {
    email = "alert1@example.com"
  }

  contact_points {
    email = "alert2@example.com"
  }
}
```

**Grafana Access:**

Grafana authentication is managed through Scaleway IAM. To retrieve the Grafana URL, use the `scaleway_cockpit_grafana` data source:

```terraform
data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_account_project.project.id
}
```

### Notes on Regionalization

- As of September 2024, Cockpit resources are regionalized for improved flexibility and resilience. Update your queries in Grafana to use the new regionalized data sources.
- Metrics, logs, and traces now have dedicated resources that allow granular control over retention policies.

### Before and After Example

#### Before: Using `scaleway_cockpit` to Retrieve Endpoints

```terraform
resource "scaleway_cockpit" "main" {
  project_id = "11111111-1111-1111-1111-111111111111"
  plan       = "premium"
}

output "endpoints" {
  value = scaleway_cockpit.main.endpoints
}
```

#### After: Using Specialized Resources

To retrieve all endpoints (metrics, logs, traces, alert manager, and Grafana):

```terraform
resource "scaleway_cockpit_source" "metrics" {
  project_id     = scaleway_account_project.project.id
  name           = "metrics-source"
  type           = "metrics"
  retention_days = 6
}

resource "scaleway_cockpit_source" "logs" {
  project_id     = scaleway_account_project.project.id
  name           = "logs-source"
  type           = "logs"
  retention_days = 30
}

resource "scaleway_cockpit_source" "traces" {
  project_id     = scaleway_account_project.project.id
  name           = "traces-source"
  type           = "traces"
  retention_days = 15
}

resource "scaleway_cockpit_alert_manager" "alert_manager" {
  project_id            = scaleway_account_project.project.id
  enable_managed_alerts = true
}

data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_account_project.project.id
}

output "endpoints" {
  value = {
    metrics       = scaleway_cockpit_source.metrics.url
    logs          = scaleway_cockpit_source.logs.url
    traces        = scaleway_cockpit_source.traces.url
    alert_manager = scaleway_cockpit_alert_manager.alert_manager.alert_manager_url
    grafana       = data.scaleway_cockpit_grafana.main.grafana_url
  }
  description = "Use your Scaleway IAM credentials to authenticate to Grafana"
}
```

## Importing Resources

### Import a Cockpit Source

To import an existing `scaleway_cockpit_source` resource:

```bash
terraform import scaleway_cockpit_source.main fr-par/11111111-1111-1111-1111-111111111111
```

### Grafana Data Source

The `scaleway_cockpit_grafana` data source automatically retrieves Grafana information. No import is required:

```terraform
data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_account_project.project.id
}
```

## Conclusion

By following this guide, you can successfully transition from the removed `scaleway_cockpit` resource to the new set of specialized resources. This ensures compatibility with the latest Terraform provider and Scaleway's updated infrastructure.
