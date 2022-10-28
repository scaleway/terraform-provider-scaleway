package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceTemDomain_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	domainName := "terraform-ds.test.local"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayTemDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_tem_domain" "main" {
						name 	= "%s"
					}
					
					data "scaleway_tem_domain" "prod" {
						name = "${scaleway_tem_domain.main.name}"
					}
					
					data "scaleway_tem_domain" "stg" {
						id = "${scaleway_tem_domain.main.id}"
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayTemDomainExists(tt, "data.scaleway_tem_domain.prod"),
					resource.TestCheckResourceAttr("data.scaleway_tem_domain.prod", "name", domainName),

					testAccCheckScalewayTemDomainExists(tt, "data.scaleway_tem_domain.stg"),
					resource.TestCheckResourceAttr("data.scaleway_tem_domain.stg", "name", domainName),
				),
			},
		},
	})
}
