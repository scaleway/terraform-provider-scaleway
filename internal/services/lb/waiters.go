package lb

import (
	"context"
	"time"

	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

func waitForLB(ctx context.Context, lbAPI *lb.ZonedAPI, zone scw.Zone, lbID string, timeout time.Duration) (*lb.LB, error) {
	retryInterval := DefaultWaitLBRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	loadBalancer, err := lbAPI.WaitForLb(&lb.ZonedAPIWaitForLBRequest{
		LBID:          lbID,
		Zone:          zone,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return loadBalancer, err
}

func waitForInstances(ctx context.Context, lbAPI *lb.ZonedAPI, zone scw.Zone, lbID string, timeout time.Duration) (*lb.LB, error) {
	retryInterval := DefaultWaitLBRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	loadBalancer, err := lbAPI.WaitForLbInstances(&lb.ZonedAPIWaitForLBInstancesRequest{
		Zone:          zone,
		LBID:          lbID,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return loadBalancer, err
}

func waitForPrivateNetworks(ctx context.Context, lbAPI *lb.ZonedAPI, zone scw.Zone, lbID string, timeout time.Duration) ([]*lb.PrivateNetwork, error) {
	retryInterval := DefaultWaitLBRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	privateNetworks, err := lbAPI.WaitForLBPN(&lb.ZonedAPIWaitForLBPNRequest{
		LBID:          lbID,
		Zone:          zone,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return privateNetworks, err
}

func waitForCertificate(ctx context.Context, lbAPI *lb.ZonedAPI, zone scw.Zone, id string, timeout time.Duration) (*lb.Certificate, error) {
	retryInterval := DefaultWaitLBRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	certificate, err := lbAPI.WaitForLBCertificate(&lb.ZonedAPIWaitForLBCertificateRequest{
		CertID:        id,
		Zone:          zone,
		Timeout:       new(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	return certificate, err
}
