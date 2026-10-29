package mailbox_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	mailboxtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mailbox/testfuncs"
)

func TestAccMailboxAlias_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	t.Cleanup(func() { _ = acctest.AnonymizeCassetteForTest(t, "") })

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mbx-alias"))
	aliasID := ""

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			mailboxtestfuncs.CheckAliasDestroyed(tt),
			mailboxtestfuncs.CheckMailboxDestroyed(tt),
		),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox" "main" {
  domain_id           = %q
  local_part          = "alias.owner"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}

resource "scaleway_mailbox_alias" "main" {
  mailbox_id  = scaleway_mailbox.main.id
  local_part  = "contact"
  description = "Public contact address"
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckAliasExists(tt, "scaleway_mailbox_alias.main"),
					resource.TestCheckResourceAttr("scaleway_mailbox_alias.main", "local_part", "contact"),
					resource.TestCheckResourceAttr("scaleway_mailbox_alias.main", "description", "Public contact address"),
					resource.TestCheckResourceAttr("scaleway_mailbox_alias.main", "status", "ready"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_alias.main", "email"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_alias.main", "created_at"),
					acctest.CheckResourceAttrUUID("scaleway_mailbox_alias.main", "id"),
					acctest.CheckResourceIDPersisted("scaleway_mailbox_alias.main", &aliasID),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox" "main" {
  domain_id           = %q
  local_part          = "alias.owner"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}

resource "scaleway_mailbox_alias" "main" {
  mailbox_id  = scaleway_mailbox.main.id
  local_part  = "contact"
  description = "Updated contact address"
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckAliasExists(tt, "scaleway_mailbox_alias.main"),
					resource.TestCheckResourceAttr("scaleway_mailbox_alias.main", "description", "Updated contact address"),
					resource.TestCheckResourceAttr("scaleway_mailbox_alias.main", "status", "ready"),
					acctest.CheckResourceIDPersisted("scaleway_mailbox_alias.main", &aliasID),
				),
			},
			{
				ResourceName:      "scaleway_mailbox_alias.main",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
