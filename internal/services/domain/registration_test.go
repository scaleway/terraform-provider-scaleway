package domain_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/env"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/domain"
)

func TestAccDomainRegistration_SingleDomainWithUpdate(t *testing.T) {
	if shouldBeSkipped() {
		t.Skip("Test skipped: must be run in a staging environment")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	singleDomain := "test-single-updates51" + ".com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "scaleway_domain_registration" "test" {
                      project_id         = "%s"
                      domain_names       = [ "%s"]
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
                `, testAccDomainRegistrationProjectID, singleDomain),
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
			             project_id         = "%s"
			             domain_names       = [ "%s"]
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

					     auto_renew = true

					     dnssec = true
			           }
			       `, testAccDomainRegistrationProjectID, singleDomain),
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

	domainName1 := "test-multiple-1243.com"
	domainName2 := "test-multiple-1253.com"
	domainName3 := "test-multiple-1263.com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
                    resource "scaleway_domain_registration" "multi" {
                      project_id         = "%s"
                      domain_names       = ["%s","%s","%s"]
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
                `, testAccDomainRegistrationProjectID, domainName1, domainName2, domainName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_registration.multi", "domain_names.0", domainName1),
					resource.TestCheckResourceAttr("scaleway_domain_registration.multi", "domain_names.1", domainName2),
					resource.TestCheckResourceAttr("scaleway_domain_registration.multi", "domain_names.2", domainName3),
				),
			},
			{
				Config: fmt.Sprintf(`
                    resource "scaleway_domain_registration" "multi" {
                      project_id         = "%s"
                      domain_names       = ["%s", "%s", "%s"]
                      duration_in_years   = 1

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
                `, testAccDomainRegistrationProjectID, domainName1, domainName2, domainName3),
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

			domainNames := extractDomainNamesFromResourceState(rs)
			if len(domainNames) == 0 {
				var err error

				domainNames, err = domain.ExtractDomainsFromTaskID(context.TODO(), rs.Primary.ID, registrarAPI)
				if err != nil {
					return err
				}
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

func TestAccDomainRegistration_ByTaskID(t *testing.T) {
	if shouldBeSkipped() {
		t.Skip("Test skipped: must be run in a staging environment")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	singleDomain := "test-import-by-task-id2.com"
	taskID := "9fb6c780-6d10-44f8-8515-977b3765496a"

	config := fmt.Sprintf(`
		resource "scaleway_domain_registration" "test" {
		  project_id        = "%s"
		  domain_names      = ["%s"]
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
	`, testAccDomainRegistrationProjectID, singleDomain)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "domain_names.0", singleDomain),
					resource.TestCheckResourceAttrSet("scaleway_domain_registration.test", "task_id"),
				),
			},
			{
				ResourceName:            "scaleway_domain_registration.test",
				ImportState:             true,
				ImportStateId:           testAccDomainRegistrationProjectID + "/" + taskID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"duration_in_years"},
			},
		},
	})
}

func TestAccDomainRegistration_ByDomainName(t *testing.T) {
	if shouldBeSkipped() {
		t.Skip("Test skipped: must be run in a staging environment")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	singleDomain := "test-import-by-domain-name3.com"

	config := fmt.Sprintf(`
		resource "scaleway_domain_registration" "test" {
		  project_id        = "%s"
		  domain_names      = ["%s"]
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
	`, testAccDomainRegistrationProjectID, singleDomain)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_registration.test", "domain_names.0", singleDomain),
				),
			},
			{
				ResourceName:            "scaleway_domain_registration.test",
				ImportState:             true,
				ImportStateId:           testAccDomainRegistrationProjectID + "/" + singleDomain,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"duration_in_years", "task_id"},
			},
		},
	})
}

func extractDomainNamesFromResourceState(rs *terraform.ResourceState) []string {
	var domainNames []string

	for key, value := range rs.Primary.Attributes {
		if strings.HasPrefix(key, "domain_names.") && value != "" {
			domainNames = append(domainNames, value)
		}
	}

	sort.Strings(domainNames)

	return domainNames
}

// shouldBeSkipped returns true when cassette recording is active but TF_ACC_DOMAIN_REGISTRATION is
// not set, preventing unintended domain purchases in staging.
func shouldBeSkipped() bool {
	if os.Getenv(env.UpdateCassettes) == "false" {
		return false
	}

	return os.Getenv(env.AccDomainRegistration) != "true"
}
