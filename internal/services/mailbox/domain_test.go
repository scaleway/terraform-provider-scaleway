package mailbox_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	mailboxtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/mailbox/testfuncs"
)

func TestAccMailboxDomain_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mailbox_domain" "basic" {
					  name = "terraform-test.example.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckDomainExists(tt, "scaleway_mailbox_domain.basic"),
					resource.TestCheckResourceAttr("scaleway_mailbox_domain.basic", "name", "terraform-test.example.com"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_domain.basic", "status"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_domain.basic", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_domain.basic", "created_at"),
					acctest.CheckResourceAttrUUID("scaleway_mailbox_domain.basic", "id"),
				),
			},
			{
				// Verify import by ID
				ResourceName:      "scaleway_mailbox_domain.basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMailboxDomain_WithProjectID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_account_project" "mailbox_proj" {
					  name = "tf-mailbox-test"
					}

					resource "scaleway_mailbox_domain" "with_project" {
					  name       = "terraform-test.example.com"
					  project_id = scaleway_account_project.mailbox_proj.id
					}
				`),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckDomainExists(tt, "scaleway_mailbox_domain.with_project"),
					resource.TestCheckResourceAttrPair(
						"scaleway_mailbox_domain.with_project", "project_id",
						"scaleway_account_project.mailbox_proj", "id",
					),
					resource.TestCheckResourceAttr("scaleway_mailbox_domain.with_project", "name", "terraform-test.example.com"),
				),
			},
		},
	})
}

func TestAccMailboxDomain_DNSRecordsExposed(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_mailbox_domain" "dns" {
					  name = "terraform-dns-test.example.com"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckDomainExists(tt, "scaleway_mailbox_domain.dns"),
					// DNS records should be populated after domain creation
					resource.TestCheckResourceAttrSet("scaleway_mailbox_domain.dns", "dns_records.#"),
				),
			},
		},
	})
}
