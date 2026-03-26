package domain_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/domain"
)

func TestAccDomainExternalDomainValidated_Basic(t *testing.T) {
	if acctest.TestDomain == "" {
		t.Skip("Test skipped: SCW_TEST_DOMAIN must be set")
	}
	rootZoneProfile := os.Getenv("SCW_TEST_ROOT_ZONE_PROFILE")
	if rootZoneProfile == "" {
		t.Skip("Test skipped: SCW_TEST_ROOT_ZONE_PROFILE must be set")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	subdomain := "tf-acc-ext-validated"
	domainName := fmt.Sprintf("%s.%s", subdomain, acctest.TestDomain)
	log.Printf("Testing external domain validation for domain: %s", domainName)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             testAccCheckExternalDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainExternalDomainValidatedConfigBasic(domainName, subdomain, rootZoneProfile),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_domain_external_domain.example", "domain", domainName),
					resource.TestCheckResourceAttrSet("scaleway_domain_external_domain.example", "validation_token"),
					resource.TestCheckResourceAttr("scaleway_domain_external_domain_validated.example", "domain", domainName),
					resource.TestCheckResourceAttrSet("scaleway_domain_external_domain_validated.example", "id"),
				),
			},
		},
	})
}

func testAccCheckExternalDomainDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_domain_external_domain" {
				continue
			}

			registrarAPI := domain.NewRegistrarDomainAPI(tt.Meta)

			_, err := registrarAPI.GetDomain(&domainSDK.RegistrarAPIGetDomainRequest{
				Domain: rs.Primary.ID,
			})
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}
				return err
			}

			return fmt.Errorf("external domain %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainExternalDomainValidatedConfigBasic(domainName, subdomain string,rootZoneProfile string) string {
	return fmt.Sprintf(`
resource "scaleway_domain_external_domain" "example" {
  domain = "%s"
}

resource "scaleway_domain_record" "validation" {
  dns_zone   = "%s"
  name       = "_scaleway-challenge.%s"
  type       = "TXT"
  data       = scaleway_domain_external_domain.example.validation_token
 provider    = scaleway.alt
}

resource "scaleway_domain_external_domain_validated" "example" {
  domain = scaleway_domain_external_domain.example.domain
}

provider "scaleway" {
  profile="%s"
  alias="alt"
}


`, domainName, acctest.TestDomain, subdomain, rootZoneProfile)
}
