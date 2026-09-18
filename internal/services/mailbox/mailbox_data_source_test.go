package mailbox_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	mailboxtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mailbox/testfuncs"
)

func TestAccDataSourceMailboxMailbox_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mailbox-dsid")+".example.com")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "ds_by_id" {
  domain_id           = %q
  local_part          = "datasource.byid"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}

data "scaleway_mailbox_mailbox" "by_id" {
  mailbox_id = scaleway_mailbox_mailbox.ds_by_id.id
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_mailbox_mailbox.by_id", "id",
						"scaleway_mailbox_mailbox.ds_by_id", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_mailbox_mailbox.by_id", "email",
						"scaleway_mailbox_mailbox.ds_by_id", "email",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_mailbox_mailbox.by_id", "status",
						"scaleway_mailbox_mailbox.ds_by_id", "status",
					),
				),
			},
		},
	})
}

func TestAccDataSourceMailboxMailbox_ByEmail(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	domainID := mailboxtestfuncs.CreateTestDomain(tt, sdkacctest.RandomWithPrefix("tf-tests-mailbox-dsem")+".example.com")

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckMailboxDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_mailbox_mailbox" "ds_by_email" {
  domain_id           = %q
  local_part          = "datasource.byemail"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}

data "scaleway_mailbox_mailbox" "by_email" {
  email = scaleway_mailbox_mailbox.ds_by_email.email
}
`, domainID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.scaleway_mailbox_mailbox.by_email", "id",
						"scaleway_mailbox_mailbox.ds_by_email", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.scaleway_mailbox_mailbox.by_email", "domain_id",
						"scaleway_mailbox_mailbox.ds_by_email", "domain_id",
					),
				),
			},
		},
	})
}
