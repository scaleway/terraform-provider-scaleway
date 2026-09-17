---
subcategory: "Billing"
page_title: "Scaleway: scaleway_billing_budget"
---

# Data Source: scaleway_billing_budget

Retrieves information about a Scaleway Budget.

Use this data source to get details of an existing budget by its ID.



## Example Usage

```terraform
data "scaleway_billing_budget" "main" {
  budget_id = "11111111-1111-1111-1111-111111111111"
}
```



## Argument Reference

- `budget_id` - (Required) The ID of the budget to retrieve.
- `organization_id` - (Optional) The organization ID. If not provided, the default organization configured in the provider is used.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the budget
- `consumption_limit` - Cost limit for the budget in cents
- `enabled` - Whether the budget is enabled or not
- `created_at` - The date and time of budget creation
- `updated_at` - The date and time when the budget was last updated
