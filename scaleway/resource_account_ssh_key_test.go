package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccScalewayAccountSSHKey_basic(t *testing.T) {
	name := "tf-test-account-ssh-key-basic"
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEEYrzDOZmhItdKaDAEqJQ4ORS2GyBMtBozYsK5kiXXX opensource@scaleway.com"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `"
						public_key = "` + SSHKey + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_account_ssh_key.main"),
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
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name+"-updated"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func TestAccScalewayAccountSSHKey_WithNewLine(t *testing.T) {
	name := "tf-test-account-ssh-key-newline"
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDjfkdWCwkYlVQMDUfiZlVrmjaGOfBYnmkucssae8Iup opensource@scaleway.com"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `"
						public_key = "\n\n` + SSHKey + `\n\n"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func TestAccScalewayAccountSSHKey_ChangeResourceName(t *testing.T) {
	name := "TestAccScalewayAccountSSHKey_ChangeResourceName"
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAICJEoOOgQBLJPs4g/XcPTKT82NywNPpxeuA20FlOPlpO opensource@scaleway.com"
	tt := NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      testAccCheckScalewayIamSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "first" {
						name 	   = "` + name + `"
						public_key = "\n\n` + SSHKey + `\n\n"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_account_ssh_key.first"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.first", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.first", "public_key", SSHKey),
				),
			},
			{
				Config: `
					resource "scaleway_account_ssh_key" "second" {
						name 	   = "` + name + `"
						public_key = "\n\n` + SSHKey + `\n\n"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayIamSSHKeyExists(tt, "scaleway_account_ssh_key.second"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.second", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.second", "public_key", SSHKey),
				),
			},
		},
	})
}
