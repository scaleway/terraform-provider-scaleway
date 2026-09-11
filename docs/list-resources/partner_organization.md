---
page_title: "Scaleway: scaleway_partner_organization"
subcategory: "Partner"
description: |-
  Lists Scaleway Partner Organizations managed by the caller.
---

# Resource: scaleway_partner_organization

Lists partner organizations managed by the caller.

For more information, see the [Partner documentation](https://www.scaleway.com/en/docs/partner/).

## Example Usage

```terraform
# List all organizations managed by the partner
list "scaleway_partner_organization" "all" {
  provider = scaleway

  config {}
}
```

```terraform
# List organizations managed by the partner filtered by customer ID
list "scaleway_partner_organization" "by_customer" {
  provider = scaleway

  config {
    customer_id = "customer-123"
  }
}
```

```terraform
# List organizations managed by the partner filtered by status
list "scaleway_partner_organization" "by_status" {
  provider = scaleway

  config {
    status = "opened"
  }
}
```

```terraform
# List organizations managed by the partner, specifying the partner ID explicitly
list "scaleway_partner_organization" "by_partner" {
  provider = scaleway

  config {
    partner_id  = "11111111-1111-1111-1111-111111111111"
    customer_id = "customer-123"
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `order_by` - (Optional) Order by field. Can be `created_at_asc` or `created_at_desc`.
- `status` - (Optional) Filter by status. Can be `unknown_status`, `opened`, `locked` or `closed`.
- `email` - (Optional) Filter by email.
- `customer_id` - (Optional) Filter by customer ID.
- `partner_id` - (Optional) The partner organization ID. Defaults to the provider's default Organization ID. This is the same as your Organization ID.
- `locked_by` - (Optional) Filter by `locked_by`. Can be `unknown_locked_by`, `partner` or `scaleway`.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each organization:

- `id` - The ID of the organization.
- `email` - The email of the organization owner.
- `organization_name` - The name of the organization.
- `partner_id` - The partner's organization ID.
- `owner_firstname` - The first name of the organization owner.
- `owner_lastname` - The last name of the organization owner.
- `phone_number` - The phone number of the organization owner.
- `customer_id` - The custom ID for the customer.
- `siren_number` - The SIREN number of the customer.
- `status` - The current status of the organization.
- `locked_by` - Originator of the lock (partner or scaleway).
- `lock_reason_message` - Human-readable reason if the organization is locked.
- `created_at` - Date of organization creation.
- `locked_at` - Date of lock.
- `picture_link` - Link to the organization's picture.
- `comment` - A comment about the organization.
