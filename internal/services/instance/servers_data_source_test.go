package instance_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccDataSourceServers_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
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
						tags = ["data_scaleway_instance_servers", "terraform-test"]
					}

					data "scaleway_instance_servers" "servers_by_name_other_zone" {
						name = "tf-server-datasource"
						zone = "fr-par-2"
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_tag", "servers.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_tag", "servers.1.id"),
					// public_ips must be a list (not null) even when empty - a schema mismatch
					// between flattenServerPublicIPs fields and the declared schema caused the plugin to crash.
					resource.TestCheckResourceAttr("data.scaleway_instance_servers.servers_by_tag", "servers.0.public_ips.#", "0"),

					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_name", "servers.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_name", "servers.1.id"),
					resource.TestCheckResourceAttr("data.scaleway_instance_servers.servers_by_name", "servers.0.public_ips.#", "0"),

					resource.TestCheckNoResourceAttr("data.scaleway_instance_servers.servers_by_name_other_zone", "servers.0.id"),
				),
			},
		},
	})
}

func TestAccDataSourceServers_PrivateIPs(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_vpc_private_network" "pn01" {
					  name = "private_network_instance_servers"
					}

					resource "scaleway_instance_server" "server1" {
						name  = "tf-server-datasource-private-ips-0"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "private-ips" ]

					    private_network {
						  pn_id = scaleway_vpc_private_network.pn01.id
					    }
					}`,
			},
			{
				Config: `
					resource "scaleway_vpc_private_network" "pn01" {
					  name = "private_network_instance_servers"
					}

					resource "scaleway_instance_server" "server1" {
						name  = "tf-server-datasource-private-ips-0"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "private-ips" ]

					    private_network {
						  pn_id = scaleway_vpc_private_network.pn01.id
					    }
					}

					resource "scaleway_instance_server" "server2" {
						name  = "tf-server-datasource-private-ips-1"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "private-ips" ]

					    private_network {
						  pn_id = scaleway_vpc_private_network.pn01.id
					    }
					}`,
			},
			{
				Config: `
					resource "scaleway_vpc_private_network" "pn01" {
					  name = "private_network_instance_servers"
					}

					resource "scaleway_instance_server" "server1" {
						name  = "tf-server-datasource-private-ips-0"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "private-ips" ]

					    private_network {
						  pn_id = scaleway_vpc_private_network.pn01.id
					    }
					}

					resource "scaleway_instance_server" "server2" {
						name  = "tf-server-datasource-private-ips-1"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						state = "stopped"
						tags  = [ "terraform-test", "data_scaleway_instance_servers", "private-ips" ]

					    private_network {
						  pn_id = scaleway_vpc_private_network.pn01.id
					    }
					}

					data "scaleway_instance_servers" "servers_by_name" {
						name = "tf-server-datasource-private-ips"
					}
					
					data "scaleway_instance_servers" "servers_by_tag" {
						tags = ["data_scaleway_instance_servers", "terraform-test", "private-ips"]
					}

					data "scaleway_instance_servers" "servers_by_name_other_zone" {
						name = "tf-server-datasource-private-ips"
						zone = "fr-par-2"
					}
					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_tag", "servers.0.private_ips.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_tag", "servers.0.private_ips.1.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_tag", "servers.1.private_ips.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_tag", "servers.1.private_ips.1.id"),

					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_name", "servers.0.private_ips.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_name", "servers.0.private_ips.1.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_name", "servers.1.private_ips.0.id"),
					resource.TestCheckResourceAttrSet("data.scaleway_instance_servers.servers_by_name", "servers.1.private_ips.1.id"),

					resource.TestCheckNoResourceAttr("data.scaleway_instance_servers.servers_by_name_other_zone", "servers.0.id"),
				),
			},
		},
	})
}

func TestAccDataSourceServers_PublicIPs(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "ip0" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "server0" {
						name = "tf-acc-server-ips-0"
						ip_ids = [scaleway_instance_ip.ip0.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips" ]
					}`,
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip0" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "server0" {
						name = "tf-acc-server-ips-0"
						ip_ids = [scaleway_instance_ip.ip0.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips" ]
					}

					resource "scaleway_instance_ip" "ip1" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "server1" {
						name = "tf-acc-server-ips-1"
						ip_ids = [scaleway_instance_ip.ip1.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips" ]
					}`,
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip0" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "server0" {
						name = "tf-acc-server-ips-0"
						ip_ids = [scaleway_instance_ip.ip0.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips" ]
					}

					resource "scaleway_instance_ip" "ip1" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "server1" {
						name = "tf-acc-server-ips-1"
						ip_ids = [scaleway_instance_ip.ip1.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips" ]
					}

					data "scaleway_instance_servers" "servers" {
					}

					data "scaleway_instance_servers" "server0" {
						name = scaleway_instance_server.server0.name
					}

					data "scaleway_instance_servers" "server1" {
						name = scaleway_instance_server.server1.name
					}

					`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server0", "servers.0.id", "scaleway_instance_server.server0", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server0", "servers.0.public_ips.0.id", "scaleway_instance_ip.ip0", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server0", "servers.0.name", "scaleway_instance_server.server0", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server0", "servers.0.type", "scaleway_instance_server.server0", "type"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server0", "servers.0.state", "scaleway_instance_server.server0", "state"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server0", "servers.0.enable_dynamic_ip", "scaleway_instance_server.server0", "enable_dynamic_ip"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server0", "servers.0.security_group_id", "scaleway_instance_server.server0", "security_group_id"),
					resource.TestCheckResourceAttr("data.scaleway_instance_servers.server0", "servers.0.tags.#", "3"),
					resource.TestCheckTypeSetElemAttr("data.scaleway_instance_servers.server0", "servers.0.tags.*", "terraform-test"),
					resource.TestCheckTypeSetElemAttr("data.scaleway_instance_servers.server0", "servers.0.tags.*", "scaleway_instance_server"),
					resource.TestCheckTypeSetElemAttr("data.scaleway_instance_servers.server0", "servers.0.tags.*", "ips"),

					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server1", "servers.0.id", "scaleway_instance_server.server1", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server1", "servers.0.public_ips.0.id", "scaleway_instance_ip.ip1", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server1", "servers.0.name", "scaleway_instance_server.server1", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server1", "servers.0.type", "scaleway_instance_server.server1", "type"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server1", "servers.0.state", "scaleway_instance_server.server1", "state"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server1", "servers.0.enable_dynamic_ip", "scaleway_instance_server.server1", "enable_dynamic_ip"),
					resource.TestCheckResourceAttrPair("data.scaleway_instance_servers.server1", "servers.0.security_group_id", "scaleway_instance_server.server1", "security_group_id"),
					resource.TestCheckResourceAttr("data.scaleway_instance_servers.server1", "servers.0.tags.#", "3"),
				),
			},
		},
	})
}
