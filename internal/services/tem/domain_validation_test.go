package tem_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

const domainNameValidation = "scaleway-terraform.com"

func TestAccDomainValidation_Validation(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	subDomainName := "validation-validation"

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

					resource scaleway_tem_domain cr01 {
						name       = scaleway_domain_zone.test.id
						accept_tos = true
						autoconfig = true
					}

					resource scaleway_tem_domain_validation valid {
  						domain_id = scaleway_tem_domain.cr01.id
						timeout = 3600
					}
				`, domainNameValidation, subDomainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_tem_domain_validation.valid", "validated", "true"),
				),
			},
		},
	})
}

func TestAccDomainValidation_TimeoutError(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	subDomainName := "validation-timeout"

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

                    resource scaleway_tem_domain cr01 {
                        name       = scaleway_domain_zone.test.id
                        accept_tos = true
                    }

                    resource scaleway_tem_domain_validation valid {
                        domain_id = scaleway_tem_domain.cr01.id
                        timeout   = 1
                    }
                `, domainNameValidation, subDomainName),
				ExpectError: regexp.MustCompile("(?i)domain validation did not complete"),
			},
		},
	})
}
