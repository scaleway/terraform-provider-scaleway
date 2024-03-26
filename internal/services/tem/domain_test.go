package tem_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	temSDK "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/tem"
)

func init() {
	resource.AddTestSweepers("scaleway_tem_domain", &resource.Sweeper{
		Name: "scaleway_tem_domain",
		F:    testSweepDomain,
	})
}

func testSweepDomain(_ string) error {
	return acctest.SweepRegions([]scw.Region{scw.RegionFrPar, scw.RegionNlAms}, func(scwClient *scw.Client, region scw.Region) error {
		temAPI := temSDK.NewAPI(scwClient)
		logging.L.Debugf("sweeper: revoking the tem domains in (%s)", region)

		listDomains, err := temAPI.ListDomains(&temSDK.ListDomainsRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing domains in (%s) in sweeper: %s", region, err)
		}

		for _, ns := range listDomains.Domains {
			if ns.Name == "test.scaleway-terraform.com" {
				logging.L.Debugf("sweeper: skipping deletion of domain %s", ns.Name)
				continue
			}
			_, err := temAPI.RevokeDomain(&temSDK.RevokeDomainRequest{
				DomainID: ns.ID,
				Region:   region,
			})
			if err != nil {
				logging.L.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error revoking domain in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccDomain_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainName := "terraform-rs.test.local"

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
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					isDomainPresent(tt, "scaleway_tem_domain.cr01"),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "name", domainName),
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      isDomainDestroyed(tt),
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
