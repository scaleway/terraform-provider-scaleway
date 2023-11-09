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
					testAccCheckScalewayTemDomainExists(tt, "data.scaleway_tem_domain.prod"),
					resource.TestCheckResourceAttr("data.scaleway_tem_domain.prod", "name", domainName),

					testAccCheckScalewayTemDomainExists(tt, "data.scaleway_tem_domain.stg"),
					resource.TestCheckResourceAttr("data.scaleway_tem_domain.stg", "name", domainName),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceTemDomain_Reputation(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	domainName := "test.scaleway-terraform.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayTemDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_tem_domain" "test" {
						name = "%s"
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayTemDomainExists(tt, "data.scaleway_tem_domain.test"),
					resource.TestCheckResourceAttr("data.scaleway_tem_domain.test", "name", domainName),
					resource.TestCheckResourceAttrSet("scaleway_tem_domain.test", "reputation.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_tem_domain.test", "reputation.0.score"),
					resource.TestCheckResourceAttrSet("scaleway_tem_domain.test", "reputation.0.scored_at"),
					resource.TestCheckResourceAttrSet("scaleway_tem_domain.test", "reputation.0.previous_score"),
					resource.TestCheckResourceAttrSet("scaleway_tem_domain.test", "reputation.0.previous_scored_at"),
				),
			},
		},
	})
}
