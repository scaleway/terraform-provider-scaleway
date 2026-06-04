package instance_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccServer_IPs(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "ip1" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-ips"
						ip_ids = [scaleway_instance_ip.ip1.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "public_ips.0.id", "scaleway_instance_ip.ip1", "id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "public_ips.0.address"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.0.gateway", "62.210.0.1"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.0.netmask", "32"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.0.family", "inet"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.0.dynamic", "false"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.0.provisioning_mode", "dhcp"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip1" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_ip" "ip2" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-ips"
						ip_ids = [scaleway_instance_ip.ip1.id, scaleway_instance_ip.ip2.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", "2"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "public_ips.0.id", "scaleway_instance_ip.ip1", "id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "public_ips.1.id", "scaleway_instance_ip.ip2", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "ip1" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_ip" "ip2" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-ips"
						ip_ids = [scaleway_instance_ip.ip2.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "public_ips.0.id", "scaleway_instance_ip.ip2", "id"),
				),
			},
		},
	})
}

func TestAccServer_IPRemoved(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "main" {}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-ip-removed"
						ip_id = scaleway_instance_ip.main.id
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ip_removed" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "public_ips.0.id", "scaleway_instance_ip.main", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "main" {}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-ip-removed"
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ip_removed" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					serverHasNoIPAssigned(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", "0"),
				),
			},
		},
	})
}

func TestAccServer_IPsRemoved(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "main" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-ips-removed"
						ip_ids = [scaleway_instance_ip.main.id]
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips_removed" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", "1"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "public_ips.0.id", "scaleway_instance_ip.main", "id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "main" {
						type = "routed_ipv4"
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-ips-removed"
						image = "ubuntu_jammy"
						type  = "PRO2-XXS"
						state = "stopped"
						tags  = [ "terraform-test", "scaleway_instance_server", "ips_removed" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.main"),
					serverHasNoIPAssigned(tt, "scaleway_instance_server.main"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", "0"),
				),
			},
		},
	})
}

func TestAccServer_WithReservedIP(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-with-reserved-ip"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						ip_id = scaleway_instance_ip.first.id
						tags  = [ "terraform-test", "scaleway_instance_server", "with_reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.first", "address", "scaleway_instance_server.base", "public_ips.0.address"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.first", "id", "scaleway_instance_server.base", "ip_id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-with-reserved-ip"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						ip_id = scaleway_instance_ip.second.id
						tags  = [ "terraform-test", "scaleway_instance_server", "with_reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					isIPAttachedToServer(tt, "scaleway_instance_ip.second", "scaleway_instance_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.second", "address", "scaleway_instance_server.base", "public_ips.0.address"),
					resource.TestCheckResourceAttrPair("scaleway_instance_ip.second", "id", "scaleway_instance_server.base", "ip_id"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-with-reserved-ip"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "with_reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					serverHasNoIPAssigned(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "public_ips.#", "0"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "ip_id", ""),
				),
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-with-reserved-ip"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						enable_dynamic_ip = true
						tags  = [ "terraform-test", "scaleway_instance_server", "with_reserved_ip" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					serverHasNoIPAssigned(tt, "scaleway_instance_server.base"),
					acctest.CheckResourceAttrIPv4("scaleway_instance_server.base", "public_ips.0.address"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "ip_id", ""),
				),
			},
		},
	})
}

func TestAccServer_Ipv6(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_instance_ip" "ip" {
						type = "routed_ipv6"
					}

					resource "scaleway_instance_server" "server01" {
						name = "tf-acc-server-ipv6"
						image = "ubuntu_focal"
						type  = "PLAY2-PICO"
						tags  = [ "terraform-test", "scaleway_instance_server", "ipv6" ]
						ip_ids = [scaleway_instance_ip.ip.id]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.server01"),
					acctest.CheckResourceAttrIPv6("scaleway_instance_server.server01", "public_ips.0.address"),
				),
			},
			{
				Config: `
					resource "scaleway_instance_server" "server01" {
						name = "tf-acc-server-ipv6"
						image = "ubuntu_focal"
						type  = "PLAY2-PICO"
						tags  = [ "terraform-test", "scaleway_instance_server", "ipv6" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.server01"),
					resource.TestCheckResourceAttr("scaleway_instance_server.server01", "public_ips.#", "0"),
				),
			},
		},
	})
}
