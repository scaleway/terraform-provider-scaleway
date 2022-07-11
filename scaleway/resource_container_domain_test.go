package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_container_domain", &resource.Sweeper{
		Name: "scaleway_container_domain",
		F:    testSweepContainerDomain,
	})
}

func testSweepContainerDomain(_ string) error {
	return sweepRegions([]scw.Region{scw.RegionFrPar}, func(scwClient *scw.Client, region scw.Region) error {
		containerAPI := container.NewAPI(scwClient)
		l.Debugf("sweeper: destroying the container domains in (%s)", region)
		domains, err := containerAPI.ListDomains(
			&container.ListDomainsRequest{
				Region: region,
			}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing container domains in (%s) in sweeper: %s", region, err)
		}

		for _, domain := range domains.Domains {
			_, err := containerAPI.DeleteDomain(&container.DeleteDomainRequest{
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

func TestAccScalewayContainerDomain_Basic(t *testing.T) {
	if !*UpdateCassettes {
		t.Skip("Skipping Container test with image as this kind of test can't dump docker pushing process on cassettes")
	}
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
		CheckDestroy: testAccCheckScalewayContainerDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_container_namespace main {}
				`,
				Check: resource.ComposeTestCheckFunc(
					// Will set up the registry with the image required for the rest of the test
					testConfigContainerNamespace(tt, "scaleway_container_namespace.main"),
				),
			},
			{
				PreConfig: func() {
					return
				},
				Config: `
					resource scaleway_container_namespace main {}

					resource scaleway_container main {
						namespace_id = scaleway_container_namespace.main.id
						registry_image = "${scaleway_container_namespace.main.registry_endpoint}/alpine:test"
						deploy = true
					}

					data "dns_a_record_set" "main" {
					  host = scaleway_container.main.domain_name
					}
					resource "scaleway_container_domain" "main" {
					  container_id = scaleway_container.main.id
					  hostname    = "${data.dns_a_record_set.main.addrs[0]}.nip.io"
					
					  depends_on = [
						scaleway_container.main,
					  ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerDomainExists(tt, "scaleway_container_domain.main"),
				),
			},
			{
				Config: `
					resource scaleway_container_namespace main {}
				`,
			},
		},
	})
}

func testAccCheckScalewayContainerDomainExists(tt *TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDomain(&container.GetDomainRequest{
			DomainID: id,
			Region:   region,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayContainerDomainDestroy(tt *TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_domain" {
				continue
			}

			api, region, id, err := containerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.DeleteDomain(&container.DeleteDomainRequest{
				DomainID: id,
				Region:   region,
			})

			if err == nil {
				return fmt.Errorf("container domain (%s) still exists", rs.Primary.ID)
			}

			if !is404Error(err) {
				return err
			}
		}

		return nil
	}
}
