---
subcategory: "Partner"
page_title: "Scaleway: scaleway_partner_organization"
---

# Resource: scaleway_partner_organization

Create and manage partner organizations in Scaleway. Partner organizations allow you to manage customer organizations through the Scaleway Partner program. This resource enables partners to create, update, and manage organization details including owner information, contact details, and customer identifiers.

Refer to the Scaleway [Partner documentation](https://www.scaleway.com/en/docs/partner-space/) and [API documentation](https://www.scaleway.com/en/developers/api/partner-api/) for more information.


## Example Usage

```terraform
resource "scaleway_partner_organization" "main" {
  email             = "contact@example.com"
  organization_name = "My Organization"
  partner_id        = "11111111-1111-1111-1111-111111111111"
  owner_firstname   = "John"
  owner_lastname    = "Doe"
  customer_id       = "customer-123"
  phone_number      = "+33123456789"
}
```

```terraform
resource "scaleway_partner_organization" "minimal" {
  email             = "contact@example.com"
  organization_name = "Minimal Organization"
  partner_id        = "11111111-1111-1111-1111-111111111111"
  owner_firstname   = "John"
  owner_lastname    = "Doe"
  customer_id       = "customer-minimal"
}
```



## Argument Reference

The following arguments are supported:

- `email` - (Required) The email of the new organization owner.
- `partner_id` - (Optional) Your personal partner_id. This is the same as your Organization ID. If not provided, the default organization configured in the provider is used.
- `organization_name` - (Required) The name of the organization you want to create. Usually the company name.
- `owner_firstname` - (Required) The first name of the new organization owner.
- `owner_lastname` - (Required) The last name of the new organization owner.
- `phone_number` - (Optional) The phone number of the new organization owner.
- `customer_id` - (Required) A custom ID for the customer in your own infrastructure.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the organization resource.
- `status` - The current status of the organization.
- `locked_by` - Originator of the lock (partner or scaleway).
- `lock_reason_message` - Human-readable reason if the organization is locked.
- `created_at` - Date of organization creation.
- `locked_at` - Date of lock.

## Import

Partner organizations can be imported using their `id`.

```bash
terraform import scaleway_partner_organization.main <id>
```
