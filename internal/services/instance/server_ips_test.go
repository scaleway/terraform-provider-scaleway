package instance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
				ResourceName:            "scaleway_instance_server.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_id"},
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
				ResourceName:            "scaleway_instance_server.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_id"},
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
			{
				ResourceName:            "scaleway_instance_server.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_id"},
			},
		},
	})
}

func TestAccServer_PublicIPsDependentUpdate(t *testing.T) {
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
						name   = "tf-acc-server-public-ips-dependent"
						ip_ids = [scaleway_instance_ip.ip1.id]
						image  = "ubuntu_jammy"
						type   = "PRO2-XXS"
						state  = "stopped"
						tags   = ["terraform-test", "scaleway_instance_server", "public_ips_dependent"]
					}

					resource "terraform_data" "dependent" {
						input = scaleway_instance_server.main.public_ips[0].address
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "public_ips.0.id", "scaleway_instance_ip.ip1", "id"),
					resource.TestCheckResourceAttrPair("terraform_data.dependent", "output", "scaleway_instance_ip.ip1", "address"),
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
						name   = "tf-acc-server-public-ips-dependent"
						ip_ids = [scaleway_instance_ip.ip2.id]
						image  = "ubuntu_jammy"
						type   = "PRO2-XXS"
						state  = "stopped"
						tags   = ["terraform-test", "scaleway_instance_server", "public_ips_dependent"]
					}

					resource "terraform_data" "dependent" {
						input = scaleway_instance_server.main.public_ips[0].address
					}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue("scaleway_instance_server.main", tfjsonpath.New("public_ips")),
						plancheck.ExpectResourceAction("terraform_data.dependent", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "public_ips.0.id", "scaleway_instance_ip.ip2", "id"),
					resource.TestCheckResourceAttrPair("terraform_data.dependent", "output", "scaleway_instance_ip.ip2", "address"),
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
				ResourceName:            "scaleway_instance_server.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_ids"},
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
			{
				ResourceName:            "scaleway_instance_server.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_ids"},
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
			{
				ResourceName:            "scaleway_instance_server.main",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_id"},
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
						tags  = [ "terraform-test", "scaleway_instance_server", "with_reserved_ip", "step1" ]
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
						tags  = [ "terraform-test", "scaleway_instance_server", "with_reserved_ip", "step2" ]
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
				ResourceName:            "scaleway_instance_server.base",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_ids"},
			},
			{
				Config: `
					resource "scaleway_instance_ip" "first" {}
					resource "scaleway_instance_ip" "second" {}
					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-with-reserved-ip"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "with_reserved_ip", "step4" ]
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
						tags  = [ "terraform-test", "scaleway_instance_server", "with_reserved_ip", "step5" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					serverHasNoIPAssigned(tt, "scaleway_instance_server.base"),
					acctest.CheckResourceAttrIPv4("scaleway_instance_server.base", "public_ips.0.address"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "ip_id", ""),
				),
			},
			{
				ResourceName:            "scaleway_instance_server.base",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_ids"},
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
				ResourceName:            "scaleway_instance_server.server01",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"replace_on_type_change", "image", "ip_id"},
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

// Importing a server sets both ip_id and ip_ids, as the import cannot know which one the configuration uses.
// Whichever attribute the configuration uses, the import must not plan an IP update (it would detach every
// public IP) and the plan after a refresh must stay empty. See issue #4394.
func TestAccServer_IPsImport(t *testing.T) {
	ips := `
		resource "scaleway_instance_ip" "ip1" {
			type = "routed_ipv4"
		}

		# Different types keep the two create requests distinguishable in the cassette.
		resource "scaleway_instance_ip" "ip2" {
			type = "routed_ipv6"
		}
	`
	server := `
		resource "scaleway_instance_server" "main" {
			name   = "tf-acc-server-ips-import"
			ip_ids = [scaleway_instance_ip.ip1.id, scaleway_instance_ip.ip2.id]
			image  = "ubuntu_jammy"
			type   = "PRO2-XXS"
			state  = "stopped"
			tags   = [ "terraform-test", "scaleway_instance_server", "ips-import" ]

			# Not read back on import, like ImportStateVerifyIgnore in the other tests.
			lifecycle {
				ignore_changes = [image, replace_on_type_change]
			}
		}
	`

	testAccServerImportThenPlan(t, "tf-acc-server-ips-import", ips, server, "2")
}

func TestAccServer_IPImport(t *testing.T) {
	ips := `
		resource "scaleway_instance_ip" "main" {
			type = "routed_ipv4"
		}
	`
	server := `
		resource "scaleway_instance_server" "main" {
			name  = "tf-acc-server-ip-import"
			ip_id = scaleway_instance_ip.main.id
			image = "ubuntu_jammy"
			type  = "PRO2-XXS"
			state = "stopped"
			tags  = [ "terraform-test", "scaleway_instance_server", "ip-import" ]

			# Not read back on import, like ImportStateVerifyIgnore in the other tests.
			lifecycle {
				ignore_changes = [image, replace_on_type_change]
			}
		}
	`

	testAccServerImportThenPlan(t, "tf-acc-server-ip-import", ips, server, "1")
}

// testAccServerImportThenPlan creates the server, removes it from the state, imports it back with an
// import block, then checks that the import is a no-op and that a later plan is empty.
func testAccServerImportThenPlan(t *testing.T, serverName, ips, server, publicIPsCount string) {
	t.Helper()

	if acctest.IsRunningOpenTofu() {
		t.Skip("Skipping: OpenTofu refreshes a resource targeted by a removed block before forgetting it and Terraform does not, so the recorded interactions do not match")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	forget := `
		removed {
			from = scaleway_instance_server.main

			lifecycle {
				destroy = false
			}
		}
	`
	importBlock := fmt.Sprintf(`
		data "scaleway_instance_server" "main" {
			name = %q
		}

		import {
			to = scaleway_instance_server.main
			id = data.scaleway_instance_server.main.id
		}
	`, serverName)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: ips + server,
				Check:  resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", publicIPsCount),
			},
			{
				Config: ips + forget,
			},
			{
				Config: ips + server + importBlock,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("scaleway_instance_server.main", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.TestCheckResourceAttr("scaleway_instance_server.main", "public_ips.#", publicIPsCount),
			},
			{
				Config:   ips + server,
				PlanOnly: true,
			},
		},
	})
}
