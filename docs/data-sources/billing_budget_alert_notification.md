---
subcategory: "Billing"
page_title: "Scaleway: scaleway_billing_budget_alert_notification"
---

# Data Source: scaleway_billing_budget_alert_notification

Retrieves information about a Scaleway Budget Alert Notification.

Use this data source to get details of an existing budget alert notification by its ID.



## Example Usage

```terraform
data "scaleway_billing_budget_alert_notification" "main" {
  notification_id = "11111111-1111-1111-1111-111111111111"
}
```



## Argument Reference

- `notification_id` - (Required) The ID of the budget alert notification to retrieve.
- `budget_alert_id` - (Optional) The ID of the budget alert. If not provided, it will be retrieved from the notification.
- `organization_id` - (Optional) The organization ID. If not provided, the default organization configured in the provider is used.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the budget alert notification
- `created_at` - The date and time of budget alert notification creation
- `updated_at` - The date and time when the budget alert notification was last updated
- `type` - The type of notification (sms, email, or webhook)
- `recipients` - List of recipients for this notification
