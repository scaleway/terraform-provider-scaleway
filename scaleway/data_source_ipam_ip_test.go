package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayDataSourceIPAMIP_Instance(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-tests-ipam-ip-datasource-instance"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						name = "tf-tests-ipam-ip-datasource-instance"
					}

					resource "scaleway_instance_server" "main" {
						name  = "tf-tests-ipam-ip-datasource-instance"
						image = "ubuntu_jammy"
						type  = "PLAY2-MICRO"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "basic" ]
					}

					resource "scaleway_instance_private_nic" "main" {
						private_network_id = scaleway_vpc_private_network.main.id
						server_id = scaleway_instance_server.main.id
					}

					data "scaleway_ipam_ip" "main" {
						mac_address = scaleway_instance_private_nic.main.mac_address
						type = "ipv4"
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_ipam_ip.main", "address"),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceIPAMIP_InstanceLB(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayInstanceServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc" "main" {
						name = "tf-tests-ipam-ip-datasource-instance"
					}

					resource "scaleway_vpc_private_network" "main" {
						vpc_id = scaleway_vpc.main.id
						name = "tf-tests-ipam-ip-datasource-instance"
					}

					resource "scaleway_instance_server" "main" {
						name  = "tf-tests-ipam-ip-datasource-instance"
						image = "ubuntu_jammy"
						type  = "PLAY2-MICRO"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "basic" ]
					}

					resource "scaleway_instance_private_nic" "main" {
						private_network_id = scaleway_vpc_private_network.main.id
						server_id = scaleway_instance_server.main.id
					}

					data "scaleway_ipam_ip" "main" {
						mac_address = scaleway_instance_private_nic.main.mac_address
						type = "ipv4"
					}

					resource "scaleway_lb_ip" "main" {}

					resource "scaleway_lb" "main" {
						ip_id = scaleway_lb_ip.main.id
						type = "LB-S"
					}
					
					resource "scaleway_lb_backend" "main" {
						lb_id = scaleway_lb.main.id
						forward_protocol = "http"
						forward_port = "80"
						server_ips = [data.scaleway_ipam_ip.main.address]
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_ipam_ip.main", "address"),
				),
			},
		},
	})
}
