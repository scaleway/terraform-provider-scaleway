package mailbox_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	mailboxtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mailbox/testfuncs"
)

func TestAccMailboxForwarding_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mbx-fwd"))
	forwardingID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			mailboxtestfuncs.CheckForwardingDestroyed(tt),
			mailboxtestfuncs.CheckMailboxDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox" "main" {
  domain_id           = %q
  local_part          = "fwd.owner"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}

resource "scaleway_mailbox_forwarding" "main" {
  mailbox_id = scaleway_mailbox.main.id
  email      = "backup@example.com"
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckForwardingExists(tt, "scaleway_mailbox_forwarding.main"),
					resource.TestCheckResourceAttr("scaleway_mailbox_forwarding.main", "email", "backup@example.com"),
					resource.TestCheckResourceAttr("scaleway_mailbox_forwarding.main", "status", "ready"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_forwarding.main", "created_at"),
					acctest.CheckResourceAttrUUID("scaleway_mailbox_forwarding.main", "id"),
					acctest.CheckResourceIDPersisted("scaleway_mailbox_forwarding.main", &forwardingID),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox" "main" {
  domain_id           = %q
  local_part          = "fwd.owner"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}

resource "scaleway_mailbox_forwarding" "main" {
  mailbox_id = scaleway_mailbox.main.id
  email      = "ops@example.com"
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckForwardingExists(tt, "scaleway_mailbox_forwarding.main"),
					resource.TestCheckResourceAttr("scaleway_mailbox_forwarding.main", "email", "ops@example.com"),
					resource.TestCheckResourceAttr("scaleway_mailbox_forwarding.main", "status", "ready"),
					acctest.CheckResourceIDPersisted("scaleway_mailbox_forwarding.main", &forwardingID),
				),
			},
			{
				ResourceName:      "scaleway_mailbox_forwarding.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMailbox_ForwardingSettings(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mbx-fwdset"))

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox" "main" {
  domain_id             = %q
  local_part            = "fwd.settings"
  password              = "S3cur3P@ssw0rd!"
  subscription_period   = "monthly"
  forwarding_enabled    = true
  forwarding_keep_copy  = true
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox.main"),
					resource.TestCheckResourceAttr("scaleway_mailbox.main", "forwarding_enabled", "true"),
					resource.TestCheckResourceAttr("scaleway_mailbox.main", "forwarding_keep_copy", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox" "main" {
  domain_id             = %q
  local_part            = "fwd.settings"
  password              = "S3cur3P@ssw0rd!"
  subscription_period   = "monthly"
  forwarding_enabled    = false
  forwarding_keep_copy  = true
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox.main"),
					resource.TestCheckResourceAttr("scaleway_mailbox.main", "forwarding_enabled", "false"),
					resource.TestCheckResourceAttr("scaleway_mailbox.main", "forwarding_keep_copy", "true"),
				),
			},
		},
	})
}
