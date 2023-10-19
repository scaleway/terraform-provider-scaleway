package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceBillingInvoices_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "scaleway_billing_invoices" "my-invoices" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_billing_invoices.my-invoices", "invoices.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_invoices.my-invoices", "invoices.0.start_date"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_invoices.my-invoices", "invoices.0.total_untaxed"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_invoices.my-invoices", "invoices.0.total_taxed"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_invoices.my-invoices", "invoices.0.invoice_type"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_invoices.my-invoices", "invoices.0.number"),
				),
			},
		},
	})
}
