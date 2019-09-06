package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
)

func TestAccScalewayAccountSSHKey(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayAccountSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccScalewayAccountSSHKeyConfig[0],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", "main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", accountSSHKey),
				),
			},
			{
				Config: testAccScalewayAccountSSHKeyConfig[1],
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", "main-updated"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", accountSSHKey),
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

		accountAPI := getAccountAPI(testAccProvider.Meta())

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

		accountAPI := getAccountAPI(testAccProvider.Meta())

		_, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{
			SSHKeyID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

const accountSSHKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7P977mH29VxAEHy+rzuZjYzcMKdx2fYlQvg+9EXnhzadFY2tqimOy0GBMMN263KwATHcJ7tqnKS8ahg3mdWJjKZyBFIeWozCggxJGNbWDjpw5qiSvnPQjfeqRYZedS5vi/rAPZpGZVyXGeLUd01QEKhnMGUdBiLtaAg1UgBeDYQ== opensource@scaleway.com"

var testAccScalewayAccountSSHKeyConfig = []string{
	fmt.Sprintf(`
		resource "scaleway_account_ssh_key" "main" {
			name 	   = "main"
			public_key = "%s"
		}
	`, accountSSHKey),
	fmt.Sprintf(`
		resource "scaleway_account_ssh_key" "main" {
			name 	   = "main-updated"
			public_key = "%s"
		}
	`, accountSSHKey),
}
