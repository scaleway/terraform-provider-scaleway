package domain_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/domain"
)

func TestAccDomainRegistration_SingleDomainWithUpdate(t *testing.T) {
	if shouldBeSkipped() {
		t.Skip("Test skipped: must be run in a staging environment")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	singleDomain := "test-single-updates37" + ".com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "scaleway_domain_registration" "test" {
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
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "domain_names.0", singleDomain),
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "duration_in_years", "1"),
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "owner_contact.0.firstname", "John"),
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "auto_renew", "false"),
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "dnssec", "false"),
					resource.TestCheckResourceAttrSet("scaleway_domain_registration.test", "task_id"),
				),
			},
			{
				Config: fmt.Sprintf(`
			           resource "scaleway_domain_registration" "test" {
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
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "auto_renew", "true"),
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "dnssec", "true"),
					resource.TestCheckResourceAttrSet("scaleway_domain_registration.test", "task_id"),
				),
			},
		},
	})
}

func TestAccDomainRegistration_MultipleDomainsUpdate(t *testing.T) {
	if shouldBeSkipped() {
		t.Skip("Test skipped: must be run in a staging environment")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainName1 := "test-multiple-121.com"
	domainName2 := "test-multiple-122.com"
	domainName3 := "test-multiple-123.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "scaleway_domain_registration" "multi" {
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
					resource.TestCheckResourceAttr("scaleway_domain_registration.multi", "domain_names.0", domainName1),
					resource.TestCheckResourceAttr("scaleway_domain_registration.multi", "domain_names.1", domainName2),
					resource.TestCheckResourceAttr("scaleway_domain_registration.multi", "domain_names.2", domainName3),
				),
			},
			{
				Config: fmt.Sprintf(`
                    resource "scaleway_domain_registration" "multi" {
                      domain_names = ["%s", "%s", "%s"]
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
                      dnssec     = true
                    }
                `, domainName1, domainName2, domainName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_registration.multi", "auto_renew", "true"),
					resource.TestCheckResourceAttr("scaleway_domain_registration.multi", "dnssec", "true"),
					testAccCheckDomainStatus(tt, domainSDK.DomainFeatureStatusEnabled.String(), domainSDK.DomainFeatureStatusEnabled.String()),
				),
			},
		},
	})
}

func testAccCheckDomainStatus(tt *acctest.TestTools, expectedAutoRenew, expectedDNSSEC string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_domain_registration" {
				continue
			}

			registrarAPI := domain.NewRegistrarDomainAPI(tt.Meta)

			domainNames, err := domain.ExtractDomainsFromTaskID(context.TODO(), rs.Primary.ID, registrarAPI)
			if err != nil {
				return fmt.Errorf("error extracting domains: %w", err)
			}

			for _, domainName := range domainNames {
				domainResp, getErr := registrarAPI.GetDomain(&domainSDK.RegistrarAPIGetDomainRequest{
					Domain: domainName,
				})

				if getErr != nil {
					return fmt.Errorf("failed to get details for domain %s: %w", domainName, getErr)
				}

				if domainResp.AutoRenewStatus.String() != expectedAutoRenew {
					return fmt.Errorf("domain %s has auto_renew status %s, expected %s", domainName, domainResp.AutoRenewStatus, expectedAutoRenew)
				}

				if domainResp.Dnssec.Status.String() != expectedDNSSEC {
					return fmt.Errorf("domain %s has dnssec status %s, expected %s", domainName, domainResp.Dnssec.Status.String(), expectedDNSSEC)
				}
			}
		}

		return nil
	}
}

func testAccCheckDomainDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_domain_registration" {
				continue
			}

			registrarAPI := domain.NewRegistrarDomainAPI(tt.Meta)

			domainNames, err := domain.ExtractDomainsFromTaskID(context.TODO(), rs.Primary.ID, registrarAPI)
			if err != nil {
				return err
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

// shouldBeSkipped determines whether the test should be skipped based on the execution environment.
//
// Running domain registration tests in a production environment is not feasible because domains
// are permanently reserved and billed upon registration. To safely execute these tests, a controlled
// test environment must be used.
//
// Test execution is controlled by the following environment variables:
//
// - `TF_UPDATE_CASSETTES`: If set to "true", additional restrictions apply based on `TF_ACC_DOMAIN_REGISTRATION`.
// - `TF_ACC_DOMAIN_REGISTRATION`: Must be set to "true" when `TF_UPDATE_CASSETTES=true` to allow domain registration tests.
//
// Example usage:
//
//	export TF_ACC_DOMAIN_REGISTRATION=true
//
// If `TF_UPDATE_CASSETTES=false`, the test **is always executed**.
// If `TF_UPDATE_CASSETTES=true`, the test is **only executed if `TF_ACC_DOMAIN_REGISTRATION=true`**.
// Otherwise, the test is skipped to prevent unintended domain reservations.
func shouldBeSkipped() bool {
	if os.Getenv("TF_UPDATE_CASSETTES") == "false" {
		return false
	}

	return os.Getenv("TF_ACC_DOMAIN_REGISTRATION") != "true"
}
