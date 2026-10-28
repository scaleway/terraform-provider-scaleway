package mailbox

import (
	"context"
	"fmt"
	"time"

	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForDomain(ctx context.Context, api *mailboxsdk.API, domainID string, timeout time.Duration) (*mailboxsdk.Domain, error) {
	retryInterval := defaultRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	domain, err := api.WaitForDomain(&mailboxsdk.WaitForDomainRequest{
		DomainID:      domainID,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for mailbox domain to be ready: %w", err)
	}

	return domain, nil
}

// waitForMailbox waits until the mailbox leaves provisioning states.
// waiting_domain is stable when the domain is not ready yet. If the domain is
// already ready, keep waiting so the mailbox can catch up to ready.
func waitForMailbox(ctx context.Context, api *mailboxsdk.API, mailboxID string, timeout time.Duration) (*mailboxsdk.Mailbox, error) {
	retryInterval := defaultRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	deadline := time.Now().Add(timeout)

	for {
		mb, err := api.GetMailbox(&mailboxsdk.GetMailboxRequest{MailboxID: mailboxID}, scw.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("getting mailbox: %w", err)
		}

		keepWaiting := false

		switch mb.Status {
		case mailboxsdk.MailboxStatusCreating,
			mailboxsdk.MailboxStatusWaitingPayment,
			mailboxsdk.MailboxStatusRenewing,
			mailboxsdk.MailboxStatusRestoring:
			keepWaiting = true
		case mailboxsdk.MailboxStatusWaitingDomain:
			domain, domainErr := api.GetDomain(&mailboxsdk.GetDomainRequest{DomainID: mb.DomainID}, scw.WithContext(ctx))
			if domainErr != nil {
				return nil, fmt.Errorf("getting mailbox domain: %w", domainErr)
			}

			if domain.Status == mailboxsdk.DomainStatusReady {
				keepWaiting = true
			}
		}

		if !keepWaiting {
			switch mb.Status {
			case mailboxsdk.MailboxStatusPaymentFailed,
				mailboxsdk.MailboxStatusLocked,
				mailboxsdk.MailboxStatusDeleting,
				mailboxsdk.MailboxStatusDeletionScheduled:
				return nil, fmt.Errorf("mailbox %s ended in unexpected status %s", mailboxID, mb.Status)
			}

			return mb, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for mailbox %s (last status: %s)", mailboxID, mb.Status)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for mailbox: %w", ctx.Err())
		case <-time.After(retryInterval):
		}
	}
}

// waitForMailboxDeleted waits until the mailbox is gone or soft-deleted (deletion_scheduled).
func waitForMailboxDeleted(ctx context.Context, api *mailboxsdk.API, mailboxID string, timeout time.Duration) (*mailboxsdk.Mailbox, error) {
	retryInterval := defaultRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	mailbox, err := api.WaitForMailbox(&mailboxsdk.WaitForMailboxRequest{
		MailboxID:     mailboxID,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for mailbox to be deleted: %w", err)
	}

	return mailbox, nil
}

func waitForAlias(ctx context.Context, api *mailboxsdk.API, aliasID string, timeout time.Duration) (*mailboxsdk.Alias, error) {
	retryInterval := defaultRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	alias, err := api.WaitForAlias(&mailboxsdk.WaitForAliasRequest{
		AliasID:       aliasID,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("waiting for mailbox alias to be ready: %w", err)
	}

	return alias, nil
}

// waitForAliasDeleted waits until the alias is gone (GetAlias returns 404).
func waitForAliasDeleted(ctx context.Context, api *mailboxsdk.API, aliasID string, timeout time.Duration) error {
	_, err := waitForAlias(ctx, api, aliasID, timeout)
	if err != nil && httperrors.Is404(err) {
		return nil
	}

	return err
}
