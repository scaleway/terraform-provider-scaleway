---
subcategory: "Billing"
page_title: "Scaleway: scaleway_billing_consumptions"
---

# scaleway_billing_consumptions

Gets information about your Consumptions.

## Example Usage

```hcl
# Find your detailed monthly consumption list
data "scaleway_billing_consumptions" "my-consumption" {
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

## Argument Reference

- `organization_id` - (Defaults to [provider](../index.md#organization_d) `organization_id`) The ID of the organization the consumption list is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `consumptions` - List of found consumptions
    - `value` - The monetary value of the consumption.
    - `description` - The description of the consumption.
    - `project_id` - The project ID of the consumption.
    - `category` - The category of the consumption.
    - `operation_path` - The unique identifier of the product.
- `updated_at` - The last consumption update date.