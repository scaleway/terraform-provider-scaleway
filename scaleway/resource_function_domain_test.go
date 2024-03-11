package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

func TestAccScalewayFunctionDomain_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	testDNSZone := "function-basic." + testDomain
	logging.L.Debugf("TestAccScalewayContainerDomain_Basic: test dns zone: %s", testDNSZone)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayFunctionDomainDestroy(tt),
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
				`,
				Check: testConfigContainerNamespace(tt, "scaleway_function_namespace.main"),
			},
			{
				Config: fmt.Sprintf(`
					resource scaleway_function_namespace main {}

					resource scaleway_function main {
						namespace_id = scaleway_function_namespace.main.id
						runtime = "go118"
						privacy = "private"
						handler = "Handle"
						zip_file = "testfixture/gofunction.zip"
						deploy = true
					}

					resource scaleway_domain_record "function" {
				  		dns_zone = "%s"
				  		name     = "container"
				  		type     = "CNAME"
				  		data     = "${scaleway_function.main.domain_name}."
				  		ttl      = 60
					}

					resource "scaleway_function_domain" "main" {
					  function_id = scaleway_function.main.id
					  hostname    = "${scaleway_domain_record.function.name}.${scaleway_domain_record.function.dns_zone}"
					}
				`, testDNSZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayFunctionDomainExists(tt, "scaleway_function_domain.main"),
				),
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
