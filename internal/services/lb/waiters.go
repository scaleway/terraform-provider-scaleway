package lb

import (
	"context"
	"fmt"
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
	if err != nil {
		return nil, fmt.Errorf("waiting for load balancer to be ready: %w", err)
	}

	return loadBalancer, nil
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
	if err != nil {
		return nil, fmt.Errorf("waiting for load balancer instances to be ready: %w", err)
	}

	return loadBalancer, nil
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
	if err != nil {
		return nil, fmt.Errorf("waiting for load balancer private networks to be ready: %w", err)
	}

	return privateNetworks, nil
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
	if err != nil {
		return nil, fmt.Errorf("waiting for load balancer certificate to be ready: %w", err)
	}

	return certificate, nil
}
