package tem_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

const domainNameValidation = "scaleway-terraform.com"

func TestAccDomainValidation_NoValidation(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isDomainDestroyed(tt),
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
				`, domainNameValidation),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_tem_domain_validation.valid", "validated", "false"),
				),
			},
		},
	})
}

func TestAccDomainValidation_Validation(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isDomainDestroyed(tt),
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

					resource "scaleway_domain_record" "dmarc" {
						dns_zone = "%s"
  						name     = scaleway_tem_domain.cr01.dmarc_name
  						type     = "TXT"
  						data     = scaleway_tem_domain.cr01.dmarc_config
					}

					resource scaleway_tem_domain_validation valid {
  						domain_id = scaleway_tem_domain.cr01.id
  						region = scaleway_tem_domain.cr01.region
						timeout = 3600
					}
				`, domainNameValidation, domainNameValidation, domainNameValidation, domainNameValidation, domainNameValidation),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_tem_domain_validation.valid", "validated", "true"),
				),
			},
		},
	})
}
