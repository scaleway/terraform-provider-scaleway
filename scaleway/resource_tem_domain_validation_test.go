package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const domainName = "scaleway-terraform.com"

func init() {
	resource.AddTestSweepers("scaleway_tem_domain_validation", &resource.Sweeper{
		Name: "scaleway_tem_domain_validation",
		F:    testSweepTemDomain,
	})
}

func TestAccScalewayTemDomainValidation_NoValidation(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayTemDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_tem_domain cr01 {
						name       = "%s"
						accept_tos = true
					}

					resource scaleway_tem_domain_validation valid {
  						domain_id = scaleway_tem_domain.cr01.id
  						region = scaleway_tem_domain.cr01.region
						timeout = 1
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_tem_domain_validation.valid", "validated", "false"),
				),
			},
		},
	})
}

func TestAccScalewayTemDomainValidation_Validation(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayTemDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_tem_domain cr01 {
						name       = "%s"
						accept_tos = true
					}
					resource "scaleway_domain_record" "spf" {
  						dns_zone = "%s"
  						type     = "TXT"
						data     = "v=spf1 ${scaleway_tem_domain.cr01.spf_config} -all"
					}
					resource "scaleway_domain_record" "dkim" {
  						dns_zone = "%s"
  						name     = "${scaleway_tem_domain.cr01.project_id}._domainkey"
  						type     = "TXT"
  						data     = scaleway_tem_domain.cr01.dkim_config
					}
					resource "scaleway_domain_record" "mx" {
  						dns_zone = "%s"
  						type     = "MX"
  						data     = "."
					}
					resource scaleway_tem_domain_validation valid {
  						domain_id = scaleway_tem_domain.cr01.id
  						region = scaleway_tem_domain.cr01.region
						timeout = 3600
					}
				`, domainName, domainName, domainName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_tem_domain_validation.valid", "validated", "true"),
				),
			},
		},
	})
}
