---
subcategory: "Billing"
page_title: "Scaleway: scaleway_billing_budget"
---

# Resource: scaleway_billing_budget

Creates and manages Scaleway Budgets.

A Budget allows you to track and control spending across your Scaleway resources.



## Example Usage

```terraform
resource "scaleway_billing_budget" "main" {
  organization_id   = "11111111-1111-1111-1111-111111111111"
  consumption_limit = 10000
  enabled           = true
}
```



## Argument Reference

- `organization_id` - (Optional) The organization ID. If not provided, the default organization configured in the provider is used.
- `consumption_limit` - (Required) Cost limit for the budget in cents.
- `enabled` - (Optional) Whether the budget is enabled or not. Defaults to `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the budget
- `created_at` - The date and time of budget creation
- `updated_at` - The date and time when the budget was last updated

## Import

Budget can be imported using the budget ID.

```bash
terraform import scaleway_billing_budget.main 11111111-1111-1111-1111-111111111111
```
