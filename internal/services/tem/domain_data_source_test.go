package tem_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceDomain_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainName := "terraform-ds.test.local"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isDomainDestroyed(tt),
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

func TestAccDataSourceDomain_Reputation(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	subDomainName := "test-reputation"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_zone" "test" {
						domain    = "%s"
						subdomain = "%s"
					}

					resource "scaleway_tem_domain" "main" {
						name       = scaleway_domain_zone.test.id
						accept_tos = true
						autoconfig = true
					}

					resource "scaleway_tem_domain_validation" "valid" {
						domain_id = scaleway_tem_domain.main.id
						region    = scaleway_tem_domain.main.region
						timeout   = 3600
					}

					data "scaleway_tem_domain" "test" {
						name = scaleway_tem_domain.main.name
					}
				`, domainNameValidation, subDomainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_tem_domain_validation.valid", "validated", "true"),
					isDomainPresent(tt, "data.scaleway_tem_domain.test"),
					resource.TestCheckResourceAttr("data.scaleway_tem_domain.test", "name", subDomainName+"."+domainNameValidation),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_domain.test", "reputation.0.status"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_domain.test", "reputation.0.score"),
					resource.TestCheckResourceAttrSet("data.scaleway_tem_domain.test", "reputation.0.scored_at"),
				),
			},
		},
	})
}
