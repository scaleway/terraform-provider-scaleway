package scaleway

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScalewayDataSourceAccountSSHKey_Basic(t *testing.T) {
	sshKeyName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayAccountSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceScalewayAccountSSHKeyConfig(sshKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("data.scaleway_account_ssh_key.prod"),
					resource.TestCheckResourceAttr("data.scaleway_account_ssh_key.prod", "name", sshKeyName),
					resource.TestCheckResourceAttr("data.scaleway_account_ssh_key.prod", "public_key", dataSourceAccountSSHKey),
					testAccCheckScalewayAccountSSHKeyExists("data.scaleway_account_ssh_key.stg"),
					resource.TestCheckResourceAttr("data.scaleway_account_ssh_key.stg", "name", sshKeyName),
					resource.TestCheckResourceAttr("data.scaleway_account_ssh_key.stg", "public_key", dataSourceAccountSSHKey),
				),
			},
		},
	})
}

const dataSourceAccountSSHKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDCtMlb2pIlcCC3Exui1z77NfFWtlf59P/L38OwfHCKnxgHZfcjWZZEOdJPxeA6/iRCUrJo+mUKOedNJTCR9Wgg5zhUM1dd/fiyCS+STLf9fzoA1evPEEIHp0iobIpVQ0pGIdNqipL2n8BcV7f1oC2AkELCUp4gogkeUkKtK71DtsjGJQJBRlg01U2gTAnU0Q3kf5wxhCIELke9J3eblTpvdnNonqpXsQuy+InpT51NDtMbBdgkwNsgw6wDxI64NY92mEXO7PK/uhSAmxIM9a4evLhgFxSr8vFNTwqq5fbSyDeeQwHG23U0CmhM0tOwLuJAbQotWrfYsLCpinnBb+sp opensource@scaleway.com"

func testAccDataSourceScalewayAccountSSHKeyConfig(name string) string {
	return `
resource "scaleway_account_ssh_key" "main" {
	name 	   = "` + name + `"
	public_key = "` + dataSourceAccountSSHKey + `"
}

data "scaleway_account_ssh_key" "prod" {
	name = "${scaleway_account_ssh_key.main.name}"
}

data "scaleway_account_ssh_key" "stg" {
	ssh_key_id = "${scaleway_account_ssh_key.main.id}"
}`
}
