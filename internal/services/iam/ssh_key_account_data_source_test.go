package iam_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	iamchecks "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/iam/testfuncs"
)

func TestAccDataSourceAccountSSHKey_Basic(t *testing.T) {
	dataSourceAccountSSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILHy/M5FVm5ydLGcal3e5LNcfTalbeN7QL/ZGCvDEdqJ foobar@example.com"
	dataSourceAccountSSHKeyWithoutComment := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILHy/M5FVm5ydLGcal3e5LNcfTalbeN7QL/ZGCvDEdqJ"
	sshKeyName := "TestAccScalewayDataSourceAccountSSHKey_Basic"

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: tt.ProviderFactories,
		CheckDestroy:      iamchecks.CheckSSHKeyDestroy(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}
					`, sshKeyName, dataSourceAccountSSHKey),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "%s"
						public_key = "%s"
					}
					
					data "scaleway_account_ssh_key" "prod" {
						name = "${scaleway_account_ssh_key.main.name}"
					}
					
					data "scaleway_account_ssh_key" "stg" {
						ssh_key_id = "${scaleway_account_ssh_key.main.id}"
					}`, sshKeyName, dataSourceAccountSSHKey),
				Check: resource.ComposeTestCheckFunc(
					iamchecks.CheckSSHKeyExists(tt, "data.scaleway_account_ssh_key.prod"),
					resource.TestCheckResourceAttr("data.scaleway_account_ssh_key.prod", "name", sshKeyName),
					resource.TestCheckResourceAttr("data.scaleway_account_ssh_key.prod", "public_key", dataSourceAccountSSHKeyWithoutComment),
					iamchecks.CheckSSHKeyExists(tt, "data.scaleway_account_ssh_key.stg"),
					resource.TestCheckResourceAttr("data.scaleway_account_ssh_key.stg", "name", sshKeyName),
					resource.TestCheckResourceAttr("data.scaleway_account_ssh_key.stg", "public_key", dataSourceAccountSSHKeyWithoutComment),
				),
			},
		},
	})
}
