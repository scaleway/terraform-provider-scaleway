package tem_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	temSDK "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/tem"
)

const domainNameValidation = "scaleway-terraform.com"

func TestAccDomain_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainName := "terraform-rs.test.local"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_tem_domain cr01 {
						name       = "%s"
						accept_tos = true
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					isDomainPresent(tt, "scaleway_tem_domain.cr01"),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "name", domainName),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "dmarc_config", "v=DMARC1; p=none"),
					resource.TestMatchResourceAttr("scaleway_tem_domain.cr01", "dmarc_name", regexp.MustCompile(`^_dmarc\.terraform-rs\.test\.local\.$`)),
					resource.TestMatchResourceAttr("scaleway_tem_domain.cr01", "dkim_name", regexp.MustCompile(`^[a-f0-9-]+\._domainkey\.terraform-rs\.test\.local\.$`)),
					resource.TestMatchResourceAttr("scaleway_tem_domain.cr01", "spf_value", regexp.MustCompile(`^v=spf1 include:.+ -all$`)),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "mx_config", "10 blackhole.tem.scaleway.com."),
					acctest.CheckResourceAttrUUID("scaleway_tem_domain.cr01", "id"),
				),
			},
		},
	})
}

func TestAccDomain_Tos(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainName := "terraform-rs.test.local"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_tem_domain cr01 {
						name       = "%s"
						accept_tos = false
					}
				`, domainName),
				ExpectError: regexp.MustCompile("you must accept"),
			},
		},
	})
}

// TestAccDomain_Autoconfig is now covered by TestAccTEM_Complete step 1
// TestAccDomain_AutoconfigUpdate was removed: updating autoconfig from false to true takes >10 minutes (API schedules next check in 10 min), causing timeout

func isDomainPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := tem.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDomain(&temSDK.GetDomainRequest{
			DomainID: id,
			Region:   region,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func isDomainDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_tem_domain" {
				continue
			}

			api, region, id, err := tem.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.RevokeDomain(&temSDK.RevokeDomainRequest{
				Region:   region,
				DomainID: id,
			}, scw.WithContext(context.Background()))
			if err != nil {
				return err
			}

			_, err = tem.WaitForDomain(context.Background(), api, region, id, tem.DefaultDomainTimeout)
			if err != nil && !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
