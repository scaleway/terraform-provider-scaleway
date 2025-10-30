package tem_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceDomain_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainName := "terraform-ds.test.local"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_tem_domain" "main" {
						name 	   = "%s"
						accept_tos = true
					}
					
					data "scaleway_tem_domain" "prod" {
						name = "${scaleway_tem_domain.main.name}"
					}
					
					data "scaleway_tem_domain" "stg" {
						domain_id = "${scaleway_tem_domain.main.id}"
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					isDomainPresent(tt, "data.scaleway_tem_domain.prod"),
					resource.TestCheckResourceAttr("data.scaleway_tem_domain.prod", "name", domainName),

					isDomainPresent(tt, "data.scaleway_tem_domain.stg"),
					resource.TestCheckResourceAttr("data.scaleway_tem_domain.stg", "name", domainName),
				),
			},
		},
	})
}

// TestAccDataSourceDomain_Reputation is now covered by TestAccTEM_Complete step 4
