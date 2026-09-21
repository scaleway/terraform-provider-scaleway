package mailbox_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	mailboxtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mailbox/testfuncs"
)

func TestAccMailboxMailbox_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mbx-basic"))

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "basic" {
  domain_id           = %q
  local_part          = "john.doe"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox_mailbox.basic"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.basic", "local_part", "john.doe"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.basic", "subscription_period", "monthly"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.basic", "domain_id", domainID),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.basic", "status", "ready"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_mailbox.basic", "email"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_mailbox.basic", "created_at"),
					acctest.CheckResourceAttrUUID("scaleway_mailbox_mailbox.basic", "id"),
				),
			},
			{
				ResourceName:            "scaleway_mailbox_mailbox.basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "password_wo", "password_wo_version"},
			},
		},
	})
}

func TestAccMailboxMailbox_PasswordChange(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mbx-pwd"))

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "pwd" {
  domain_id           = %q
  local_part          = "pwd.change"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox_mailbox.pwd"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.pwd", "status", "ready"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "pwd" {
  domain_id           = %q
  local_part          = "pwd.change"
  password            = "N3wS3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`, domainID),
				Check: mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox_mailbox.pwd"),
			},
		},
	})
}

func TestAccMailboxMailbox_PasswordWO(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mbx-pwo"))

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "pwd_wo" {
  domain_id           = %q
  local_part          = "pwd.wo"
  password_wo         = "S3cur3P@ssw0rd!"
  password_wo_version = 1
  subscription_period = "monthly"
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox_mailbox.pwd_wo"),
					resource.TestCheckNoResourceAttr("scaleway_mailbox_mailbox.pwd_wo", "password_wo"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.pwd_wo", "password_wo_version", "1"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.pwd_wo", "status", "ready"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "pwd_wo" {
  domain_id           = %q
  local_part          = "pwd.wo"
  password_wo         = "N3wS3cur3P@ssw0rd!"
  password_wo_version = 2
  subscription_period = "monthly"
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("scaleway_mailbox_mailbox.pwd_wo", "password_wo"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.pwd_wo", "password_wo_version", "2"),
				),
			},
		},
	})
}

func TestAccMailboxMailbox_ForceNewOnLocalPartChange(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mbx-fn"))

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "force_new" {
  domain_id           = %q
  local_part          = "original"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`, domainID),
				Check: resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.force_new", "local_part", "original"),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "force_new" {
  domain_id           = %q
  local_part          = "renamed"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`, domainID),
				Check: resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.force_new", "local_part", "renamed"),
			},
		},
	})
}
