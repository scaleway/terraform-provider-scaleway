package mailboxtestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
)

// AddTestSweepers registers sweepers that clean up test resources.
func AddTestSweepers() {
	resource.AddTestSweepers("scaleway_mailbox_mailbox", &resource.Sweeper{
		Name: "scaleway_mailbox_mailbox",
		F:    testSweepMailboxes,
	})

	resource.AddTestSweepers("scaleway_mailbox_domain", &resource.Sweeper{
		Name:         "scaleway_mailbox_domain",
		F:            testSweepDomains,
		Dependencies: []string{"scaleway_mailbox_mailbox"},
	})
}

func testSweepMailboxes(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := mailboxsdk.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting test mailboxes")

		resp, err := api.ListMailboxes(&mailboxsdk.ListMailboxesRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing mailboxes in sweeper: %w", err)
		}

		for _, mb := range resp.Mailboxes {
			_, err := api.DeleteMailbox(&mailboxsdk.DeleteMailboxRequest{MailboxID: mb.ID})
			if err != nil {
				logging.L.Debugf("sweeper: error deleting mailbox %s: %s", mb.ID, err)
			}
		}

		return nil
	})
}

func testSweepDomains(_ string) error {
	return acctest.Sweep(func(scwClient *scw.Client) error {
		api := mailboxsdk.NewAPI(scwClient)

		logging.L.Debugf("sweeper: deleting test mailbox domains")

		resp, err := api.ListDomains(&mailboxsdk.ListDomainsRequest{}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("error listing mailbox domains in sweeper: %w", err)
		}

		for _, domain := range resp.Domains {
			_, err := api.DeleteDomain(&mailboxsdk.DeleteDomainRequest{DomainID: domain.ID})
			if err != nil {
				logging.L.Debugf("sweeper: error deleting domain %s: %s", domain.ID, err)
			}
		}

		return nil
	})
}
