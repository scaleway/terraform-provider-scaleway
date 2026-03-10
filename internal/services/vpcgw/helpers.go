package vpcgw

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	defaultTimeout             = 10 * time.Minute
	defaultRetry               = 5 * time.Second
	defaultIPReverseDNSTimeout = 10 * time.Minute
)

func newAPIWithZone(d *schema.ResourceData, m any) (*vpcgw.API, scw.Zone, error) {
	api := vpcgw.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, zone, nil
}

func NewAPIWithZoneAndID(m any, id string) (*vpcgw.API, scw.Zone, string, error) {
	api := vpcgw.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, zone, ID, nil
}

func retryUpdateGatewayReverseDNS(ctx context.Context, api *vpcgw.API, req *vpcgw.UpdateIPRequest, timeout time.Duration) error {
	timeoutChannel := time.After(timeout)

	for {
		select {
		case <-time.After(defaultRetry):
			_, err := api.UpdateIP(req, scw.WithContext(ctx))
			if err != nil && instance.IsIPReverseDNSResolveError(err) {
				continue
			}

			return err
		case <-timeoutChannel:
			_, err := api.UpdateIP(req, scw.WithContext(ctx))

			return err
		}
	}
}

func flattenIPNetList(ipNets []scw.IPNet) ([]string, error) {
	res := make([]string, 0, len(ipNets))

	for _, ipNet := range ipNets {
		flattened, err := types.FlattenIPNet(ipNet)
		if err != nil {
			return nil, err
		}

		res = append(res, flattened)
	}

	return res, nil
}

func expandIpamConfig(raw any) (bool, *string) {
	if raw == nil || len(raw.([]any)) != 1 {
		return false, nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	pushDefaultRoute := rawMap["push_default_route"].(bool)

	var ipamIPID *string

	if IPID, ok := rawMap["ipam_ip_id"].(string); ok && IPID != "" {
		ipamIPID = new(regional.ExpandID(IPID).ID)
	}

	return pushDefaultRoute, ipamIPID
}

func setPrivateIPs(ctx context.Context, d *schema.ResourceData, api *vpcgw.API, gn *vpcgw.GatewayNetwork, m any) diag.Diagnostics {
	resourceID := gn.ID

	region, err := gn.Zone.Region()
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Warning,
				Summary:  "Unable to get gateway network's private IPs",
				Detail:   err.Error(),
			},
		}
	}

	projectID, err := getGatewayProjectID(ctx, api, gn.Zone, gn.GatewayID)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Warning,
				Summary:  "Unable to get gateway network's private IPs",
				Detail:   err.Error(),
			},
		}
	}

	resourceType := ipamAPI.ResourceTypeVpcGatewayNetwork
	opts := &ipam.GetResourcePrivateIPsOptions{
		ResourceID:       &resourceID,
		ResourceType:     &resourceType,
		PrivateNetworkID: &gn.PrivateNetworkID,
		ProjectID:        &projectID,
	}

	privateIPs, err := ipam.GetResourcePrivateIPs(ctx, m, region, opts)

	switch {
	case err == nil:
		_ = d.Set("private_ip", privateIPs)
	case httperrors.Is403(err):
		return diag.Diagnostics{
			{
				Severity:      diag.Warning,
				Summary:       "Unauthorized to read gateway networks' private IPs, please check your IAM permissions",
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("private_ip"),
			},
		}
	default:
		return diag.Diagnostics{
			{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("Unable to get private IP for gateway network %s (gateway_id: %s)", gn.ID, gn.GatewayID),
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("private_ip"),
			},
		}
	}

	return nil
}

func getGatewayProjectID(ctx context.Context, api *vpcgw.API, zone scw.Zone, gatewayID string) (string, error) {
	gateway, err := api.GetGateway(&vpcgw.GetGatewayRequest{
		Zone:      zone,
		GatewayID: gatewayID,
	}, scw.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("get gateway network project ID: error getting gateway %s", gatewayID)
	}

	if gateway.ProjectID == "" {
		return "", fmt.Errorf("no project ID found for gateway %s", gatewayID)
	}

	return gateway.ProjectID, nil
}
