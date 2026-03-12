package instance_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	instancechecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/testfuncs"
)

func TestAccServer_PrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network"
						region = "fr-par"
					}

					resource scaleway_vpc_private_network internal {
						name = "tf-acc-server-private-network"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-private-network"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						zone = "fr-par-2"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network" ]

						private_network {
							pn_id = scaleway_vpc_private_network.internal.id
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.zone"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "private_network.0.pn_id",
						"scaleway_vpc_private_network.internal", "id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_ips.0.id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_ips.0.address"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network"
						region = "fr-par"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "tf-acc-server-private-network-01"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-private-network"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						zone  = "fr-par-1"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network" ]

						private_network {
							pn_id = scaleway_vpc_private_network.pn01.id
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.zone"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "private_network.0.pn_id",
						"scaleway_vpc_private_network.pn01", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network"
						region = "fr-par"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "tf-acc-server-private-network-01"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network pn02 {
						name = "tf-acc-server-private-network-02"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-private-network"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network" ]

						private_network {
							pn_id = scaleway_vpc_private_network.pn02.id
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.zone"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "private_network.0.pn_id",
						"scaleway_vpc_private_network.pn02", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network"
						region = "fr-par"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "tf-acc-server-private-network-01"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network pn02 {
						name = "tf-acc-server-private-network-02"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-private-network"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network" ]

						private_network {
							pn_id = scaleway_vpc_private_network.pn02.id
						}

						private_network {
							pn_id = scaleway_vpc_private_network.pn01.id
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					arePrivateNICsPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base",
						"private_network.0.pn_id",
						"scaleway_vpc_private_network.pn02", "id"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "2"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.0.zone"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.base", "private_network.1.pn_id",
						"scaleway_vpc_private_network.pn01", "id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.1.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.1.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.1.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.base", "private_network.1.zone"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network"
						region = "fr-par"
					}

					resource scaleway_vpc_private_network pn01 {
						name = "tf-acc-server-private-network-01"
						vpc_id = scaleway_vpc.main.id
					}

					resource scaleway_vpc_private_network pn02 {
						name = "tf-acc-server-private-network-02"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "base" {
						name = "tf-acc-server-private-network"
						image = "ubuntu_focal"
						type  = "DEV1-S"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network" ]
					}`,
				Check: resource.ComposeTestCheckFunc(
					instancechecks.IsServerPresent(tt, "scaleway_instance_server.base"),
					resource.TestCheckResourceAttr("scaleway_instance_server.base", "private_network.#", "0"),
				),
			},
		},
	})
}

func TestAccServer_PrivateNetworkMissingPNIC(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             instancechecks.IsServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network-missing-pnic"
					}

					resource scaleway_vpc_private_network pn {
						name = "tf-acc-server-private-network-missing-pnic"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-private-network-missing-pnic"
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network_missing_pnic" ]
						private_network {
							pn_id = scaleway_vpc_private_network.pn.id
						}
					}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PLAY2-PICO"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.zone"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.pnic_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "private_network.0.pn_id",
						"scaleway_vpc_private_network.pn", "id"),
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network-missing-pnic"
					}

					resource scaleway_vpc_private_network pn {
						name = "tf-acc-server-private-network-missing-pnic"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-private-network-missing-pnic"
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network_missing_pnic" ]
						private_network {
							pn_id = scaleway_vpc_private_network.pn.id
						}
					}

					resource scaleway_instance_private_nic pnic {
						private_network_id = scaleway_vpc_private_network.pn.id
						server_id = scaleway_instance_server.main.id
					}
`,
				ResourceName: "scaleway_instance_private_nic.pnic",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					serverID := state.RootModule().Resources["scaleway_instance_server.main"].Primary.ID

					pnicID, exists := state.RootModule().Resources["scaleway_instance_server.main"].Primary.Attributes["private_network.0.pnic_id"]
					if !exists {
						return "", errors.New("private_network.0.pnic_id not found")
					}

					id := serverID + "/" + pnicID

					return id, nil
				},
				ImportStatePersist: true,
			},
			{ // We import private nic as a separate resource to trigger its deletion.
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network-missing-pnic"
					}

					resource scaleway_vpc_private_network pn {
						name = "tf-acc-server-private-network-missing-pnic"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-private-network-missing-pnic"
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network_missing_pnic" ]
						private_network {
							pn_id = scaleway_vpc_private_network.pn.id
						}
					}

					resource scaleway_instance_private_nic pnic {
						private_network_id = scaleway_vpc_private_network.pn.id
						server_id = scaleway_instance_server.main.id
					}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PLAY2-PICO"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.zone"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.pnic_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "private_network.0.pn_id",
						"scaleway_vpc_private_network.pn", "id"),
					func(state *terraform.State) error {
						serverPNICID, exists := state.RootModule().Resources["scaleway_instance_server.main"].Primary.Attributes["private_network.0.pnic_id"]
						if !exists {
							return errors.New("private_network.0.pnic_id not found")
						}

						localizedPNICID := state.RootModule().Resources["scaleway_instance_private_nic.pnic"].Primary.ID

						_, pnicID, _, err := zonal.ParseNestedID(localizedPNICID)
						if err != nil {
							return err
						}

						if serverPNICID != pnicID {
							return fmt.Errorf("expected server pnic (%s) to equal standalone pnic id (%s)", serverPNICID, pnicID)
						}

						return nil
					},
				),
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network-missing-pnic"
					}

					resource scaleway_vpc_private_network pn {
						name = "tf-acc-server-private-network-missing-pnic"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-private-network-missing-pnic"
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network_missing_pnic" ]
						private_network {
							pn_id = scaleway_vpc_private_network.pn.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PLAY2-PICO"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.zone"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.pnic_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "private_network.0.pn_id",
						"scaleway_vpc_private_network.pn", "id"),
				),
				ExpectNonEmptyPlan: true, // pnic get deleted and the plan is not empty after the apply as private_network is now missing
			},
			{
				Config: `
					resource scaleway_vpc main {
						name = "tf-acc-server-private-network-missing-pnic"
					}

					resource scaleway_vpc_private_network pn {
						name = "tf-acc-server-private-network-missing-pnic"
						vpc_id = scaleway_vpc.main.id
					}

					resource "scaleway_instance_server" "main" {
						name = "tf-acc-server-private-network-missing-pnic"
						image = "ubuntu_jammy"
						type  = "PLAY2-PICO"
						tags  = [ "terraform-test", "scaleway_instance_server", "private_network_missing_pnic" ]
						private_network {
							pn_id = scaleway_vpc_private_network.pn.id
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "type", "PLAY2-PICO"),
					resource.TestCheckResourceAttr("scaleway_instance_server.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.mac_address"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.status"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.zone"),
					resource.TestCheckResourceAttrSet("scaleway_instance_server.main", "private_network.0.pnic_id"),
					resource.TestCheckResourceAttrPair("scaleway_instance_server.main", "private_network.0.pn_id",
						"scaleway_vpc_private_network.pn", "id"),
				),
			},
		},
	})
}
