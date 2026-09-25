package domain

import (
	"context"
	"time"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultDomainRecordTimeout       = 5 * time.Minute
	defaultDomainZoneTimeout         = 5 * time.Minute
	defaultDomainZoneRetryInterval   = 5 * time.Second
	defaultDomainRegistrationTimeout = 30 * time.Minute
)

func waitForDNSZone(ctx context.Context, domainAPI *domain.API, dnsZone string, timeout time.Duration) (*domain.DNSZone, error) {
	retryInterval := defaultDomainZoneRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var zone *domain.DNSZone

	err := transport.RetryOn403(ctx, func() error {
		var err error

		zone, err = domainAPI.WaitForDNSZone(&domain.WaitForDNSZoneRequest{
			DNSZone:       dnsZone,
			Timeout:       new(timeout),
			RetryInterval: new(retryInterval),
		}, scw.WithContext(ctx))

		return err
	})

	return zone, err
}

func waitForDNSRecordExist(ctx context.Context, domainAPI *domain.API, dnsZone, recordName string, recordType domain.RecordType, timeout time.Duration) (*domain.Record, error) {
	retryInterval := defaultDomainZoneRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var record *domain.Record

	err := transport.RetryOn403(ctx, func() error {
		var err error

		record, err = domainAPI.WaitForDNSRecordExist(&domain.WaitForDNSRecordExistRequest{
			DNSZone:       dnsZone,
			RecordName:    recordName,
			RecordType:    recordType,
			Timeout:       new(timeout),
			RetryInterval: new(retryInterval),
		}, scw.WithContext(ctx))

		return err
	})

	return record, err
}

func waitForDomainsRegistration(ctx context.Context, api *domain.RegistrarAPI, domainName string, timeout time.Duration) (*domain.Domain, error) {
	retryInterval := defaultWaitDomainsRegistrationRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var registeredDomain *domain.Domain

	err := transport.RetryOn403(ctx, func() error {
		var err error

		registeredDomain, err = api.WaitForOrderDomain(&domain.WaitForOrderDomainRequest{
			Domain:        domainName,
			Timeout:       new(timeout),
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return registeredDomain, err
}

func waitForAutoRenewStatus(ctx context.Context, api *domain.RegistrarAPI, domainName string, timeout time.Duration) (*domain.Domain, error) {
	retryInterval := defaultWaitDomainsRegistrationRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var registeredDomain *domain.Domain

	err := transport.RetryOn403(ctx, func() error {
		var err error

		registeredDomain, err = api.WaitForAutoRenewStatus(&domain.WaitForAutoRenewStatusRequest{
			Domain:        domainName,
			Timeout:       new(timeout),
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return registeredDomain, err
}

func waitForDNSSECStatus(ctx context.Context, api *domain.RegistrarAPI, domainName string, timeout time.Duration) (*domain.Domain, error) {
	retryInterval := defaultWaitDomainsRegistrationRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	var registeredDomain *domain.Domain

	err := transport.RetryOn403(ctx, func() error {
		var err error

		registeredDomain, err = api.WaitForDNSSECStatus(&domain.WaitForDNSSECStatusRequest{
			Domain:        domainName,
			Timeout:       new(timeout),
			RetryInterval: &retryInterval,
		}, scw.WithContext(ctx))

		return err
	})

	return registeredDomain, err
}
