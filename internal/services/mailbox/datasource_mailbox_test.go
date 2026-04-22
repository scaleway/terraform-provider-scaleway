package mailbox_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceMailboxMailbox_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "ds_by_id" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "datasource.byid"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}

data "scaleway_mailbox_mailbox" "by_id" {
  mailbox_id = scaleway_mailbox_mailbox.ds_by_id.id
}
`,
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

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testConfigDomain(testDomainName) + `
resource "scaleway_mailbox_mailbox" "ds_by_email" {
  domain_id           = scaleway_mailbox_domain.domain.id
  local_part          = "datasource.byemail"
  password            = "S3cur3P@ssw0rd!"
  subscription_period = "monthly"
}

data "scaleway_mailbox_mailbox" "by_email" {
  email = scaleway_mailbox_mailbox.ds_by_email.email
}
`,
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
