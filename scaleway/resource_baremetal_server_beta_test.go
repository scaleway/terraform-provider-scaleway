package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
)

func TestAccScalewayBaremetalServerBetaMinimal1(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayBaremetalServerBetaDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayBaremetalServerBetaConfigMinimal1[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerBetaExists("scaleway_baremetal_server_beta.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "name", "namo-centos"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "offer_id", "964f9b38-577e-470f-a220-7d762f9e8672"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "os_id", "d17d6872-0412-45d9-a198-af82c34d3c5c"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.1", "scaleway_baremetal_server_beta"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.2", "minimal"),
				),
			},
			{
				Config: testAccCheckScalewayBaremetalServerBetaConfigMinimal1[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerBetaExists("scaleway_baremetal_server_beta.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "name", "namo-centos"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "offer_id", "964f9b38-577e-470f-a220-7d762f9e8672"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "os_id", "d17d6872-0412-45d9-a198-af82c34d3c5c"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.1", "scaleway_baremetal_server_beta"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.2", "minimal"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.3", "edited"),
				),
			},
			{
				Config: testAccCheckScalewayBaremetalServerBetaConfigMinimal1[2],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerBetaExists("scaleway_baremetal_server_beta.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "name", "namo-ubuntu"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "offer_id", "964f9b38-577e-470f-a220-7d762f9e8672"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "os_id", "d859aa89-8b4a-4551-af42-ff7c0c27260a"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.1", "scaleway_baremetal_server_beta"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.2", "minimal"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.3", "edited"),
				),
			},
			{
				Config: testAccCheckScalewayBaremetalServerBetaConfigMinimal1[3],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayBaremetalServerBetaExists("scaleway_baremetal_server_beta.base"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "name", "namo-ubuntu"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "offer_id", "964f9b38-577e-470f-a220-7d762f9e8672"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "os_id", "d859aa89-8b4a-4551-af42-ff7c0c27260a"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "description", "test a description"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.0", "terraform-test"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.1", "scaleway_baremetal_server_beta"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.2", "minimal"),
					resource.TestCheckResourceAttr("scaleway_baremetal_server_beta.base", "tags.3", "edited"),
					resource.TestMatchResourceAttr("scaleway_baremetal_server_beta.base", "ssh_key_ids.0", UUIDRegex),
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

		baremetalAPI, zone, ID, err := getBaremetalAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
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

		baremetalAPI, zone, ID, err := getBaremetalAPIWithZoneAndID(testAccProvider.Meta(), rs.Primary.ID)
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

const accountSSHKeyBm = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCve+G4fdcWk3YjfvAXDp9LWSLLbb+lm7XpMAbXW8D5f9l4RxQkekISLox7XeRN/az0q+6ob90xrWaX4L1oRxgZubARYh3XANXj8e9QxMTU9L0723Q2iWJ7Y0F3XBgTaDXDhdnIGZVpLuB98pKeZSOl8V3lhS2jSEFX+VV2DfkWfsV1u20A9PWpXGjwWtKm0uBwSbmitII8Zi5TXUrHyR1xLuuUWxql5yhzJjPAME5zf+TxVD12oBRsieKENFzgXmGHuS/JrMPmSQZSp1vOWN7/DSB5baWfXIZh0vEaEt5A+6MyVtuIiyZalgua201VIY263BmRXNBuOWMC4PkAuRQr opensource@scaleway.com"

var testAccCheckScalewayBaremetalServerBetaConfigMinimal1 = []string{`
resource "scaleway_account_ssh_key" "main" {
	name 	   = "main"
	public_key = "` + accountSSHKeyBm + `"
}

resource "scaleway_baremetal_server_beta" "base" {
	name        = "namo-centos"
	zone        = "fr-par-2"
	description = "test a description"
	offer_id    = "964f9b38-577e-470f-a220-7d762f9e8672"
	os_id       = "d17d6872-0412-45d9-a198-af82c34d3c5c"

	tags = [ "terraform-test", "scaleway_baremetal_server_beta", "minimal" ]
	ssh_key_ids = [ "${scaleway_account_ssh_key.main.id}" ]
}
`, `
resource "scaleway_account_ssh_key" "main" {
	name 	   = "main"
	public_key = "` + accountSSHKeyBm + `"
}

resource "scaleway_baremetal_server_beta" "base" {
	name        = "namo-centos"
	zone        = "fr-par-2"
	description = "test a description"
	offer_id    = "964f9b38-577e-470f-a220-7d762f9e8672"
	os_id       = "d17d6872-0412-45d9-a198-af82c34d3c5c"

	tags = [ "terraform-test", "scaleway_baremetal_server_beta", "minimal", "edited" ]
	ssh_key_ids = [ "${scaleway_account_ssh_key.main.id}" ]
}
`, `
resource "scaleway_account_ssh_key" "main" {
	name 	   = "main"
	public_key = "` + accountSSHKeyBm + `"
}

resource "scaleway_baremetal_server_beta" "base" {
	name        = "namo-ubuntu"
	zone        = "fr-par-2"
	description = "test a description"
	offer_id    = "964f9b38-577e-470f-a220-7d762f9e8672"
	os_id       = "d859aa89-8b4a-4551-af42-ff7c0c27260a"

	tags = [ "terraform-test", "scaleway_baremetal_server_beta", "minimal", "edited" ]
	ssh_key_ids = [ "${scaleway_account_ssh_key.main.id}" ]
}
`, `
resource "scaleway_account_ssh_key" "main" {
	name 	   = "main"
	public_key = "` + accountSSHKeyBm + `"
}

resource "scaleway_baremetal_server_beta" "base" {
	name        = "namo-ubuntu"
	zone        = "fr-par-2"
	description = "test a description"
	offer_id    = "964f9b38-577e-470f-a220-7d762f9e8672"
	os_id       = "d859aa89-8b4a-4551-af42-ff7c0c27260a"

	tags = [ "terraform-test", "scaleway_baremetal_server_beta", "minimal", "edited" ]
	ssh_key_ids = [ "${scaleway_account_ssh_key.main.id}" ]
}
`,
}
