package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
)

func TestAccScalewayAccountSSHKey(t *testing.T) {
	name := newRandomName("ssh-key")
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEEYrzDOZmhItdKaDAEqJQ4ORS2GyBMtBozYsK5kiXXX opensource@scaleway.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayAccountSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `"
						public_key = "` + SSHKey + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `-updated"
						public_key = "` + SSHKey + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name+"-updated"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func TestAccScalewayAccountSSHKey_WithNewLine(t *testing.T) {
	name := newRandomName("ssh-key")
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDjfkdWCwkYlVQMDUfiZlVrmjaGOfBYnmkucssae8Iup opensource@scaleway.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayAccountSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `"
						public_key = "\n\n` + SSHKey + `\n\n"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func testAccCheckScalewayAccountSSHKeyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_account_ssh_key" {
			continue
		}

		accountAPI := accountAPI(testAccProvider.Meta())

		_, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{
			SSHKeyID: rs.Primary.ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("SSH key (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}

func testAccCheckScalewayAccountSSHKeyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		accountAPI := accountAPI(testAccProvider.Meta())

		_, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{
			SSHKeyID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
