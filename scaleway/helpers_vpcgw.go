package scaleway

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"golang.org/x/exp/slices"
)

const (
	defaultVPCGatewayTimeout                   = 10 * time.Minute
	defaultVPCGatewayRetry                     = 5 * time.Second
	defaultVPCPublicGatewayIPReverseDNSTimeout = 5 * time.Minute
)

// vpcgwAPIWithZone returns a new VPC API and the zone for a Create request
func vpcgwAPIWithZone(d *schema.ResourceData, m interface{}) (*vpcgw.API, scw.Zone, error) {
	meta := m.(*Meta)
	vpcgwAPI := vpcgw.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return vpcgwAPI, zone, nil
}

// vpcgwAPIWithZoneAndID
func vpcgwAPIWithZoneAndID(m interface{}, id string) (*vpcgw.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	vpcgwAPI := vpcgw.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	if err != nil {
		return nil, "", "", err
	}
	return vpcgwAPI, zone, ID, nil
}

func waitForVPCPublicGateway(ctx context.Context, api *vpcgw.API, zone scw.Zone, id string, timeout time.Duration) (*vpcgw.Gateway, error) {
	retryInterval := defaultVPCGatewayRetry
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	gateway, err := api.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		Timeout:       scw.TimeDurationPtr(timeout),
		GatewayID:     id,
		RetryInterval: &retryInterval,
		Zone:          zone,
	}, scw.WithContext(ctx))

	return gateway, err
}

func waitForVPCGatewayNetwork(ctx context.Context, api *vpcgw.API, zone scw.Zone, id string, timeout time.Duration) (*vpcgw.GatewayNetwork, error) {
	retryIntervalGWNetwork := defaultVPCGatewayRetry
	if DefaultWaitRetryInterval != nil {
		retryIntervalGWNetwork = *DefaultWaitRetryInterval
	}

	gatewayNetwork, err := api.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: id,
		Timeout:          scw.TimeDurationPtr(timeout),
		RetryInterval:    &retryIntervalGWNetwork,
		Zone:             zone,
	}, scw.WithContext(ctx))

	return gatewayNetwork, err
}

func waitForDHCPEntries(ctx context.Context, api *vpcgw.API, zone scw.Zone, gatewayID string, macAddress string, timeout time.Duration) (*vpcgw.ListDHCPEntriesResponse, error) {
	retryIntervalDHCPEntries := defaultVPCGatewayRetry
	if DefaultWaitRetryInterval != nil {
		retryIntervalDHCPEntries = *DefaultWaitRetryInterval
	}

	req := &vpcgw.WaitForDHCPEntriesRequest{
		MacAddress:    macAddress,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryIntervalDHCPEntries,
	}

	if gatewayID != "" {
		req.GatewayNetworkID = &gatewayID
	}

	dhcpEntries, err := api.WaitForDHCPEntries(req, scw.WithContext(ctx))
	return dhcpEntries, err
}

func isGatewayIPReverseResolved(ctx context.Context, api *vpcgw.API, reverse string, timeout time.Duration, id string, zone scw.Zone) bool {
	ticker := time.Tick(time.Millisecond * 500)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	getIPReq := &vpcgw.GetIPRequest{
		Zone: zone,
		IPID: id,
	}
	IP, err := api.GetIP(getIPReq, scw.WithContext(ctx))
	if err != nil {
		return false
	}

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			conn, err := d.DialContext(ctx, network, "ns0.dom.scw.cloud:53")
			if err != nil {
				conn, err = d.DialContext(ctx, network, "ns1.dom.scw.cloud:53")
			}
			return conn, err
		},
	}

	for {
		select {
		case <-ticker:
			address, err := r.LookupHost(ctx, reverse)
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return false
				}
			} else if slices.Contains(address, IP.Address.String()) {
				return true
			}
		case <-ctx.Done():
			return false
		}
	}
}

func findDefaultReverse(address string) string {
	parts := strings.Split(address, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "-") + ".instances.scw.cloud"
}
