package domain_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/domain"
)

func TestAccDomainRegistration_SingleDomainWithUpdate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	singleDomain := "test-single-updates34" + ".com" // à adapter

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "scaleway_domain_domains_registration" "test" {
                      domain_names = [ "%s"]
                      duration_in_years = 1

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
                `, singleDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.test", "domain_names.0", singleDomain),
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.test", "duration_in_years", "1"),
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.test", "owner_contact.0.firstname", "John"),
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.test", "auto_renew", "false"),
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.test", "dnssec", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
			           resource "scaleway_domain_domains_registration" "test" {
			             domain_names = [ "%s"]
			             duration_in_years = 1

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

					     auto_renew = true

					     dnssec = true
			           }
			       `, singleDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.test", "auto_renew", "true"),
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.test", "dnssec", "true")),
			},
		},
	})
}

func TestAccDomainRegistration_MultipleDomainsNoUpdate(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainName1 := "test-multiple-1.com"
	domainName2 := "test-multiple-2.com"
	domainName3 := "test-multiple-3.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "scaleway_domain_domains_registration" "multi" {
                      domain_names = ["%s","%s","%s"]

                      duration_in_years = 1

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
                `, domainName1, domainName2, domainName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.multi", "domain_names.0", domainName1),
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.multi", "domain_names.0", domainName2),
					resource.TestCheckResourceAttr("scaleway_domain_domains_registration.multi", "domain_names.0", domainName3),
				),
			},
		},
	})
}

func testAccCheckDomainDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_domain_domains_registration" {
				continue
			}

			registrarAPI := domain.NewRegistrarDomainAPI(tt.Meta)

			domainNames, err := domain.ExtractDomainsFromTaskID(nil, rs.Primary.ID, registrarAPI)
			if err != nil {
				return nil
			}

			for _, domainName := range domainNames {
				domainResp, getErr := registrarAPI.GetDomain(&domainSDK.RegistrarAPIGetDomainRequest{
					Domain: domainName,
				})
				if getErr != nil {
					if httperrors.Is404(getErr) {
						continue
					}
					return fmt.Errorf("failed to get domain details for %s: %w", domainName, getErr)
				}

				if domainResp.AutoRenewStatus != domainSDK.DomainFeatureStatusDisabled {
					return fmt.Errorf(
						"domain %s still exists, and auto-renew is not disabled (current: %s)",
						domainName,
						domainResp.AutoRenewStatus,
					)
				}
			}
		}
		return nil
	}
}
