### Example Usage

# List invoices starting after a certain date
data "scaleway_billing_invoices" "my-invoices" {
  started_after = "2023-10-01T00:00:00Z"
}

# List invoices by type
data "scaleway_billing_invoices" "my-invoices" {
  invoice_type = "periodic"
}
