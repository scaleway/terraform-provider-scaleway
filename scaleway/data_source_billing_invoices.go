package scaleway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	billing "github.com/scaleway/scaleway-sdk-go/api/billing/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func dataSourceScalewayBillingInvoices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayBillingInvoicesRead,
		Schema: map[string]*schema.Schema{
			"started_after": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Invoice's start date is greater or equal to `started_after`",
			},
			"started_before": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Invoice's start date precedes `started_before`",
			},
			"invoice_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The invoice type. It can either be `periodic` or `purchase`",
			},
			"invoices": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Invoice ID",
						},
						"organization_name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Organization name",
						},
						"start_date": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Start date of the billing period",
						},
						"stop_date": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Stop date of the billing period",
						},
						"billing_period": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Billing period of the invoice in the YYYY-MM format",
						},
						"issued_date": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Date when the invoice was sent to the customer",
						},
						"due_date": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Payment time limit, set according to the Organization's payment conditions",
						},
						"total_untaxed": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Total amount of the invoice, untaxed",
						},
						"total_taxed": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Total amount of the invoice, taxed",
						},
						"total_tax": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "The total tax amount of the invoice",
						},
						"total_discount": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "The total discount amount of the invoice",
						},
						"total_undiscount": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "The total amount of the invoice before applying the discount",
						},
						"invoice_type": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Type of invoice, either periodic or purchase",
						},
						"state": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "The state of the invoice",
						},
						"number": {
							Computed:    true,
							Type:        schema.TypeInt,
							Description: "The invoice number",
						},
						"seller_name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "The name of the seller (Scaleway)",
						},
					},
				},
			},
			"organization_id": organizationIDSchema(),
		},
	}
}

func dataSourceScalewayBillingInvoicesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := billingAPI(m.(*meta.Meta))

	res, err := api.ListInvoices(&billing.ListInvoicesRequest{
		OrganizationID:           expandStringPtr(d.Get("organization_id")),
		BillingPeriodStartAfter:  expandTimePtr(d.Get("started_after").(string)),
		BillingPeriodStartBefore: expandTimePtr(d.Get("started_before").(string)),
		InvoiceType:              billing.InvoiceType(d.Get("invoice_type").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	invoices := []interface{}(nil)
	for _, invoice := range res.Invoices {
		rawInvoice := make(map[string]interface{})
		rawInvoice["id"] = invoice.ID
		rawInvoice["organization_name"] = invoice.OrganizationName
		rawInvoice["start_date"] = flattenTime(invoice.StartDate)
		rawInvoice["stop_date"] = flattenTime(invoice.StopDate)
		rawInvoice["billing_period"] = flattenTime(invoice.BillingPeriod)
		rawInvoice["issued_date"] = flattenTime(invoice.IssuedDate)
		rawInvoice["due_date"] = flattenTime(invoice.DueDate)
		rawInvoice["total_untaxed"] = invoice.TotalUntaxed.String()
		rawInvoice["total_taxed"] = invoice.TotalTaxed.String()
		rawInvoice["total_tax"] = invoice.TotalTax.String()
		rawInvoice["total_discount"] = invoice.TotalDiscount.String()
		rawInvoice["total_undiscount"] = invoice.TotalUndiscount.String()
		rawInvoice["invoice_type"] = invoice.Type.String()
		rawInvoice["state"] = invoice.State
		rawInvoice["number"] = invoice.Number
		rawInvoice["seller_name"] = invoice.SellerName

		invoices = append(invoices, rawInvoice)
	}

	constraints := fmt.Sprintf("%s-%s-%s-%s",
		d.Get("started_after").(string),
		d.Get("started_before").(string),
		d.Get("invoice_type").(string),
		d.Get("organization_id").(string))

	hashedConstraints := sha256.Sum256([]byte(constraints))
	d.SetId(hex.EncodeToString(hashedConstraints[:]))
	_ = d.Set("invoices", invoices)

	return nil
}
