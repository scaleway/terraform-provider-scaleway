---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_alert_manager"
---

# Resource: scaleway_cockpit_alert_manager

Creates and manages Scaleway Cockpit Alert Managers.

For more information consult the [documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/#grafana-users).

## Example Usage

```terraform

resource "scaleway_account_project" "project" {
  name = "tf_test_project"
}

resource "scaleway_cockpit_alert_manager" "alert_manager" {
  project_id = scaleway_account_project.project.id
  enable_managed_alerts     = true
  emails     = ["alert1@example.com", "alert2@example.com"]
}
```


## Argument Reference

- `enable_managed_alerts` - (Optional, Boolean) Indicates whether the alert manager should be enabled. Defaults to true.
- `emails` - (Optional, List of String) A list of email addresses for the alert receivers.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the cockpit is associated with.
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which alert_manager should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `alert_manager_url` - Alert manager URL.


## Import

Alert managers can be imported using the project ID, e.g.

```bash
$ terraform import scaleway_cockpit_alert_manager.main fr-par/11111111-1111-1111-1111-111111111111
```
