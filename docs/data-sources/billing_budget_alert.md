---
subcategory: "Billing"
page_title: "Scaleway: scaleway_billing_budget_alert"
---

# Data Source: scaleway_billing_budget_alert

Retrieves information about a Scaleway Budget Alert.

Use this data source to get details of an existing budget alert by its ID.



## Example Usage

```terraform
data "scaleway_billing_budget_alert" "main" {
  alert_id = "11111111-1111-1111-1111-111111111111"
}
```



## Argument Reference

- `alert_id` - (Required) The ID of the budget alert to retrieve.
- `budget_id` - (Optional) The ID of the budget. If not provided, it will be retrieved from the alert.
- `organization_id` - (Optional) The organization ID. If not provided, the default organization configured in the provider is used.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the budget alert
- `threshold` - Threshold percentage above which the alert is sent
- `created_at` - The date and time of budget alert creation
- `updated_at` - The date and time when the budget alert was last updated
