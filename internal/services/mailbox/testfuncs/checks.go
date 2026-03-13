package mailboxtestfuncs

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
)

// CheckMailboxDestroyed verifies that all mailbox resources in state have been deleted.
func CheckMailboxDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		api := mailboxsdk.NewAPI(tt.Meta.ScwClient())

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mailbox_mailbox" {
				continue
			}

			_, err := api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: rs.Primary.ID})
			if err == nil {
				return fmt.Errorf("mailbox %s still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return fmt.Errorf("unexpected error checking mailbox %s: %w", rs.Primary.ID, err)
			}
		}

		return nil
	}
}

// CheckDomainDestroyed verifies that all mailbox domain resources in state have been deleted.
func CheckDomainDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		api := mailboxsdk.NewAPI(tt.Meta.ScwClient())

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mailbox_domain" {
				continue
			}

			_, err := api.GetDomain(&mailboxsdk.GetDomainRequest{DomainID: rs.Primary.ID})
			if err == nil {
				return fmt.Errorf("mailbox domain %s still exists", rs.Primary.ID)
			}

			if !httperrors.Is404(err) {
				return fmt.Errorf("unexpected error checking domain %s: %w", rs.Primary.ID, err)
			}
		}

		return nil
	}
}

// CheckMailboxExists verifies a mailbox resource exists in both state and API.
func CheckMailboxExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource %q not found in state", n)
		}

		api := mailboxsdk.NewAPI(tt.Meta.ScwClient())

		_, err := api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: rs.Primary.ID})
		if err != nil {
			return fmt.Errorf("error reading mailbox %s: %w", rs.Primary.ID, err)
		}

		return nil
	}
}

// CheckDomainExists verifies a domain resource exists in both state and API.
func CheckDomainExists(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource %q not found in state", n)
		}

		api := mailboxsdk.NewAPI(tt.Meta.ScwClient())

		_, err := api.GetDomain(&mailboxsdk.GetDomainRequest{DomainID: rs.Primary.ID})
		if err != nil {
			return fmt.Errorf("error reading domain %s: %w", rs.Primary.ID, err)
		}

		return nil
	}
}
