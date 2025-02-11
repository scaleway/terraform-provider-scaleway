---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_alert_manager"
---

# Resource: scaleway_cockpit_alert_manager

The `scaleway_cockpit_alert_manager` resource allows you to enable and manage the Scaleway Cockpit [alert manager](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#alert-manager).

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.


## Example Usage

### Enable the alert manager and configure managed alerts

The following commands allow you to:

- enable the alert manager in a Project named `tf_test_project`
- enable [managed alerts](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#managed-alerts)
- set up [contact points](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#contact-points) to receive alert notifications

```terraform

resource "scaleway_account_project" "project" {
  name = "tf_test_project"
}

resource "scaleway_cockpit_alert_manager" "alert_manager" {
  project_id = scaleway_account_project.project.id
  enable_managed_alerts     = true

  contact_points {
      email = "alert1@example.com"
    }

  contact_points {
      email = "alert2@example.com"
    }
}
```


## Argument Reference

This section lists the arguments that are supported:

- `enable_managed_alerts` - (Optional, Boolean) Specifies whether the alert manager should be enabled. Defaults to true.
- `contact_points` - (Optional, List of Map) A list of contact points with email addresses that will receive alerts. Each map should contain a single key email.
- `project_id` - (Defaults to the Project ID specified in the [provider configuration](../index.md#project_id)) The ID of the Project the Cockpit is associated with.
- `region` - (Defaults to the region specified in the [provider configuration](../index.md#arguments-reference)) The [region](../guides/regions_and_zones.md#regions) where the [alert manager](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#alert-manager) should be enabled.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `alert_manager_url` - The URL of the alert manager.


## Import

This section explains how to import alert managers using the ID of the Project associated with Cockpit.

```bash
terraform import scaleway_cockpit_alert_manager.main fr-par/11111111-1111-1111-1111-111111111111
```
