package mailboxtestfuncs

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

// CreateTestDomain creates a mailbox domain via the API for mailbox resource tests.
// Soft-deleted mailboxes block domain deletion, so the domain is not managed by Terraform.
func CreateTestDomain(tt *acctest.TestTools, name string) string {
	tt.T.Helper()

	api := mailboxsdk.NewAPI(tt.Meta.ScwClient())

	domain, err := api.CreateDomain(&mailboxsdk.CreateDomainRequest{
		ProjectID: TestProjectID(tt),
		Name:      name,
	}, scw.WithContext(tt.T.Context()))
	if err != nil {
		tt.T.Fatalf("failed to create test mailbox domain %q: %v", name, err)
	}

	retryInterval := 5 * time.Second
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	timeout := 5 * time.Minute

	domain, err = api.WaitForDomain(&mailboxsdk.WaitForDomainRequest{
		DomainID:      domain.ID,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(tt.T.Context()))
	if err != nil {
		tt.T.Fatalf("failed waiting for test mailbox domain %q: %v", name, err)
	}

	// Soft-deleted mailboxes block domain deletion until the API purge window
	// elapses, so cleanup is best-effort via the sweeper rather than t.Cleanup.
	return domain.ID
}

// CheckMailboxDestroyed verifies that all mailbox resources in state have been deleted
// or soft-deleted (deletion_scheduled).
func CheckMailboxDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		api := mailboxsdk.NewAPI(tt.Meta.ScwClient())

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "scaleway_mailbox_mailbox" {
				continue
			}

			mb, err := api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: rs.Primary.ID}, scw.WithContext(tt.T.Context()))
			if err != nil {
				if httperrors.Is404(err) {
					continue
				}

				return fmt.Errorf("unexpected error checking mailbox %s: %w", rs.Primary.ID, err)
			}

			if mb.Status != mailboxsdk.MailboxStatusDeletionScheduled {
				return fmt.Errorf("mailbox %s still exists with status %s", rs.Primary.ID, mb.Status)
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

			_, err := api.GetDomain(&mailboxsdk.GetDomainRequest{DomainID: rs.Primary.ID}, scw.WithContext(tt.T.Context()))
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

		_, err := api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: rs.Primary.ID}, scw.WithContext(tt.T.Context()))
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

		_, err := api.GetDomain(&mailboxsdk.GetDomainRequest{DomainID: rs.Primary.ID}, scw.WithContext(tt.T.Context()))
		if err != nil {
			return fmt.Errorf("error reading domain %s: %w", rs.Primary.ID, err)
		}

		return nil
	}
}
