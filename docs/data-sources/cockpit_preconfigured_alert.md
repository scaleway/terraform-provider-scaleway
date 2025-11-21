---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_preconfigured_alert"
---

# Data Source: scaleway_cockpit_preconfigured_alert

Gets information about preconfigured alert rules available in Scaleway Cockpit.

Preconfigured alerts are ready-to-use alert rules that monitor common metrics for Scaleway services.
You can enable these alerts in your Alert Manager using the `scaleway_cockpit_alert_manager` resource.

For more information, refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api).

## Example Usage

### Basic usage

```terraform
data "scaleway_cockpit_preconfigured_alert" "main" {
  project_id = scaleway_account_project.project.id
}

output "available_alerts" {
  value = data.scaleway_cockpit_preconfigured_alert.main.alerts
}
```

### Filter by status

```terraform
data "scaleway_cockpit_preconfigured_alert" "enabled" {
  project_id  = scaleway_account_project.project.id
  rule_status = "enabled"
}

data "scaleway_cockpit_preconfigured_alert" "disabled" {
  project_id  = scaleway_account_project.project.id
  rule_status = "disabled"
}
```

### Use with Alert Manager

```terraform
resource "scaleway_account_project" "project" {
  name = "my-observability-project"
}

resource "scaleway_cockpit" "main" {
  project_id = scaleway_account_project.project.id
}

data "scaleway_cockpit_preconfigured_alert" "all" {
  project_id = scaleway_cockpit.main.project_id
}

resource "scaleway_cockpit_alert_manager" "main" {
  project_id = scaleway_cockpit.main.project_id
  
  # Enable specific alerts by their preconfigured_rule_id
  preconfigured_alert_ids = [
    for alert in data.scaleway_cockpit_preconfigured_alert.all.alerts :
    alert.preconfigured_rule_id
    if alert.product_name == "instance" && alert.rule_status == "disabled"
  ]

  contact_points {
    email = "alerts@example.com"
  }
}
```

## Argument Reference

- `project_id` - (Optional) The ID of the project the alerts are associated with. If not provided, the default project configured in the provider is used.
- `region` - (Optional, defaults to provider region) The region in which the alerts exist.
- `data_source_id` - (Optional) Filter alerts by data source ID.
- `rule_status` - (Optional) Filter alerts by rule status. Valid values are `enabled` or `disabled`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the resource (project ID with region).
- `alerts` - List of preconfigured alerts. Each alert contains:
  - `name` - Name of the alert rule.
  - `rule` - PromQL expression defining the alert condition.
  - `duration` - Duration for which the condition must be true before the alert fires (e.g., "5m").
  - `rule_status` - Status of the alert rule (`enabled`, `disabled`, `enabling`, `disabling`).
  - `state` - Current state of the alert (`inactive`, `pending`, `firing`).
  - `annotations` - Map of annotations attached to the alert.
  - `preconfigured_rule_id` - Unique identifier of the preconfigured rule. Use this ID in `scaleway_cockpit_alert_manager` resource.
  - `display_name` - Human-readable name of the alert.
  - `display_description` - Human-readable description of the alert.
  - `product_name` - Scaleway product associated with the alert (e.g., "instance", "rdb", "kubernetes").
  - `product_family` - Family of the product (e.g., "compute", "storage", "network").
  - `data_source_id` - ID of the data source containing the alert rule.


