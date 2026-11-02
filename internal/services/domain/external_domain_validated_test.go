package domain_test

import (
	"errors"
	"fmt"
	"log"
	"os"
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

// TF_TEST_EXTERNAL_DOMAIN is the real FQDN used when recording (parent zone on
// Route53, not registered at Scaleway). Replay uses the placeholder default,
// which must match the committed cassette after scrub.
const (
	testExternalDomainEnv     = "TF_TEST_EXTERNAL_DOMAIN"
	testExternalDomainDefault = "tf-tests-ext.example.com"
)

func TestAccDomainExternalDomainValidated_Basic(t *testing.T) {
	domainName, subdomain, dnsZone := externalDomainTestParts(t)

	requireAWSCredsForExternalDomain(t, dnsZone)

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	if *acctest.UpdateCassettes {
		recordedDomain := domainName

		t.Cleanup(func() {
			_ = acctest.AnonymizeCassetteForTest(t, "")
			scrubExternalDomainInCassette(t, recordedDomain, testExternalDomainDefault)
		})
	}

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
					testAccPublishExternalDomainChallengeTXT(t, "scaleway_domain_external.example", subdomain, dnsZone),
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

func externalDomainTestParts(t *testing.T) (domainName, subdomain, dnsZone string) {
	t.Helper()

	domainName = os.Getenv(testExternalDomainEnv)
	if domainName == "" {
		domainName = testExternalDomainDefault
	}

	parts := strings.SplitN(domainName, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		t.Fatalf("%s=%q must be a FQDN with at least two labels", testExternalDomainEnv, domainName)
	}

	return domainName, parts[0], parts[1]
}

// scrubExternalDomainInCassette replaces the recorded FQDN with the cassette
// placeholder so real zones are not committed.
func scrubExternalDomainInCassette(t *testing.T, from, to string) {
	t.Helper()

	if from == "" || from == to {
		return
	}

	pkgFolder, err := os.Getwd()
	if err != nil {
		t.Fatalf("scrub cassette: getwd: %v", err)
	}

	path := acctest.BuildCassetteName(t.Name(), pkgFolder, ".cassette") + ".yaml"

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("scrub cassette: read %s: %v", path, err)
	}

	updated := strings.ReplaceAll(string(data), from, to)
	if updated == string(data) {
		t.Fatalf("scrub cassette: %q not found in %s", from, path)
	}

	if err := os.WriteFile(path, []byte(updated), 0o600); err != nil {
		t.Fatalf("scrub cassette: write %s: %v", path, err)
	}
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
