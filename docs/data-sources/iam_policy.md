---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_policy"
---

# scaleway_iam_policy

Use this data source to get information on an existing IAM policy based on its ID.
For more information refer to the [IAM API documentation](https://developers.scaleway.com/en/products/iam/api/).

## Example Usage

```hcl
# Get policy by id
data "scaleway_iam_policy" "find_by_id" {
  policy_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `policy_id` - The ID of the IAM policy.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the IAM policy.
- `created_at` - The date and time of the creation of the IAM policy.
- `updated_at` - The date and time of the last update of the IAM policy.
- `editable` - Whether the IAM policy is editable.
