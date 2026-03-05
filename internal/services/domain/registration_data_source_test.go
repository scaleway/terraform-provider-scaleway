package domain_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceDomainRegistration_Basic(t *testing.T) {
	if shouldBeSkipped() {
		t.Skip("Test skipped: must be run in a staging environment")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	// Fixed domain to match cassette (VCR requires exact body match)
	domainName := "test-ds-reg-2-942430570701024891.com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_registration" "test" {
						domain_names       = ["%s"]
						duration_in_years  = 1

						owner_contact {
							firstname                   = "John"
							lastname                    = "DOE"
							email                       = "john.doe@example.com"
							phone_number                = "+1.23456789"
							address_line_1              = "123 Main Street"
							city                        = "Paris"
							zip                         = "75001"
							country                     = "FR"
							legal_form                  = "individual"
							vat_identification_code     = "FR12345678901"
							company_identification_code = "123456789"
						}
					}

					data "scaleway_domain_registration" "test" {
						domain_name = scaleway_domain_registration.test.domain_names[0]
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "domain_names.0", domainName),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_registration.test", "task_id",
						"scaleway_domain_registration.test", "task_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_registration.test", "project_id",
						"scaleway_domain_registration.test", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_registration.test", "domain_names",
						"scaleway_domain_registration.test", "domain_names",
					),
				),
			},
		},
	})
}

func TestAccDataSourceDomainRegistration_WithProjectID(t *testing.T) {
	if shouldBeSkipped() {
		t.Skip("Test skipped: must be run in a staging environment")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	// Fixed domain to match cassette (VCR requires exact body match)
	domainName := "test-ds-reg-project--576332352888738072.com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_domain_registration" "test" {
						domain_names       = ["%s"]
						duration_in_years  = 1

						owner_contact {
							firstname                   = "John"
							lastname                    = "DOE"
							email                       = "john.doe@example.com"
							phone_number                = "+1.23456789"
							address_line_1              = "123 Main Street"
							city                        = "Paris"
							zip                         = "75001"
							country                     = "FR"
							legal_form                  = "individual"
							vat_identification_code     = "FR12345678901"
							company_identification_code = "123456789"
						}
					}

					data "scaleway_domain_registration" "test" {
						domain_name = scaleway_domain_registration.test.domain_names[0]
						project_id  = scaleway_domain_registration.test.project_id
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "domain_names.0", domainName),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_registration.test", "task_id",
						"scaleway_domain_registration.test", "task_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_registration.test", "project_id",
						"scaleway_domain_registration.test", "project_id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_registration.test", "domain_names",
						"scaleway_domain_registration.test", "domain_names",
					),
				),
			},
		},
	})
}
