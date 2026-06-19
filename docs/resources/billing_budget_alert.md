---
subcategory: "Billing"
page_title: "Scaleway: scaleway_billing_budget_alert"
---

# Resource: scaleway_billing_budget_alert

Creates and manages Scaleway Budget Alerts.

A Budget Alert triggers notifications when the spending threshold is reached.



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
```



## Argument Reference

- `budget_id` - (Required) The ID of the budget to create alert for.
- `threshold` - (Required) Threshold percentage above which the alert is sent (0-100).
- `organization_id` - (Optional) The organization ID. If not provided, the default organization configured in the provider is used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the budget alert
- `created_at` - The date and time of budget alert creation
- `updated_at` - The date and time when the budget alert was last updated

## Import

Budget Alert can be imported using the alert ID.

```bash
terraform import scaleway_billing_budget_alert.main 11111111-1111-1111-1111-111111111111
```
