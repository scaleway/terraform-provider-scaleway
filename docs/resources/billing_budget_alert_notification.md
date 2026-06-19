---
subcategory: "Billing"
page_title: "Scaleway: scaleway_billing_budget_alert_notification"
---

# Resource: scaleway_billing_budget_alert_notification

Creates and manages Scaleway Budget Alert Notifications.

A Budget Alert Notification defines how to notify recipients when a budget alert is triggered.



## Example Usage

```terraform
resource "scaleway_billing_budget" "main" {
  organization_id   = "11111111-1111-1111-1111-111111111111"
  consumption_limit = 10000
  enabled           = true
}

resource "scaleway_billing_budget_alert" "main" {
  budget_id = scaleway_billing_budget.main.id
  threshold = 80
}

resource "scaleway_billing_budget_alert_notification" "email" {
  budget_alert_id = scaleway_billing_budget_alert.main.id
  email_addresses = ["alerts@example.com", "billing@example.com"]
}

resource "scaleway_billing_budget_alert_notification" "sms" {
  budget_alert_id   = scaleway_billing_budget_alert.main.id
  sms_phone_numbers = ["+33612345678"]
}

resource "scaleway_billing_budget_alert_notification" "webhook" {
  budget_alert_id = scaleway_billing_budget_alert.main.id
  webhook_urls    = ["https://example.com/webhook"]
}
```



## Argument Reference

- `budget_alert_id` - (Required) The ID of the budget alert to create notification for.
- `sms_phone_numbers` - (Optional) List of phone numbers to receive SMS notifications. Precisely one of sms_phone_numbers, email_addresses, or webhook_urls must be set.
- `email_addresses` - (Optional) List of email addresses to receive email notifications. Precisely one of sms_phone_numbers, email_addresses, or webhook_urls must be set.
- `webhook_urls` - (Optional) List of webhook URLs to receive webhook notifications. Precisely one of sms_phone_numbers, email_addresses, or webhook_urls must be set.
- `organization_id` - (Optional) The organization ID. If not provided, the default organization configured in the provider is used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the budget alert notification
- `created_at` - The date and time of budget alert notification creation
- `updated_at` - The date and time when the budget alert notification was last updated
- `type` - The type of notification (sms, email, or webhook)

## Import

Budget Alert Notification can be imported using the notification ID.

```bash
terraform import scaleway_billing_budget_alert_notification.main 11111111-1111-1111-1111-111111111111
```
