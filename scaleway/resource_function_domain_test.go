package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_function_domain", &resource.Sweeper{
		Name: "scaleway_function_domain",
		F:    testSweepFunctionDomain,
	})
}

func testSweepFunctionDomain(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		functionAPI := function.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the function domains in (%s)", region)
		domains, err := functionAPI.ListDomains(
			&function.ListDomainsRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing function domains in (%s) in sweeper: %s", region, err)
		}

		for _, domain := range domains.Domains {
			_, err := functionAPI.DeleteDomain(&function.DeleteDomainRequest{
				DomainID: domain.ID,
				Region:   region,
			})
			if err != nil {
				l.Debugf("sweeper: error (%s)", err)

				return fmt.Errorf("error deleting domain in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayFunctionDomain_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"dns": {
				Source: "hashicorp/dns",
			},
		},
		CheckDestroy: testAccCheckScalewayFunctionDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_function_namespace main {}

					resource scaleway_function main {
						namespace_id = scaleway_function_namespace.main.id
						runtime = "go118"
						privacy = "private"
						handler = "Handle"
						zip_file = "testfixture/gofunction.zip"
						deploy = true
					}

					provider "dns" {}
					data "dns_a_record_set" "main" {
					  host = scaleway_function.main.domain_name
					}

					resource "scaleway_function_domain" "main" {
					  function_id = scaleway_function.main.id
					  hostname    = "${data.dns_a_record_set.main.addrs[0]}.nip.io"
					
					  depends_on = [
						scaleway_function.main,
					  ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionDomainExists(tt, "scaleway_function_domain.main"),
				),
			},
			{
				Config: `
					resource scaleway_function_namespace main {}

					resource scaleway_function main {
						namespace_id = scaleway_function_namespace.main.id
						runtime = "go118"
						privacy = "private"
						handler = "Handle"
						zip_file = "testfixture/gofunction.zip"
						deploy = true
					}
				`,
			},
		},
	})
}

func testAccCheckScalewayFunctionDomainExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := functionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDomain(&function.GetDomainRequest{
			DomainID: id,
			Region:   region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayFunctionDomainDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_function_domain" {
				continue
			}

			api, region, id, err := functionAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteDomain(&function.DeleteDomainRequest{
				DomainID: id,
				Region:   region,
			})

			if err == nil {
				return fmt.Errorf("function domain (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
