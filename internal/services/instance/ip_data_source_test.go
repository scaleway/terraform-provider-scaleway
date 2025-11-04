package instance_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccDataSourceIP_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsIPDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {}

					resource "scaleway_instance_ip" "ip_fr-par-2" {
						zone = "fr-par-2"
					}

					data "scaleway_instance_ip" "ip-from-address" {
						address = scaleway_instance_ip.ip.address
					}

					data "scaleway_instance_ip" "ip-from-id" {
						id = scaleway_instance_ip.ip.id
					}

					data "scaleway_instance_ip" "test_another_zone" {
						address = scaleway_instance_ip.ip_fr-par-2.address
						zone    = "fr-par-2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrIP("scaleway_instance_ip.ip", "address"),
					acctest.CheckResourceAttrIP("data.scaleway_instance_ip.ip-from-address", "address"),
					acctest.CheckResourceAttrIP("data.scaleway_instance_ip.ip-from-id", "address"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.ip", "address", "data.scaleway_instance_ip.ip-from-address", "address"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.ip", "address", "data.scaleway_instance_ip.ip-from-id", "address"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.ip_fr-par-2", "address", "data.scaleway_instance_ip.test_another_zone", "address"),
				),
			},
		},
	})
}
