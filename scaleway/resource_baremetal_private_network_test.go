package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayBaremetalPrivateNetwork_Basic(t *testing.T) {
	tt := NewTestTools(t)
	defer tt.Cleanup()

	SSHKeyName := "TestAccScalewayBaremetalPrivateNetwork_Basic"
	name := "TestAccScalewayBaremetalPrivateNetwork_Basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
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
						zone = "fr-par-2"
						name = "baremetal_private_network"
					} 

					resource "scaleway_account_ssh_key" "base" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					resource "scaleway_baremetal_server" "base" {
						name        = "%s"
						zone        = "fr-par-2"
						offer       = data.scaleway_baremetal_offer.my_offer.offer_id
						os          = data.scaleway_baremetal_os.my_os.os_id
					
						ssh_key_ids = [ scaleway_account_ssh_key.base.id ]
						option_ids = [ data.scaleway_baremetal_option.private_network.option_id ]
					}
					
					resource "scaleway_baremetal_private_network" "base" {
					    zone = "fr-par-2"
				
					    server_id = scaleway_baremetal_server.base.id
					    private_network_ids = [scaleway_vpc_private_network.pn.id]
				    }
				`, SSHKeyName, SSHKeyBaremetal, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_baremetal_private_network.base", "zone", "fr-par-2"),
					resource.TestCheckResourceAttrPair("scaleway_baremetal_private_network.base", "server_id", "scaleway_baremetal_server.base", "id"),
					resource.TestCheckResourceAttrPair("scaleway_baremetal_private_network.base", "private_network_ids.0", "scaleway_vpc_private_network.pn", "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
