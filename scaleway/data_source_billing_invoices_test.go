package scaleway

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceBillingInvoices_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	currentDate := time.Now().Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				data "scaleway_billing_invoices" "my-invoices" {
					started_after = "%s"
				}`, currentDate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.scaleway_billing_invoices.my-invoices", "invoices.#", "1"),
					resource.TestCheckResourceAttrSet("data.scaleway_billing_invoices.my-invoices", "invoices.0.id"),
				),
			},
		},
	})
}
