package mailbox_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	mailboxtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mailbox/testfuncs"
)

const testDomainName = "terraform-test.example.com"

// testConfigDomain returns a domain resource config reused across mailbox tests.
func testConfigDomain(name string) string {
	return `
resource "scaleway_mailbox_domain" "domain" {
  name = "` + name + `"
}
`
}

func TestAccMailboxMailbox_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "basic" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "john.doe"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`,
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox_mailbox.basic"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.basic", "local_part", "john.doe"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.basic", "subscription_period", "monthly"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_mailbox.basic", "email"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_mailbox.basic", "status"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_mailbox.basic", "created_at"),
					resource.TestCheckResourceAttrPair(
						"scaleway_mailbox_mailbox.basic", "domain_id",
						"scaleway_mailbox_domain.domain", "id",
					),
					acctest.CheckResourceAttrUUID("scaleway_mailbox_mailbox.basic", "id"),
				),
			},
			{
				// Verify import by ID
				ResourceName:            "scaleway_mailbox_mailbox.basic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccMailboxMailbox_WithOptionalFields(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "full" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "jane.doe"
  password            = "S3cur3P@ssw0rd!"
  display_name        = "Jane Doe"
  recovery_email      = "recovery@external.example.com"
  subscription_period = "monthly"
}
`,
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox_mailbox.full"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.full", "display_name", "Jane Doe"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.full", "recovery_email", "recovery@external.example.com"),
				),
			},
		},
	})
}

func TestAccMailboxMailbox_Update(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "update" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "update.test"
  password            = "S3cur3P@ssw0rd!"
  display_name        = "Update Test"
  subscription_period = "monthly"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.update", "display_name", "Update Test"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.update", "subscription_period", "monthly"),
				),
			},
			{
				// Update display_name and recovery_email in-place
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "update" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "update.test"
  password            = "S3cur3P@ssw0rd!"
  display_name        = "Updated Name"
  recovery_email      = "new-recovery@external.example.com"
  subscription_period = "monthly"
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.update", "display_name", "Updated Name"),
					resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.update", "recovery_email", "new-recovery@external.example.com"),
				),
			},
		},
	})
}

func TestAccMailboxMailbox_PasswordChange(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "pwd" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "pwd.change"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`,
				Check: mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox_mailbox.pwd"),
			},
			{
				// Password change should trigger an in-place update, not a recreation
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "pwd" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "pwd.change"
  password            = "N3wS3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`,
				Check: mailboxtestfuncs.CheckMailboxExists(tt, "scaleway_mailbox_mailbox.pwd"),
			},
		},
	})
}

func TestAccMailboxMailbox_ForceNewOnLocalPartChange(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "force_new" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "original"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`,
				Check: resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.force_new", "local_part", "original"),
			},
			{
				// Changing local_part must force a new resource
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "force_new" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "renamed"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}
`,
				Check: resource.TestCheckResourceAttr("scaleway_mailbox_mailbox.force_new", "local_part", "renamed"),
			},
		},
	})
}
