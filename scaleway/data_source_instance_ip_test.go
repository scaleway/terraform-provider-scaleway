package scaleway

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceInstanceIP_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	ipregexp := regexp.MustCompile("(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `resource "scaleway_instance_ip" "ip" {}`,
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {}

					data "scaleway_instance_ip" "ip-from-address" {
						address = "${scaleway_instance_ip.ip.address}"
					}

					data "scaleway_instance_ip" "ip-from-id" {
						id = "${scaleway_instance_ip.ip.id}"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("scaleway_instance_ip.ip", "address", ipregexp),
					resource.TestMatchResourceAttr("data.scaleway_instance_ip.ip-from-address", "address", ipregexp),
					resource.TestMatchResourceAttr("data.scaleway_instance_ip.ip-from-id", "address", ipregexp),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.ip", "address", "data.scaleway_instance_ip.ip-from-address", "address"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.ip", "address", "data.scaleway_instance_ip.ip-from-id", "address"),
				),
			},
		},
	})
}
