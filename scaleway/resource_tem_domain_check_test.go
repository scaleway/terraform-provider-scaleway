package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayTemDomainCheck_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	domainName := "test-tem-check." + testDomain

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayTemDomainDestroy(tt),
		Steps: []resource.TestStep{
			// {
			// 	Config: fmt.Sprintf(`
			// 		data "scaleway_account_project" "main" {}

			// 		resource "scaleway_tem_domain" "main" {
			// 			name       = "%s"
			// 			accept_tos = true
			// 		}

			// 		resource "scaleway_domain_record" "dkim" {
			// 			dns_zone = scaleway_tem_domain.main.name
			// 			name     = "${data.scaleway_account_project.main.id}._domainkey"
			// 			type     = "TXT"
			// 			data     = scaleway_tem_domain.main.dkim_config
			// 			ttl      = 3600
			// 		}

			// 		resource "scaleway_tem_domain_check" "main" {
			// 			triggers = {
			// 				dkim = scaleway_domain_record.dkim.id
			// 			}

			// 			domain_id = scaleway_tem_domain.main.id
			// 		}
			// 	`, domainName),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		testAccCheckScalewayTemDomainExists(tt, "scaleway_tem_domain.main"),
			// 		testAccCheckScalewayDomainRecordExists(tt, "scaleway_domain_record.dkim"),
			// 		resource.TestCheckResourceAttr("scaleway_tem_domain_check.main", "is_ready", "false"),
			// 	),
			// },
			{
				Config: fmt.Sprintf(`
					data "scaleway_account_project" "main" {}
					
					resource "scaleway_tem_domain" "main" {
						name       = "%s"
						accept_tos = true
					}
					
					resource "scaleway_domain_record" "dkim" {
						dns_zone = scaleway_tem_domain.main.name
						name     = "${data.scaleway_account_project.main.id}._domainkey"
						type     = "TXT"
						data     = scaleway_tem_domain.main.dkim_config
						ttl      = 3600
					}
					
					resource "scaleway_domain_record" "spf" {
						dns_zone = scaleway_tem_domain.main.name
						name     = ""
						type     = "TXT"
						data     = "v=spf1 ${scaleway_tem_domain.main.spf_config} -all"
						ttl      = 3600
					}
					
					resource "scaleway_domain_record" "mx" {
						dns_zone = scaleway_tem_domain.main.name
						name     = ""
						type     = "MX"
						data     = "."
						ttl      = 3600
						priority = 0
					}

					resource "scaleway_tem_domain_check" "main" {
						triggers = {
							dkim = scaleway_domain_record.dkim.id
							spf  = scaleway_domain_record.spf.id
							mx   = scaleway_domain_record.mx.id
						}

						domain_id = scaleway_tem_domain.main.id
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayTemDomainExists(tt, "scaleway_tem_domain.main"),
					testAccCheckScalewayDomainRecordExists(tt, "scaleway_domain_record.dkim"),
					testAccCheckScalewayDomainRecordExists(tt, "scaleway_domain_record.spf"),
					testAccCheckScalewayDomainRecordExists(tt, "scaleway_domain_record.mx"),
					resource.TestCheckResourceAttr("scaleway_tem_domain_check.main", "is_ready", "true"),
				),
			},
		},
	})
}
