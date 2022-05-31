package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceInstanceServers_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_server" "server1" {
						name  = "tf-server-datasource0"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "basic" ]
					}`,
			},
			{
				Config: `
					resource "scaleway_instance_server" "server1" {
						name  = "tf-server-datasource0"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "basic" ]
					}

					resource "scaleway_instance_server" "server2" {
						name  = "tf-server-datasource1"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "basic" ]
					}`,
			},
			{
				Config: `
					resource "scaleway_instance_server" "server1" {
						name  = "tf-server-datasource0"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "basic" ]
					}

					resource "scaleway_instance_server" "server2" {
						name  = "tf-server-datasource1"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "basic" ]
					}

					data "scaleway_instance_servers" "servers_by_name" {
						name = "tf-server-datasource"
					}
					
					data "scaleway_instance_servers" "servers_by_tag" {
						tags = ["data_scaleway_instance_servers"]
					}

					data "scaleway_instance_servers" "servers_by_name_other_zone" {
						name = "tf-server-datasource"
						zone = "fr-par-2"
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_tag", "servers.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_tag", "servers.1.id"),

					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_name", "servers.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_name", "servers.1.id"),

					resource.TestCheckNoResourceAttr("data.scaleway_instance_servers.servers_by_name_other_zone", "servers.0.id"),
				),
			},
		},
	})
}
