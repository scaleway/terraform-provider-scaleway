package scaleway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/scaleway"
)

const SSHKeyBaremetal = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM7HUxRyQtB2rnlhQUcbDGCZcTJg7OvoznOiyC9W6IxH opensource@scaleway.com"

func init() {
	resource.AddTestSweepers("scaleway_baremetal_server", &resource.Sweeper{
		Name: "scaleway_baremetal_server",
		F:    testSweepBaremetalServer,
	})
}

func testSweepBaremetalServer(_ string) error {
	return sweepZones([]scw.Zone{scw.ZoneFrPar2}, func(scwClient *scw.Client, zone scw.Zone) error {
		baremetalAPI := baremetal.NewAPI(scwClient)
		logging.L.Debugf("sweeper: destroying the baremetal server in (%s)", zone)
		listServers, err := baremetalAPI.ListServers(&baremetal.ListServersRequest{Zone: zone}, scw.WithAllPages())
		if err != nil {
			logging.L.Warningf("error listing servers in (%s) in sweeper: %s", zone, err)
			return nil
		}

		for _, server := range listServers.Servers {
			_, err := baremetalAPI.DeleteServer(&baremetal.DeleteServerRequest{
				Zone:     zone,
				ServerID: server.ID,
			})
			if err != nil {
				return fmt.Errorf("error deleting server in sweeper: %s", err)
			}
		}

		return nil
	})
}

func TestAccScalewayBaremetalServer_Basic(t *testing.T) {
	t.Skip("Skipping Baremetal Server test as no stock is available currently")
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_Basic"
	name := "TestAccScalewayBaremetalServer_Basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBaremetalServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
					  zone    = "fr-par-2"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "fr-par-2"
						description = "test a description"
						offer       = "EM-B112X-SSD"
						os    = data.scaleway_baremetal_os.my_os.os_id
					
						tags = [ "terraform-test", "scaleway_baremetal_server", "minimal" ]
						ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "name", name),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "offer_id", "fr-par-2/a5065ba4-dde2-45f3-adec-1ebbb27b766b"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "os", "fr-par-2/96e5f0f2-d216-4de2-8a15-68730d877885"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "tags.1", "scaleway_baremetal_server"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "tags.2", "minimal"),
					testCheckResourceAttrUUID("scaleway_baremetal_server.base", "ssh_key_ids.0"),
				),
			},
			{
				// Trigger a reinstall and update tags
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
					  zone    = "fr-par-2"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}

					resource "scaleway_iam_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "fr-par-2"
						description = "test a description"
						offer       = "EM-B112X-SSD"
						os          = data.scaleway_baremetal_os.my_os.os_id
					
						tags = [ "terraform-test", "scaleway_baremetal_server", "minimal", "edited" ]
						ssh_key_ids = [ scaleway_iam_ssh_key.main.id ]
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "name", name),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "offer_id", "fr-par-2/a5065ba4-dde2-45f3-adec-1ebbb27b766b"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "os", "fr-par-2/96e5f0f2-d216-4de2-8a15-68730d877885"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "tags.#", "4"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "tags.1", "scaleway_baremetal_server"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "tags.2", "minimal"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "tags.3", "edited"),
					testCheckResourceAttrUUID("scaleway_baremetal_server.base", "ssh_key_ids.0"),
				),
			},
		},
	})
}

func TestAccScalewayBaremetalServer_RequiredInstallConfig(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBaremetalServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_baremetal_server" "base" {
						name        = "TestAccScalewayBaremetalServer_RequiredInstallConfig"
						zone        = "fr-par-2"
						offer       = "EM-B112X-SSD"
						os          = "7e865c16-1a63-4dc7-8181-dabc020fc21b" // Proxmox

						ssh_key_ids = []
					}`,
				ExpectError: regexp.MustCompile("attribute is required"),
			},
		},
	})
}

func TestAccScalewayBaremetalServer_WithoutInstallConfig(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBaremetalServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					data "scaleway_baremetal_offer" "my_offer" {
					  zone = "fr-par-2"
					  name = "EM-B112X-SSD"
					}

					resource "scaleway_baremetal_server" "base" {
					  name 			             = "TestAccScalewayBaremetalServer_WithoutInstallConfig"
                      zone     			         = "fr-par-2"
					  offer     				 = data.scaleway_baremetal_offer.my_offer.offer_id
					  install_config_afterward   = true
					}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "name", "TestAccScalewayBaremetalServer_WithoutInstallConfig"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "offer_id", "fr-par-2/a5065ba4-dde2-45f3-adec-1ebbb27b766b"),
					resource.TestCheckNoResourceAttr("scaleway_baremetal_server.base", "os"),
				),
			},
		},
	})
}

func TestAccScalewayBaremetalServer_CreateServerWithOption(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_CreateServerWithOption"
	name := "TestAccScalewayBaremetalServer_CreateServerWithOption"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBaremetalServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				data "scaleway_baremetal_os" "my_os" {
				  zone    = "fr-par-2"
				  name    = "Ubuntu"
				  version = "22.04 LTS (Jammy Jellyfish)"
				}
				
				data "scaleway_baremetal_offer" "my_offer" {
				  zone = "fr-par-2"
				  name = "EM-B112X-SSD"
				}
				
				data "scaleway_baremetal_option" "private_network" {
				  zone = "fr-par-2"
				  name = "Private Network"
				}
				
				resource "scaleway_iam_ssh_key" "base" {
				  name       = "%s"
				  public_key = "%s"
				}
				
				resource "scaleway_baremetal_server" "base" {
				  name  = "%s"
				  zone  = "fr-par-2"
				  offer = data.scaleway_baremetal_offer.my_offer.offer_id
				  os    = data.scaleway_baremetal_os.my_os.os_id
				
				  ssh_key_ids = [scaleway_iam_ssh_key.base.id]
				  options {
					id = data.scaleway_baremetal_option.private_network.option_id
				  }
				}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckScalewayBaremetalServerHasOptions(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_baremetal_server.base", "options.0.id", "data.scaleway_baremetal_option.private_network", "option_id"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "ips.#", "2"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "ipv4.#", "1"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "ipv4.0.version", "IPv4"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "ipv6.#", "1"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server.base", "ipv6.0.version", "IPv6"),
				),
			},
		},
	})
}

func TestAccScalewayBaremetalServer_AddOption(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_AddOption"
	name := "TestAccScalewayBaremetalServer_AddOption"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBaremetalServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "by_id" {
					  zone    = "fr-par-2"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}
					
					data "scaleway_baremetal_offer" "my_offer" {
					  zone = "fr-par-2"
					  name = "EM-B112X-SSD"
					}
					
					resource "scaleway_iam_ssh_key" "base" {
					  name       = "%s"
					  public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
					  name  = "%s"
					  zone  = "fr-par-2"
					  offer = data.scaleway_baremetal_offer.my_offer.offer_id
					  os    = data.scaleway_baremetal_os.by_id.os_id
					
					  ssh_key_ids = [scaleway_iam_ssh_key.base.id]
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
				),
			},
			{
				Config: fmt.Sprintf(`
				data "scaleway_baremetal_os" "my_os" {
				  zone    = "fr-par-2"
				  name    = "Ubuntu"
				  version = "22.04 LTS (Jammy Jellyfish)"
				}
				
				data "scaleway_baremetal_offer" "my_offer" {
				  zone = "fr-par-2"
				  name = "EM-B112X-SSD"
				}
				
				data "scaleway_baremetal_option" "private_network" {
				  zone = "fr-par-2"
				  name = "Private Network"
				}
				
				resource "scaleway_iam_ssh_key" "base" {
				  name       = "%s"
				  public_key = "%s"
				}
				
				resource "scaleway_baremetal_server" "base" {
				  name  = "%s"
				  zone  = "fr-par-2"
				  offer = data.scaleway_baremetal_offer.my_offer.offer_id
				  os    = data.scaleway_baremetal_os.my_os.os_id
				
				  ssh_key_ids = [scaleway_iam_ssh_key.base.id]
				  options {
					id = data.scaleway_baremetal_option.private_network.option_id
				  }
				}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckScalewayBaremetalServerHasOptions(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_baremetal_server.base", "options.0.id", "data.scaleway_baremetal_option.private_network", "option_id"),
				),
			},
		},
	})
}

func TestAccScalewayBaremetalServer_AddTwoOptionsThenDeleteOne(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_AddTwoOptionsThenDeleteOne"
	name := "TestAccScalewayBaremetalServer_AddTwoOptionsThenDeleteOne"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayBaremetalServerDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "by_id" {
					  zone    = "fr-par-2"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}
					
					data "scaleway_baremetal_offer" "my_offer" {
					  zone = "fr-par-2"
					  name = "EM-B112X-SSD"
					}
					
					resource "scaleway_iam_ssh_key" "base" {
					  name       = "%s"
					  public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
					  name  = "%s"
					  zone  = "fr-par-2"
					  offer = data.scaleway_baremetal_offer.my_offer.offer_id
					  os    = data.scaleway_baremetal_os.by_id.os_id
					
					  ssh_key_ids = [scaleway_iam_ssh_key.base.id]
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
					  zone    = "fr-par-2"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}
					
					data "scaleway_baremetal_offer" "my_offer" {
					  zone = "fr-par-2"
					  name = "EM-B112X-SSD"
					}
					
					data "scaleway_baremetal_option" "remote_access" {
					  zone = "fr-par-2"
					  name = "Remote Access"
					}
					
					data "scaleway_baremetal_option" "private_network" {
					  zone = "fr-par-2"
					  name = "Private Network"
					}
					
					resource "scaleway_iam_ssh_key" "base" {
					  name       = "%s"
					  public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
					  name        = "%s"
					  zone        = "fr-par-2"
					  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
					  os          = data.scaleway_baremetal_os.my_os.os_id
					  ssh_key_ids = [scaleway_iam_ssh_key.base.id]
					
					  options {
						id = data.scaleway_baremetal_option.private_network.option_id
					  }
					  options {
						id         = data.scaleway_baremetal_option.remote_access.option_id
						expires_at = "2025-07-06T09:00:00Z"
					  }
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckScalewayBaremetalServerHasOptions(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_baremetal_server.base", "options.*.id", "data.scaleway_baremetal_option.remote_access", "option_id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_baremetal_server.base", "options.*.id", "data.scaleway_baremetal_option.private_network", "option_id"),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_baremetal_server.base", "options.*", map[string]string{
						"id":         "fr-par-2/931df052-d713-4674-8b58-96a63244c8e2",
						"expires_at": "2025-07-06T09:00:00Z",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_baremetal_server.base", "options.*", map[string]string{
						"id": "fr-par-2/cd4158d7-2d65-49be-8803-c4b8ab6f760c",
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
					  zone    = "fr-par-2"
					  name    = "Ubuntu"
					  version = "22.04 LTS (Jammy Jellyfish)"
					}
					
					data "scaleway_baremetal_offer" "my_offer" {
					  zone = "fr-par-2"
					  name = "EM-B112X-SSD"
					}
					
					data "scaleway_baremetal_option" "remote_access" {
					  zone = "fr-par-2"
					  name = "Remote Access"
					}
					
					resource "scaleway_iam_ssh_key" "base" {
					  name       = "%s"
					  public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
					  name        = "%s"
					  zone        = "fr-par-2"
					  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
					  os          = data.scaleway_baremetal_os.my_os.os_id
					  ssh_key_ids = [scaleway_iam_ssh_key.base.id]
					
					  options {
						id         = data.scaleway_baremetal_option.remote_access.option_id
						expires_at = "2025-07-06T09:00:00Z"
					  }
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckScalewayBaremetalServerHasOptions(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_baremetal_server.base", "options.0.id", "data.scaleway_baremetal_option.remote_access", "option_id"),
					resource.TestCheckTypeSetElemNestedAttrs("scaleway_baremetal_server.base", "options.*", map[string]string{
						"id":         "fr-par-2/931df052-d713-4674-8b58-96a63244c8e2",
						"expires_at": "2025-07-06T09:00:00Z",
					}),
				),
			},
		},
	})
}

func TestAccScalewayBaremetalServer_CreateServerWithPrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_CreateServerWithPrivateNetwork"
	name := "TestAccScalewayBaremetalServer_CreateServerWithPrivateNetwork"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayBaremetalServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
						zone = "fr-par-2"
						name = "Ubuntu"
						version = "22.04 LTS (Jammy Jellyfish)"						
					}

					data "scaleway_baremetal_offer" "my_offer" {
						zone = "fr-par-2"
						name = "EM-B112X-SSD"
					}

					data "scaleway_baremetal_option" "private_network" {
						zone = "fr-par-2"
						name = "Private Network"
					}

					resource "scaleway_vpc_private_network" "pn" {
						name = "baremetal_private_network"
					} 

					resource "scaleway_iam_ssh_key" "base" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "fr-par-2"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
					
						ssh_key_ids = [ scaleway_iam_ssh_key.base.id ]
						options {
						  id = data.scaleway_baremetal_option.private_network.option_id
						}
						private_network {
						  id = scaleway_vpc_private_network.pn.id
						}
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckScalewayBaremetalServerHasPrivateNetwork(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_baremetal_server.base", "private_network.0.id", "scaleway_vpc_private_network.pn", "id"),
				),
			},
		},
	})
}

func TestAccScalewayBaremetalServer_AddPrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_AddPrivateNetwork"
	name := "TestAccScalewayBaremetalServer_AddPrivateNetwork"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayBaremetalServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
						zone = "fr-par-2"
						name = "Ubuntu"
						version = "22.04 LTS (Jammy Jellyfish)"						
					}

					data "scaleway_baremetal_offer" "my_offer" {
						zone = "fr-par-2"
						name = "EM-B112X-SSD"
					}

					data "scaleway_baremetal_option" "private_network" {
						zone = "fr-par-2"
						name = "Private Network"
					}

					resource "scaleway_vpc_private_network" "pn" {
						name = "baremetal_private_network"
					} 

					resource "scaleway_iam_ssh_key" "base" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "fr-par-2"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
					
						ssh_key_ids = [ scaleway_iam_ssh_key.base.id ]
						options {
						  id = data.scaleway_baremetal_option.private_network.option_id
						}
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
						zone = "fr-par-2"
						name = "Ubuntu"
						version = "22.04 LTS (Jammy Jellyfish)"						
					}

					data "scaleway_baremetal_offer" "my_offer" {
						zone = "fr-par-2"
						name = "EM-B112X-SSD"
					}

					data "scaleway_baremetal_option" "private_network" {
						zone = "fr-par-2"
						name = "Private Network"
					}

					resource "scaleway_vpc_private_network" "pn" {
						name = "baremetal_private_network"
					} 

					resource "scaleway_iam_ssh_key" "base" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "fr-par-2"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
					
						ssh_key_ids = [ scaleway_iam_ssh_key.base.id ]
						options {
						  id = data.scaleway_baremetal_option.private_network.option_id
						}
						private_network {
						  id = scaleway_vpc_private_network.pn.id
						}
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckScalewayBaremetalServerHasPrivateNetwork(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_baremetal_server.base", "private_network.0.id", "scaleway_vpc_private_network.pn", "id"),
				),
			},
		},
	})
}

func TestAccScalewayBaremetalServer_AddAnotherPrivateNetwork(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalServer_AddAnotherPrivateNetwork"
	name := "TestAccScalewayBaremetalServer_AddAnotherPrivateNetwork"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.TestAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckScalewayBaremetalServerDestroy(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
						zone = "fr-par-2"
						name = "Ubuntu"
						version = "22.04 LTS (Jammy Jellyfish)"						
					}

					data "scaleway_baremetal_offer" "my_offer" {
						zone = "fr-par-2"
						name = "EM-B112X-SSD"
					}

					data "scaleway_baremetal_option" "private_network" {
						zone = "fr-par-2"
						name = "Private Network"
					}

					resource "scaleway_vpc_private_network" "pn" {
						name = "baremetal_private_network"
					}

					resource "scaleway_iam_ssh_key" "base" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "fr-par-2"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
					
						ssh_key_ids = [ scaleway_iam_ssh_key.base.id ]
						options {
						  id = data.scaleway_baremetal_option.private_network.option_id
						}
						private_network {
						  id = scaleway_vpc_private_network.pn.id
						}
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckScalewayBaremetalServerHasPrivateNetwork(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckResourceAttrPair("scaleway_baremetal_server.base", "private_network.0.id", "scaleway_vpc_private_network.pn", "id"),
				),
			},
			{
				Config: fmt.Sprintf(`
					data "scaleway_baremetal_os" "my_os" {
						zone = "fr-par-2"
						name = "Ubuntu"
						version = "22.04 LTS (Jammy Jellyfish)"						
					}

					data "scaleway_baremetal_offer" "my_offer" {
						zone = "fr-par-2"
						name = "EM-B112X-SSD"
					}

					data "scaleway_baremetal_option" "private_network" {
						zone = "fr-par-2"
						name = "Private Network"
					}

					resource "scaleway_vpc_private_network" "pn" {
						name = "baremetal_private_network"
					} 

					resource "scaleway_vpc_private_network" "pn2" {
						name = "baremetal_private_network2"
					} 

					resource "scaleway_iam_ssh_key" "base" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "fr-par-2"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
					
						ssh_key_ids = [ scaleway_iam_ssh_key.base.id ]
						options {
						  id = data.scaleway_baremetal_option.private_network.option_id
						}
						private_network {
						  id = scaleway_vpc_private_network.pn.id
						}
						private_network {
						  id = scaleway_vpc_private_network.pn2.id
						}
					}
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerExists(tt, "scaleway_baremetal_server.base"),
					testAccCheckScalewayBaremetalServerHasPrivateNetwork(tt, "scaleway_baremetal_server.base"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_baremetal_server.base", "private_network.*.id", "scaleway_vpc_private_network.pn", "id"),
					resource.TestCheckTypeSetElemAttrPair("scaleway_baremetal_server.base", "private_network.*.id", "scaleway_vpc_private_network.pn2", "id"),
				),
			},
		},
	})
}

func testAccCheckScalewayBaremetalServerExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		baremetalAPI, zonedID, err := scaleway.BaremetalAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = baremetalAPI.GetServer(&baremetal.GetServerRequest{
			ServerID: zonedID.ID,
			Zone:     zonedID.Zone,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayBaremetalServerDestroy(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_baremetal_server" {
				continue
			}

			baremetalAPI, zonedID, err := scaleway.BaremetalAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = baremetalAPI.GetServer(&baremetal.GetServerRequest{
				ServerID: zonedID.ID,
				Zone:     zonedID.Zone,
			})

			// If no error resource still exist
			if err == nil {
				return fmt.Errorf("server (%s) still exists", rs.Primary.ID)
			}

			// Unexpected api error we return it
			if !httperrors.Is404(err) {
				return err
			}
		}
		return nil
	}
}

func testAccCheckScalewayBaremetalServerHasOptions(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		baremetalAPI, zonedID, err := scaleway.BaremetalAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		server, err := baremetalAPI.GetServer(&baremetal.GetServerRequest{
			ServerID: zonedID.ID,
			Zone:     zonedID.Zone,
		})
		if err != nil {
			return err
		}

		if len(server.Options) == 0 {
			return fmt.Errorf("server (%s) has no options enabled", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckScalewayBaremetalServerHasPrivateNetwork(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		_, zonedID, err := scaleway.BaremetalAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		baremetalPrivateNetworkAPI, _, err := scaleway.BaremetalPrivateNetworkAPIWithZoneAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		listPrivateNetworks, err := baremetalPrivateNetworkAPI.ListServerPrivateNetworks(&baremetal.PrivateNetworkAPIListServerPrivateNetworksRequest{
			Zone:     zonedID.Zone,
			ServerID: &zonedID.ID,
		})
		if err != nil {
			return err
		}

		if len(listPrivateNetworks.ServerPrivateNetworks) == 0 {
			return fmt.Errorf("server (%s) has no private networks attached to it", rs.Primary.ID)
		}

		return nil
	}
}
