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

	const domainName = "tf-tests-mailbox-basic.example.com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_mailbox_domain" "basic" {
					  name = %q
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckDomainExists(tt, "scaleway_mailbox_domain.basic"),
					resource.TestCheckResourceAttr("scaleway_mailbox_domain.basic", "name", domainName),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_domain.basic", "status"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_domain.basic", "project_id"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_domain.basic", "created_at"),
					acctest.CheckResourceAttrUUID("scaleway_mailbox_domain.basic", "id"),
				),
			},
			{
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

	// Fixed project_id must match the VCR cassette response body; CI's default
	// project differs and would cause "inconsistent result after apply".
	projectID := "46fd79d8-1a35-4548-bfb8-03df51a0ebae"
	if *acctest.UpdateCassettes {
		var ok bool

		projectID, ok = tt.Meta.ScwClient().GetDefaultProjectID()
		if !ok {
			t.Skip("default project ID is required to record this test")
		}
	}

	const domainName = "tf-tests-mailbox-proj.example.com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_mailbox_domain" "with_project" {
					  name       = %q
					  project_id = %q
					}
				`, domainName, projectID),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckDomainExists(tt, "scaleway_mailbox_domain.with_project"),
					resource.TestCheckResourceAttr("scaleway_mailbox_domain.with_project", "project_id", projectID),
					resource.TestCheckResourceAttr("scaleway_mailbox_domain.with_project", "name", domainName),
				),
			},
		},
	})
}

func TestAccMailboxDomain_DNSRecordsExposed(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	const domainName = "tf-tests-mailbox-dns.example.com"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             mailboxtestfuncs.CheckDomainDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_mailbox_domain" "dns" {
					  name = %q
					}
				`, domainName),
				Check: resource.ComposeTestCheckFunc(
					mailboxtestfuncs.CheckDomainExists(tt, "scaleway_mailbox_domain.dns"),
					resource.TestCheckResourceAttrSet("scaleway_mailbox_domain.dns", "dns_records.#"),
				),
			},
		},
	})
}
