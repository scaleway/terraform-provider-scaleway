package scaleway

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	billing "github.com/scaleway/scaleway-sdk-go/api/billing/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
							Computed: true,
							Type:     schema.TypeString,
						},
						"start_date": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"issued_date": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"due_date": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"total_untaxed": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"total_taxed": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"invoice_type": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"number": {
							Computed: true,
							Type:     schema.TypeInt,
						},
					},
				},
			},
			"organization_id": organizationIDSchema(),
		},
	}
}

func dataSourceScalewayBillingInvoicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := billingAPI(meta)

	res, err := api.ListInvoices(&billing.ListInvoicesRequest{
		OrganizationID: expandStringPtr(d.Get("organization_id")),
		StartedAfter:   expandTimePtr(d.Get("started_after").(string)),
		StartedBefore:  expandTimePtr(d.Get("started_before").(string)),
		InvoiceType:    billing.InvoiceType(d.Get("invoice_type").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	invoices := []interface{}(nil)
	for _, invoice := range res.Invoices {
		rawInvoice := make(map[string]interface{})
		rawInvoice["id"] = invoice.ID
		rawInvoice["start_date"] = flattenTime(invoice.StartDate)
		rawInvoice["issued_date"] = flattenTime(invoice.IssuedDate)
		rawInvoice["due_date"] = flattenTime(invoice.DueDate)
		rawInvoice["total_untaxed"] = invoice.TotalUntaxed.String()
		rawInvoice["total_taxed"] = invoice.TotalTaxed.String()
		rawInvoice["invoice_type"] = invoice.InvoiceType.String()
		rawInvoice["number"] = invoice.Number

		invoices = append(invoices, rawInvoice)
	}

	constraints := fmt.Sprintf("%s-%s-%s-%s",
		d.Get("started_after").(string),
		d.Get("started_before").(string),
		d.Get("invoice_type").(string),
		d.Get("organization_id").(string))

	hashedConstraints := sha256.Sum256([]byte(constraints))
	d.SetId(fmt.Sprintf("%x", hashedConstraints))
	_ = d.Set("invoices", invoices)

	return nil
}
