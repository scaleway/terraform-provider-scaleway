---
subcategory: "Billing"
page_title: "Scaleway: scaleway_billing_invoices"
---

# scaleway_billing_invoices

Gets information about your Invoices.

## Example Usage

```hcl
# List invoices starting after a certain date
data "scaleway_billing_invoices" "my-invoices" {
  started_after = "2023-10-01T00:00:00Z"
}

# List invoices by type
data "scaleway_billing_invoices" "my-invoices" {
  invoice_type = "periodic"
}
```

## Argument Reference

- `started_after` - (Optional) Invoices with a start date that are greater or equal to `started_after` are listed (RFC 3339 format).

- `started_before` - (Optional) Invoices with a start date that precedes `started_before` are listed (RFC 3339 format).

- `invoice_type` - (Optional) Invoices with the given type are listed. Valid values are `periodic` and `purchase`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `invoices` - List of found invoices
    - `id` - The associated invoice ID.
    - `start_date` - The start date of the billing period (RFC 3339 format).
    - `issued_date` - The date when the invoice was sent to the customer (RFC 3339 format).
    - `due_date` - The payment time limit, set according to the Organization's payment conditions (RFC 3339 format).
    - `total_untaxed` - The total amount, untaxed.
    - `total_taxed` - The total amount, taxed.
    - `invoice_type` - The type of invoice.
    - `number` - The invoice number.