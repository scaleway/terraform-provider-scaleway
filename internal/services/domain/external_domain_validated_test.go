package domain_test

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/domain"
)

// Placeholder FQDN scrubbed into the committed cassette (replay only).
const testExternalDomainDefault = "tf-tests-ext.example.com"

func TestAccDomainExternalDomainValidated_Basic(t *testing.T) {
	// Recording needs a real external zone + manual TXT challenge publish.
	// Keep the committed cassette for CI replay; no AWS/DNS automation.
	if *acctest.UpdateCassettes {
		t.Skip("external domain validation recording requires a manual TXT challenge; use the committed cassette for replay")
	}

	domainName := testExternalDomainDefault

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	log.Printf("Testing external domain validation for domain: %s", domainName)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckExternalDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainExternalDomainConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalDomainExists(tt, "scaleway_domain_external.example"),
					resource.TestCheckResourceAttr("scaleway_domain_external.example", "domain", domainName),
					resource.TestCheckResourceAttrSet("scaleway_domain_external.example", "validation_token"),
				),
			},
			{
				Config: testAccDomainExternalDomainValidatedConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalDomainExists(tt, "scaleway_domain_external.example"),
					resource.TestCheckResourceAttr("scaleway_domain_external_validated.example", "domain", domainName),
					resource.TestCheckResourceAttrSet("scaleway_domain_external_validated.example", "id"),
					resource.TestCheckResourceAttr("scaleway_domain_external_validated.example", "validated", "true"),
					resource.TestCheckResourceAttrSet("scaleway_domain_external_validated.example", "ns_servers.#"),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_external.example", "domain",
						"scaleway_domain_external.example", "domain",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_domain_external.example", "status",
						"scaleway_domain_external.example", "status",
					),
				),
			},
			{
				ResourceName:      "scaleway_domain_external.example",
				ImportState:       true,
				ImportStateVerify: true,
				// status / token / ns_servers move during the validation lifecycle
				ImportStateVerifyIgnore: []string{"validation_token", "status", "ns_servers"},
			},
			{
				ResourceName:      "scaleway_domain_external_validated.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckExternalDomainExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		registrarAPI := domain.NewRegistrarDomainAPI(tt.Meta)

		_, err := registrarAPI.GetDomain(&domainSDK.RegistrarAPIGetDomainRequest{
			Domain: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckExternalDomainDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_domain_external" {
				continue
			}

			registrarAPI := domain.NewRegistrarDomainAPI(tt.Meta)

			// DeleteExternalDomain can be eventually consistent.
			var lastErr error

			for range 30 {
				_, err := registrarAPI.GetDomain(&domainSDK.RegistrarAPIGetDomainRequest{
					Domain: rs.Primary.ID,
				})
				if err != nil {
					if httperrors.Is404(err) || (httperrors.Is403(err) && strings.Contains(err.Error(), "not found")) {
						lastErr = nil

						break
					}

					return err
				}

				lastErr = fmt.Errorf("external domain %s still exists", rs.Primary.ID)

				time.Sleep(3 * time.Second) // lintignore:R018
			}

			if lastErr != nil {
				return lastErr
			}
		}

		return nil
	}
}

func testAccDomainExternalDomainConfig(domainName string) string {
	return fmt.Sprintf(`
resource "scaleway_domain_external" "example" {
  domain = "%s"
}
`, domainName)
}

func testAccDomainExternalDomainValidatedConfig(domainName string) string {
	return fmt.Sprintf(`
resource "scaleway_domain_external" "example" {
  domain = "%s"
}

resource "scaleway_domain_external_validated" "example" {
  domain = scaleway_domain_external.example.domain
}

data "scaleway_domain_external" "example" {
  domain = scaleway_domain_external.example.domain
}
`, domainName)
}
