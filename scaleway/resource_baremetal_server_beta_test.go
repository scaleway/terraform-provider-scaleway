package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_baremetal_server_beta", &resource.Sweeper{
		Name: "scaleway_baremetal_server_beta",
		F:    testSweepBaremetalServer,
	})
}

func testSweepBaremetalServer(region string) error {
	scwClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	baremetalAPI := baremetal.NewAPI(scwClient)

	scwRegion, err := scw.ParseRegion(region)
	if err != nil {
		return fmt.Errorf("error parsing region: %s", err)
	}

	for _, zone := range scwRegion.GetZones() {
		l.Debugf("sweeper: destroying the baremetal server in (%s)", zone)
		listServers, err := baremetalAPI.ListServers(&baremetal.ListServersRequest{
			Zone: zone,
		}, scw.WithAllPages())
		if err != nil {
			l.Warningf("error listing servers in (%s) in sweeper: %s", zone, err)
			continue
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
	}

	return nil
}

func TestAccScalewayBaremetalServerBetaMinimal1(t *testing.T) {
	SSHKeyName := newRandomName("ssh-key")
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM7HUxRyQtB2rnlhQUcbDGCZcTJg7OvoznOiyC9W6IxH opensource@scaleway.com"
	name := newRandomName("bm")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayBaremetalServerBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + SSHKeyName + `"
						public_key = "` + SSHKey + `"
					}
					
					resource "scaleway_baremetal_server_beta" "base" {
						name        = "` + name + `"
						zone        = "fr-par-2"
						description = "test a description"
						offer       = "GP-BM1-M"
						os_id       = "d17d6872-0412-45d9-a198-af82c34d3c5c"
					
						tags = [ "terraform-test", "scaleway_baremetal_server_beta", "minimal" ]
						ssh_key_ids = [ "${scaleway_account_ssh_key.main.id}" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerBetaExists("scaleway_baremetal_server_beta.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "name", name),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "offer", "GP-BM1-M/964f9b38-577e-470f-a220-7d762f9e8672"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "os_id", "d17d6872-0412-45d9-a198-af82c34d3c5c"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.1", "scaleway_baremetal_server_beta"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.2", "minimal"),
					testCheckResourceAttrUUID("scaleway_baremetal_server_beta.base", "ssh_key_ids.0"),
				),
			},
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + SSHKeyName + `"
						public_key = "` + SSHKey + `"
					}
					
					resource "scaleway_baremetal_server_beta" "base" {
						name        = "` + name + `"
						zone        = "fr-par-2"
						description = "test a description"
						offer       = "GP-BM1-M"
						os_id       = "d17d6872-0412-45d9-a198-af82c34d3c5c"
					
						tags = [ "terraform-test", "scaleway_baremetal_server_beta", "minimal", "edited" ]
						ssh_key_ids = [ "${scaleway_account_ssh_key.main.id}" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerBetaExists("scaleway_baremetal_server_beta.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "name", name),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "offer", "GP-BM1-M/964f9b38-577e-470f-a220-7d762f9e8672"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "os_id", "d17d6872-0412-45d9-a198-af82c34d3c5c"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.1", "scaleway_baremetal_server_beta"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.2", "minimal"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.3", "edited"),
					testCheckResourceAttrUUID("scaleway_baremetal_server_beta.base", "ssh_key_ids.0"),
				),
			},
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + SSHKeyName + `"
						public_key = "` + SSHKey + `"
					}
					
					resource "scaleway_baremetal_server_beta" "base" {
						name        = "` + name + `"
						zone        = "fr-par-2"
						description = "test a description"
						offer       = "GP-BM1-M"
						os_id       = "d859aa89-8b4a-4551-af42-ff7c0c27260a"
					
						tags = [ "terraform-test", "scaleway_baremetal_server_beta", "minimal", "edited" ]
						ssh_key_ids = [ "${scaleway_account_ssh_key.main.id}" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerBetaExists("scaleway_baremetal_server_beta.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "name", name),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "offer", "GP-BM1-M/964f9b38-577e-470f-a220-7d762f9e8672"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "os_id", "d859aa89-8b4a-4551-af42-ff7c0c27260a"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.1", "scaleway_baremetal_server_beta"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.2", "minimal"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.3", "edited"),
					testCheckResourceAttrUUID("scaleway_baremetal_server_beta.base", "ssh_key_ids.0"),
				),
			},
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + SSHKeyName + `"
						public_key = "` + SSHKey + `"
					}
					
					resource "scaleway_baremetal_server_beta" "base" {
						name        = "` + name + `"
						zone        = "fr-par-2"
						description = "test a description"
						offer       = "964f9b38-577e-470f-a220-7d762f9e8672"
						os_id       = "d859aa89-8b4a-4551-af42-ff7c0c27260a"
					
						tags = [ "terraform-test", "scaleway_baremetal_server_beta", "minimal", "edited" ]
						ssh_key_ids = [ "${scaleway_account_ssh_key.main.id}" ]
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerBetaExists("scaleway_baremetal_server_beta.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "name", name),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "offer", "GP-BM1-M/964f9b38-577e-470f-a220-7d762f9e8672"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "os_id", "d859aa89-8b4a-4551-af42-ff7c0c27260a"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.1", "scaleway_baremetal_server_beta"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.2", "minimal"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.3", "edited"),
					testCheckResourceAttrUUID("scaleway_baremetal_server_beta.base", "ssh_key_ids.0"),
				),
			},
		},
	})
}

func testAccCheckScalewayBaremetalServerBetaExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		baremetalAPI, zone, ID, err := baremetalAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = baremetalAPI.GetServer(&baremetal.GetServerRequest{ServerID: ID, Zone: zone})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckScalewayBaremetalServerBetaDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_baremetal_server_beta" {
			continue
		}

		baremetalAPI, zone, ID, err := baremetalAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = baremetalAPI.GetServer(&baremetal.GetServerRequest{
			ServerID: ID,
			Zone:     zone,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("server (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}
