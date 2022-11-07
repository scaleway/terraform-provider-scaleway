package scaleway

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_tem_domain", &resource.Sweeper{
		Name: "scaleway_tem_domain",
		F:    testSweepTemDomain,
	})
}

func testSweepTemDomain(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar, scw.RegionNlAms}, func(scwClient *scw.Client, region scw.Region) error {
		temAPI := tem.NewAPI(scwClient)
		l.Debugf("sweeper: revoking the tem domains in (%s)", region)

		listDomains, err := temAPI.ListDomains(&tem.ListDomainsRequest{Region: region}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing domains in (%s) in sweeper: %s", region, err)
		}

		for _, ns := range listDomains.Domains {
			_, err := temAPI.RevokeDomain(&tem.RevokeDomainRequest{
				DomainID: ns.ID,
				Region:   region,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error revoking domain in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayTemDomain_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	domainName := "terraform-rs.test.local"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayTemDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource scaleway_tem_domain cr01 {
						name = "%s"
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayTemDomainExists(tt, "scaleway_tem_domain.cr01"),
					resource.TestCheckResourceAttr("scaleway_tem_domain.cr01", "name", domainName),
					testCheckResourceAttrUUID("scaleway_tem_domain.cr01", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayTemDomainExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := temAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDomain(&tem.GetDomainRequest{
			DomainID: id,
			Region:   region,
		})

		if err != nil {
			return fmt.Errorf("can't get domain: %w", err)
		}

		return nil
	}
}

func testAccCheckScalewayTemDomainDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_tem_domain" {
				continue
			}

			api, region, id, err := temAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.RevokeDomain(&tem.RevokeDomainRequest{
				DomainID: id,
				Region:   region,
			})
			if err != nil {
				return fmt.Errorf("can't revoke domain: %w", err)
			}

			_, err = api.RevokeDomain(&tem.RevokeDomainRequest{
				Region:   region,
				DomainID: id,
			}, scw.WithContext(context.Background()))
			if err != nil {
				return fmt.Errorf("can't revoke domain: %w", err)
			}

			_, err = waitForTemDomain(context.Background(), api, region, id, defaultTemDomainTimeout)
			if err != nil && !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
