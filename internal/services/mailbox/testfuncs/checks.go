package mailboxtestfuncs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	domainsdk "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultTestDomainTimeout = 15 * time.Minute
	defaultDNSRecordTTL      = uint32(3600)
)

// CreateTestDomain creates a mailbox domain under acctest.TestDomain, applies the
// required DNS records on a dedicated zone, validates them, and waits until ready.
// Soft-deleted mailboxes block domain deletion, so the domain is not managed by Terraform.
func CreateTestDomain(tt *acctest.TestTools, subdomainPrefix string) string {
	tt.T.Helper()

	if acctest.TestDomain == "" {
		tt.T.Skip("TF_TEST_DOMAIN is required for mailbox acceptance tests")
	}

	ctx := tt.T.Context()
	projectID := TestProjectID(tt)
	subdomain := subdomainPrefix

	if subdomain == "" {
		subdomain = sdkacctest.RandomWithPrefix("tf-tests-mbx")
	}

	zoneName := subdomain + "." + acctest.TestDomain
	domainAPI := domainsdk.NewAPI(tt.Meta.ScwClient())
	mailboxAPI := mailboxsdk.NewAPI(tt.Meta.ScwClient())

	retryInterval := 5 * time.Second
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	timeout := defaultTestDomainTimeout

	dnsZone, err := domainAPI.CreateDNSZone(&domainsdk.CreateDNSZoneRequest{
		ProjectID: projectID,
		Domain:    acctest.TestDomain,
		Subdomain: subdomain,
	}, scw.WithContext(ctx))
	if err != nil {
		tt.T.Fatalf("failed to create DNS zone %q: %v", zoneName, err)
	}

	if dnsZone != nil && dnsZone.Domain != "" {
		if dnsZone.Subdomain == "" {
			zoneName = dnsZone.Domain
		} else {
			zoneName = dnsZone.Subdomain + "." + dnsZone.Domain
		}
	}

	cleanupZone := zoneName

	tt.T.Cleanup(func() {
		_, cleanupErr := domainAPI.DeleteDNSZone(&domainsdk.DeleteDNSZoneRequest{
			DNSZone: cleanupZone,
		}, scw.WithContext(tt.T.Context()))
		if cleanupErr != nil && !httperrors.Is404(cleanupErr) {
			tt.T.Logf("cleanup: failed to delete DNS zone %q: %v", cleanupZone, cleanupErr)
		}
	})

	_, err = domainAPI.WaitForDNSZone(&domainsdk.WaitForDNSZoneRequest{
		DNSZone:       zoneName,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		tt.T.Fatalf("failed waiting for DNS zone %q: %v", zoneName, err)
	}

	domain, err := mailboxAPI.CreateDomain(&mailboxsdk.CreateDomainRequest{
		ProjectID: projectID,
		Name:      zoneName,
	}, scw.WithContext(ctx))
	if err != nil {
		tt.T.Fatalf("failed to create test mailbox domain %q: %v", zoneName, err)
	}

	domain, err = mailboxAPI.WaitForDomain(&mailboxsdk.WaitForDomainRequest{
		DomainID:      domain.ID,
		Timeout:       &timeout,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		tt.T.Fatalf("failed waiting for test mailbox domain %q: %v", zoneName, err)
	}

	// On cassette replay the API response keeps the recorded FQDN; use it for
	// subsequent DNS updates so relative record names and URLs match the cassette.
	if domain.Name != "" {
		zoneName = domain.Name
	}

	records, err := mailboxAPI.GetDomainRecords(&mailboxsdk.GetDomainRecordsRequest{
		DomainID: domain.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		tt.T.Fatalf("failed to get DNS records for mailbox domain %q: %v", zoneName, err)
	}

	err = applyRequiredDNSRecords(ctx, domainAPI, zoneName, records)
	if err != nil {
		tt.T.Fatalf("failed to apply required DNS records for %q: %v", zoneName, err)
	}

	err = mailboxAPI.ValidateDomainRecords(&mailboxsdk.ValidateDomainRecordsRequest{
		DomainID: domain.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		tt.T.Fatalf("failed to validate DNS records for mailbox domain %q: %v", zoneName, err)
	}

	domain, err = waitForDomainReady(ctx, mailboxAPI, domain.ID, timeout, retryInterval)
	if err != nil {
		tt.T.Fatalf("failed waiting for mailbox domain %q to become ready: %v", zoneName, err)
	}

	if domain.Status != mailboxsdk.DomainStatusReady {
		tt.T.Fatalf("mailbox domain %q ended in status %s, want ready", zoneName, domain.Status)
	}

	// Soft-deleted mailboxes block domain deletion until the API purge window
	// elapses, so domain cleanup is best-effort via the sweeper rather than t.Cleanup.
	return domain.ID
}

func applyRequiredDNSRecords(
	ctx context.Context,
	domainAPI *domainsdk.API,
	zoneName string,
	records *mailboxsdk.GetDomainRecordsResponse,
) error {
	dnsRecords := make([]*domainsdk.Record, 0)

	for _, rec := range requiredMailboxDNSRecords(records) {
		dnsRec, err := mailboxRecordToDomainRecord(rec, zoneName)
		if err != nil {
			return err
		}

		dnsRecords = append(dnsRecords, dnsRec)
	}

	if len(dnsRecords) == 0 {
		return fmt.Errorf("no required DNS records returned for zone %s", zoneName)
	}

	_, err := domainAPI.UpdateDNSZoneRecords(&domainsdk.UpdateDNSZoneRecordsRequest{
		DNSZone: zoneName,
		Changes: []*domainsdk.RecordChange{
			{
				Add: &domainsdk.RecordChangeAdd{
					Records: dnsRecords,
				},
			},
		},
		ReturnAllRecords: new(false),
	}, scw.WithContext(ctx))

	return err
}

func requiredMailboxDNSRecords(resp *mailboxsdk.GetDomainRecordsResponse) []*mailboxsdk.DomainRecord {
	if resp == nil {
		return nil
	}

	candidates := []*mailboxsdk.DomainRecord{
		resp.DomainValidation,
		resp.Mx,
		resp.Dmarc,
		resp.Dkim,
		resp.Spf,
	}

	out := make([]*mailboxsdk.DomainRecord, 0, len(candidates))

	for _, rec := range candidates {
		if rec == nil {
			continue
		}

		if rec.Level != mailboxsdk.DomainRecordLevelRequired {
			continue
		}

		out = append(out, rec)
	}

	return out
}

func mailboxRecordToDomainRecord(rec *mailboxsdk.DomainRecord, zoneName string) (*domainsdk.Record, error) {
	recordType, err := mailboxDNSTypeToDomain(rec.DNSType)
	if err != nil {
		return nil, err
	}

	name := relativeDNSName(rec.DNSName, zoneName)
	data := strings.TrimSpace(rec.DNSValue)
	priority := uint32(0)

	if recordType == domainsdk.RecordTypeMX {
		parts := strings.Fields(data)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid MX value %q for %s", rec.DNSValue, rec.DNSName)
		}

		parsedPriority, parseErr := strconv.ParseUint(parts[0], 10, 32)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid MX priority in %q: %w", rec.DNSValue, parseErr)
		}

		priority = uint32(parsedPriority)
		data = parts[1]
	}

	return &domainsdk.Record{
		Name:     name,
		Type:     recordType,
		Data:     data,
		TTL:      defaultDNSRecordTTL,
		Priority: priority,
	}, nil
}

func mailboxDNSTypeToDomain(dnsType mailboxsdk.DomainRecordDNSType) (domainsdk.RecordType, error) {
	switch dnsType {
	case mailboxsdk.DomainRecordDNSTypeTxtDNSType:
		return domainsdk.RecordTypeTXT, nil
	case mailboxsdk.DomainRecordDNSTypeMxDNSType:
		return domainsdk.RecordTypeMX, nil
	case mailboxsdk.DomainRecordDNSTypeCnameDNSType:
		return domainsdk.RecordTypeCNAME, nil
	case mailboxsdk.DomainRecordDNSTypeSrvDNSType:
		return domainsdk.RecordTypeSRV, nil
	default:
		return "", fmt.Errorf("unsupported mailbox DNS type %s", dnsType)
	}
}

func relativeDNSName(dnsName, zoneName string) string {
	dnsName = strings.TrimSuffix(dnsName, ".")
	zoneName = strings.TrimSuffix(zoneName, ".")

	if strings.EqualFold(dnsName, zoneName) {
		return ""
	}

	suffix := "." + zoneName
	if before, ok := strings.CutSuffix(dnsName, suffix); ok {
		return before
	}

	return dnsName
}

func waitForDomainReady(
	ctx context.Context,
	api *mailboxsdk.API,
	domainID string,
	timeout time.Duration,
	retryInterval time.Duration,
) (*mailboxsdk.Domain, error) {
	deadline := time.Now().Add(timeout)

	var last *mailboxsdk.Domain

	for {
		domain, err := api.GetDomain(&mailboxsdk.GetDomainRequest{DomainID: domainID}, scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		last = domain

		switch domain.Status {
		case mailboxsdk.DomainStatusReady:
			return domain, nil
		case mailboxsdk.DomainStatusValidationFailed:
			return domain, fmt.Errorf("domain %s validation failed", domainID)
		case mailboxsdk.DomainStatusDeleting:
			return domain, fmt.Errorf("domain %s is deleting", domainID)
		case mailboxsdk.DomainStatusWaitingValidation:
			// DNS updates can lag behind the first validate-records call; retry
			// validation while we wait.
			_ = api.ValidateDomainRecords(&mailboxsdk.ValidateDomainRecordsRequest{
				DomainID: domainID,
			}, scw.WithContext(ctx))
		}

		if time.Now().After(deadline) {
			return last, fmt.Errorf("timeout waiting for domain %s to become ready (last status: %s)", domainID, domain.Status)
		}

		select {
		case <-ctx.Done():
			return last, ctx.Err()
		case <-time.After(retryInterval):
		}
	}
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
