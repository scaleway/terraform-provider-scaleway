---
page_title: "Cockpit Alert Manager Migration Guide"
---

# Cockpit Alert Manager Migration Guide

This guide explains how to migrate from the deprecated `enable_managed_alerts` field to the new `preconfigured_alert_ids` field in the `scaleway_cockpit_alert_manager` resource.

## Background

The `enable_managed_alerts` field is being deprecated in favor of a more flexible approach using `preconfigured_alert_ids`. This change provides:

- **Granular control**: Select specific alerts instead of enabling all managed alerts
- **Better visibility**: Explicitly declare which alerts are enabled in your Terraform configuration
- **Improved state management**: Terraform accurately tracks which alerts are active

## Migration Steps

### Before Migration (Deprecated)

```terraform
resource "scaleway_cockpit_alert_manager" "main" {
  project_id            = scaleway_account_project.project.id
  enable_managed_alerts = true

  contact_points {
    email = "alerts@example.com"
  }
}
```

### After Migration (Recommended)

#### Step 1: List Available Preconfigured Alerts

Use the data source to discover available alerts:

```terraform
data "scaleway_cockpit_preconfigured_alert" "all" {
  project_id = scaleway_account_project.project.id
}

output "available_alerts" {
  value = data.scaleway_cockpit_preconfigured_alert.all.alerts
}
```

Run `terraform apply` and review the output to see available alerts.

#### Step 2: Select Specific Alerts

Choose the alerts you want to enable:

```terraform
resource "scaleway_cockpit_alert_manager" "main" {
  project_id = scaleway_account_project.project.id

  # Enable specific alerts by product/family
  preconfigured_alert_ids = [
    for alert in data.scaleway_cockpit_preconfigured_alert.all.alerts :
    alert.preconfigured_rule_id
    if contains(["PostgreSQL", "MySQL"], alert.product_name)
  ]

  contact_points {
    email = "alerts@example.com"
  }
}
```

Or use specific alert IDs:

```terraform
resource "scaleway_cockpit_alert_manager" "main" {
  project_id = scaleway_account_project.project.id

  preconfigured_alert_ids = [
    "6c6843af-1815-46df-9e52-6feafcf31fd7", # PostgreSQL Too Many Connections
    "eb8a941e-698d-47d6-b62d-4b6c13f7b4b7", # MySQL Too Many Connections
  ]

  contact_points {
    email = "alerts@example.com"
  }
}
```

## Filtering Alerts

### By Product Name

```terraform
preconfigured_alert_ids = [
  for alert in data.scaleway_cockpit_preconfigured_alert.all.alerts :
  alert.preconfigured_rule_id
  if alert.product_name == "Kubernetes"
]
```

### By Product Family

```terraform
preconfigured_alert_ids = [
  for alert in data.scaleway_cockpit_preconfigured_alert.all.alerts :
  alert.preconfigured_rule_id
  if alert.product_family == "Managed Databases"
]
```

### Multiple Criteria

```terraform
preconfigured_alert_ids = [
  for alert in data.scaleway_cockpit_preconfigured_alert.all.alerts :
  alert.preconfigured_rule_id
  if alert.product_family == "Load Balancer" && alert.product_name == "LB"
]
```

## Important Notes

### Behavioral Changes

- **No automatic alerts**: Unlike `enable_managed_alerts = true`, the API will not automatically enable additional alerts
- **Explicit configuration**: You must explicitly list all alerts you want to enable
- **State accuracy**: Terraform state will only track alerts you've configured

### Compatibility

- The deprecated `enable_managed_alerts` field will be removed in a future major version
- Both fields can coexist during migration, but `preconfigured_alert_ids` takes precedence
- If neither field is specified, no preconfigured alerts will be enabled

## Troubleshooting

### "Insufficient permissions" Error

If you see permission errors when using the `scaleway_cockpit_preconfigured_alert` data source, ensure your IAM policy includes:

```json
{
  "permission_sets": [
    {
      "name": "CockpitManager",
      "permissions": [
        "read:cockpit"
      ]
    }
  ]
}
```

### Unexpected State Changes

If Terraform shows unexpected changes to `preconfigured_alert_ids`:

1. Verify the alert IDs still exist by querying the data source
2. Check that alerts are in `enabled` or `enabling` state
3. Ensure no manual changes were made outside Terraform

## Additional Resources

- [Cockpit Alert Manager Resource Documentation](../resources/cockpit_alert_manager.md)
- [Cockpit Preconfigured Alert Data Source Documentation](../data-sources/cockpit_preconfigured_alert.md)
- [Scaleway Cockpit Documentation](https://www.scaleway.com/en/docs/observability/cockpit/)
