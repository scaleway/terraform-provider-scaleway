package scaleway_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

func TestAccScalewayContainerDomain_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	testDNSZone := "container-basic." + testDomain
	logging.L.Debugf("TestAccScalewayContainerDomain_Basic: test dns zone: %s", testDNSZone)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayContainerDomainDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
				resource scaleway_container_namespace main {}
				`,
				Check: testConfigContainerNamespace(tt, "scaleway_container_namespace.main"),
			},
			{
				Config: fmt.Sprintf(`
				resource scaleway_container_namespace main {}

				resource scaleway_container app {
					registry_image = "${scaleway_container_namespace.main.registry_endpoint}/nginx:test"
					namespace_id = scaleway_container_namespace.main.id
					port = 80
					deploy = true
				}

				resource scaleway_domain_record "container" {
				  dns_zone = "%s"
				  name     = "container"
				  type     = "CNAME"
				  data     = "${scaleway_container.app.domain_name}."
				  ttl      = 60
				}

				resource scaleway_container_domain "app" {
				  container_id = scaleway_container.app.id
				  hostname = "${scaleway_domain_record.container.name}.${scaleway_domain_record.container.dns_zone}"
				}
			`, testDNSZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayContainerDomainExists(tt, "scaleway_container_domain.app"),
				),
			},
		},
	})
}

func testAccCheckScalewayContainerDomainExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := scaleway.ContainerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetDomain(&container.GetDomainRequest{
			Region:   region,
			DomainID: id,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayContainerDomainDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_container_domain" {
				continue
			}

			api, region, id, err := scaleway.ContainerAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetDomain(&container.GetDomainRequest{
				Region:   region,
				DomainID: id,
			})
			if httperrors.Is404(err) {
				return nil
			}

			if err != nil {
				return fmt.Errorf("failed to check if container domain exists: %w", err)
			}

			return nil
		}

		return nil
	}
}
