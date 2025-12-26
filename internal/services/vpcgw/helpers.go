package vpcgw

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultTimeout             = 10 * time.Minute
	defaultRetry               = 5 * time.Second
	defaultIPReverseDNSTimeout = 10 * time.Minute
)

// newAPIWithZone returns a new VPC API and the zone for a Create request
func newAPIWithZone(d *schema.ResourceData, m any) (*vpcgw.API, scw.Zone, error) {
	api := vpcgw.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, zone, nil
}

func newAPIWithZoneV2(d *schema.ResourceData, m any) (*v2.API, scw.Zone, error) {
	api := v2.NewAPI(meta.ExtractScwClient(m))

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

func NewAPIWithZoneAndIDv2(m any, id string) (*v2.API, scw.Zone, string, error) {
	api := v2.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, zone, ID, nil
}

func retryUpdateGatewayReverseDNS(ctx context.Context, api *v2.API, req *v2.UpdateIPRequest, timeout time.Duration) error {
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

func expandUpdateIpamConfig(raw any) *vpcgw.UpdateGatewayNetworkRequestIpamConfig {
	if raw == nil || len(raw.([]any)) != 1 {
		return nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	updateIpamConfig := &vpcgw.UpdateGatewayNetworkRequestIpamConfig{
		PushDefaultRoute: scw.BoolPtr(rawMap["push_default_route"].(bool)),
	}

	if ipamIPID, ok := rawMap["ipam_ip_id"].(string); ok && ipamIPID != "" {
		updateIpamConfig.IpamIPID = scw.StringPtr(regional.ExpandID(ipamIPID).ID)
	}

	return updateIpamConfig
}

func expandIpamConfigV2(raw any) (bool, *string) {
	if raw == nil || len(raw.([]any)) != 1 {
		return false, nil
	}

	rawMap := raw.([]any)[0].(map[string]any)

	pushDefaultRoute := rawMap["push_default_route"].(bool)

	var ipamIPID *string

	if IPID, ok := rawMap["ipam_ip_id"].(string); ok && IPID != "" {
		ipamIPID = scw.StringPtr(regional.ExpandID(IPID).ID)
	}

	return pushDefaultRoute, ipamIPID
}

func flattenIpamConfig(config *vpcgw.IpamConfig, region scw.Region) any {
	if config == nil {
		return nil
	}

	return []map[string]any{
		{
			"push_default_route": config.PushDefaultRoute,
			"ipam_ip_id":         regional.NewIDString(region, config.IpamIPID),
		},
	}
}

// FlattenIPNetList turns a slice of scw.IPNet into a slice of string CIDRs.
func FlattenIPNetList(ipNets []scw.IPNet) ([]string, error) {
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

// readVPCGWResourceDataV1 sets the resource data using a v1 gateway
func readVPCGWResourceDataV1(d *schema.ResourceData, gw *vpcgw.Gateway) diag.Diagnostics {
	_ = d.Set("name", gw.Name)
	_ = d.Set("status", string(gw.Status))
	_ = d.Set("organization_id", gw.OrganizationID)
	_ = d.Set("project_id", gw.ProjectID)
	_ = d.Set("zone", gw.Zone)
	_ = d.Set("tags", gw.Tags)
	_ = d.Set("upstream_dns_servers", gw.UpstreamDNSServers)
	_ = d.Set("bastion_enabled", gw.BastionEnabled)
	_ = d.Set("bastion_port", int(gw.BastionPort))
	_ = d.Set("enable_smtp", gw.SMTPEnabled)

	if gw.Type != nil {
		_ = d.Set("type", gw.Type.Name)
	}

	if gw.IP != nil {
		_ = d.Set("ip_id", zonal.NewID(gw.IP.Zone, gw.IP.ID).String())
	}

	if gw.CreatedAt != nil {
		_ = d.Set("created_at", gw.CreatedAt.Format(time.RFC3339))
	}

	if gw.UpdatedAt != nil {
		_ = d.Set("updated_at", gw.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}

// readVPCGWResourceDataV2 sets the resource data using a v2 gateway
func readVPCGWResourceDataV2(d *schema.ResourceData, gw *v2.Gateway) diag.Diagnostics {
	_ = d.Set("name", gw.Name)
	_ = d.Set("type", gw.Type)
	_ = d.Set("status", gw.Status.String())
	_ = d.Set("organization_id", gw.OrganizationID)
	_ = d.Set("project_id", gw.ProjectID)
	_ = d.Set("zone", gw.Zone)
	_ = d.Set("tags", gw.Tags)
	_ = d.Set("bastion_enabled", gw.BastionEnabled)
	_ = d.Set("bastion_port", int(gw.BastionPort))
	_ = d.Set("enable_smtp", gw.SMTPEnabled)
	_ = d.Set("bandwidth", int(gw.Bandwidth))
	_ = d.Set("upstream_dns_servers", nil)

	ips, err := FlattenIPNetList(gw.BastionAllowedIPs)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("allowed_ip_ranges", ips)

	if gw.IPv4 != nil {
		_ = d.Set("ip_id", zonal.NewID(gw.IPv4.Zone, gw.IPv4.ID).String())
	}

	if gw.CreatedAt != nil {
		_ = d.Set("created_at", gw.CreatedAt.Format(time.RFC3339))
	}

	if gw.UpdatedAt != nil {
		_ = d.Set("updated_at", gw.UpdatedAt.Format(time.RFC3339))
	}

	err = identity.SetZonalIdentity(d, gw.Zone, gw.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// readVPCGWNetworkResourceDataV1 sets the resource data using a v1 gateway network
func readVPCGWNetworkResourceDataV1(d *schema.ResourceData, gatewayNetwork *vpcgw.GatewayNetwork, diags diag.Diagnostics) diag.Diagnostics {
	fetchRegion, err := gatewayNetwork.Zone.Region()
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}

	_ = d.Set("private_network_id", regional.NewIDString(fetchRegion, gatewayNetwork.PrivateNetworkID))
	_ = d.Set("gateway_id", zonal.NewIDString(gatewayNetwork.Zone, gatewayNetwork.GatewayID))
	_ = d.Set("enable_masquerade", gatewayNetwork.EnableMasquerade)
	_ = d.Set("status", string(gatewayNetwork.Status))
	_ = d.Set("zone", gatewayNetwork.Zone)

	if macAddress := gatewayNetwork.MacAddress; macAddress != nil {
		_ = d.Set("mac_address", types.FlattenStringPtr(macAddress).(string))
	}

	if gatewayNetwork.CreatedAt != nil {
		_ = d.Set("created_at", gatewayNetwork.CreatedAt.Format(time.RFC3339))
	}

	if gatewayNetwork.UpdatedAt != nil {
		_ = d.Set("updated_at", gatewayNetwork.UpdatedAt.Format(time.RFC3339))
	}

	var cleanUpDHCPValue bool

	cleanUpDHCP, cleanUpDHCPExist := d.GetOk("cleanup_dhcp")

	if cleanUpDHCPExist {
		cleanUpDHCPValue = *types.ExpandBoolPtr(cleanUpDHCP)
	}

	_ = d.Set("cleanup_dhcp", cleanUpDHCPValue)

	if ipamConfig := gatewayNetwork.IpamConfig; ipamConfig != nil {
		_ = d.Set("ipam_config", flattenIpamConfig(ipamConfig, fetchRegion))
	}

	return nil
}

// readVPCGWNetworkResourceDataV2 sets the resource data using a v1 gateway network
func readVPCGWNetworkResourceDataV2(d *schema.ResourceData, gatewayNetwork *v2.GatewayNetwork, diags diag.Diagnostics) diag.Diagnostics {
	fetchRegion, err := gatewayNetwork.Zone.Region()
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}

	_ = d.Set("private_network_id", regional.NewIDString(fetchRegion, gatewayNetwork.PrivateNetworkID))
	_ = d.Set("gateway_id", zonal.NewIDString(gatewayNetwork.Zone, gatewayNetwork.GatewayID))
	_ = d.Set("enable_masquerade", gatewayNetwork.MasqueradeEnabled)
	_ = d.Set("status", string(gatewayNetwork.Status))
	_ = d.Set("zone", gatewayNetwork.Zone)

	if macAddress := gatewayNetwork.MacAddress; macAddress != nil {
		_ = d.Set("mac_address", types.FlattenStringPtr(macAddress).(string))
	}

	if gatewayNetwork.CreatedAt != nil {
		_ = d.Set("created_at", gatewayNetwork.CreatedAt.Format(time.RFC3339))
	}

	if gatewayNetwork.UpdatedAt != nil {
		_ = d.Set("updated_at", gatewayNetwork.UpdatedAt.Format(time.RFC3339))
	}

	ipamConfig := []map[string]any{
		{
			"push_default_route": gatewayNetwork.PushDefaultRoute,
			"ipam_ip_id":         gatewayNetwork.IpamIPID,
		},
	}

	_ = d.Set("ipam_config", ipamConfig)

	err = identity.SetZonalIdentity(d, gatewayNetwork.Zone, gatewayNetwork.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func setPrivateIPsV1(ctx context.Context, d *schema.ResourceData, api *vpcgw.API, gn *vpcgw.GatewayNetwork, m any) diag.Diagnostics {
	resourceID := gn.ID

	region, err := gn.Zone.Region()
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Warning,
				Summary:  "Unable to get gateway network's private IP",
				Detail:   err.Error(),
			},
		}
	}

	projectID, err := getGatewayProjectIDV1(ctx, api, gn.Zone, gn.GatewayID)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Warning,
				Summary:  "Unable to get gateway network's private IP",
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

func setPrivateIPsV2(ctx context.Context, d *schema.ResourceData, api *v2.API, gn *v2.GatewayNetwork, m any) diag.Diagnostics {
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

	projectID, err := getGatewayProjectIDV2(ctx, api, gn.Zone, gn.GatewayID)
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

func getGatewayProjectIDV1(ctx context.Context, api *vpcgw.API, zone scw.Zone, gatewayID string) (string, error) {
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

func getGatewayProjectIDV2(ctx context.Context, api *v2.API, zone scw.Zone, gatewayID string) (string, error) {
	gateway, err := api.GetGateway(&v2.GetGatewayRequest{
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

// updateGatewayV1 performs the update of the public gateway using the v1 API
func updateGatewayV1(ctx context.Context, d *schema.ResourceData, apiV1 *vpcgw.API, zone scw.Zone, id string) error {
	v1UpdateRequest := &vpcgw.UpdateGatewayRequest{
		GatewayID: id,
		Zone:      zone,
	}

	if d.HasChange("name") {
		v1UpdateRequest.Name = scw.StringPtr(d.Get("name").(string))
	}

	if d.HasChange("tags") {
		v1UpdateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("bastion_port") {
		v1UpdateRequest.BastionPort = scw.Uint32Ptr(uint32(d.Get("bastion_port").(int)))
	}

	if d.HasChange("bastion_enabled") {
		v1UpdateRequest.EnableBastion = scw.BoolPtr(d.Get("bastion_enabled").(bool))
	}

	if d.HasChange("enable_smtp") {
		v1UpdateRequest.EnableSMTP = scw.BoolPtr(d.Get("enable_smtp").(bool))
	}

	if _, err := apiV1.UpdateGateway(v1UpdateRequest, scw.WithContext(ctx)); err != nil {
		return err
	}

	_, err := waitForVPCPublicGateway(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	if d.HasChange("refresh_ssh_keys") {
		if _, err := apiV1.RefreshSSHKeys(&vpcgw.RefreshSSHKeysRequest{
			Zone:      zone,
			GatewayID: id,
		}, scw.WithContext(ctx)); err != nil {
			return err
		}
	}

	_, err = waitForVPCPublicGateway(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	if d.HasChange("type") {
		if _, err := apiV1.UpgradeGateway(&vpcgw.UpgradeGatewayRequest{
			Zone:      zone,
			GatewayID: id,
			Type:      types.ExpandUpdatedStringPtr(d.Get("type")),
		}, scw.WithContext(ctx)); err != nil {
			return err
		}
	}

	_, err = waitForVPCPublicGateway(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	return nil
}

// updateGatewayV2 performs the update of the public gateway using the v2 API
func updateGatewayV2(ctx context.Context, d *schema.ResourceData, api *v2.API, zone scw.Zone, id string) error {
	updateRequest := &v2.UpdateGatewayRequest{
		GatewayID: id,
		Zone:      zone,
	}

	if d.HasChange("name") {
		updateRequest.Name = scw.StringPtr(d.Get("name").(string))
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("bastion_port") {
		updateRequest.BastionPort = scw.Uint32Ptr(uint32(d.Get("bastion_port").(int)))
	}

	if d.HasChange("bastion_enabled") {
		updateRequest.EnableBastion = scw.BoolPtr(d.Get("bastion_enabled").(bool))
	}

	if d.HasChange("enable_smtp") {
		updateRequest.EnableSMTP = scw.BoolPtr(d.Get("enable_smtp").(bool))
	}

	if _, err := api.UpdateGateway(updateRequest, scw.WithContext(ctx)); err != nil {
		return err
	}

	_, err := waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	if d.HasChange("refresh_ssh_keys") {
		if _, err := api.RefreshSSHKeys(&v2.RefreshSSHKeysRequest{
			Zone:      zone,
			GatewayID: id,
		}, scw.WithContext(ctx)); err != nil {
			return err
		}
	}

	_, err = waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	if d.HasChange("type") {
		if _, err := api.UpgradeGateway(&v2.UpgradeGatewayRequest{
			Zone:      zone,
			GatewayID: id,
			Type:      types.ExpandUpdatedStringPtr(d.Get("type")),
		}, scw.WithContext(ctx)); err != nil {
			return err
		}
	}

	_, err = waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	if d.HasChange("allowed_ip_ranges") {
		listIPs := d.Get("allowed_ip_ranges").(*schema.Set).List()

		_, err := api.SetBastionAllowedIPs(&v2.SetBastionAllowedIPsRequest{
			GatewayID: id,
			Zone:      zone,
			IPRanges:  types.ExpandStrings(listIPs),
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}
	}

	_, err = waitForVPCPublicGatewayV2(ctx, api, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	return nil
}

// updateGateway wraps the update process, trying the v2 API first and falling back to v1 on a 412 error
func updateGateway(ctx context.Context, d *schema.ResourceData, api *v2.API, apiV1 *vpcgw.API, zone scw.Zone, id string) error {
	err := updateGatewayV2(ctx, d, api, zone, id)
	if err != nil {
		if httperrors.Is412(err) {
			return updateGatewayV1(ctx, d, apiV1, zone, id)
		}

		return err
	}

	return nil
}

// updateGWNetworkV2 performs the update of the gateway network using the v2 API
func updateGWNetworkV2(ctx context.Context, d *schema.ResourceData, api *v2.API, zone scw.Zone, id string) error {
	updateRequest := &v2.UpdateGatewayNetworkRequest{
		GatewayNetworkID: id,
		Zone:             zone,
	}

	if d.HasChange("enable_masquerade") {
		updateRequest.EnableMasquerade = types.ExpandBoolPtr(d.Get("enable_masquerade"))
	}

	if d.HasChange("ipam_config") {
		pushDefaultRoute, ipamIPID := expandIpamConfigV2(d.Get("ipam_config"))

		updateRequest.PushDefaultRoute = scw.BoolPtr(pushDefaultRoute)
		updateRequest.IpamIPID = ipamIPID
	}

	if _, err := api.UpdateGatewayNetwork(updateRequest, scw.WithContext(ctx)); err != nil {
		return err
	}

	_, err := waitForVPCGatewayNetworkV2(ctx, api, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return err
	}

	return nil
}

// updateGWNetworkV1 performs the update of the gateway network using the v1 API
func updateGWNetworkV1(ctx context.Context, d *schema.ResourceData, apiV1 *vpcgw.API, zone scw.Zone, id string) error {
	v1UpdateRequest := &vpcgw.UpdateGatewayNetworkRequest{
		GatewayNetworkID: id,
		Zone:             zone,
	}

	if d.HasChange("enable_masquerade") {
		v1UpdateRequest.EnableMasquerade = types.ExpandBoolPtr(d.Get("enable_masquerade"))
	}

	if d.HasChange("enable_dhcp") {
		v1UpdateRequest.EnableDHCP = types.ExpandBoolPtr(d.Get("enable_dhcp"))
	}

	if d.HasChange("dhcp_id") {
		dhcpID := zonal.ExpandID(d.Get("dhcp_id").(string)).ID
		v1UpdateRequest.DHCPID = &dhcpID
	}

	if d.HasChange("ipam_config") {
		v1UpdateRequest.IpamConfig = expandUpdateIpamConfig(d.Get("ipam_config"))
	}

	if d.HasChange("static_address") {
		if staticAddress, ok := d.GetOk("static_address"); ok {
			address, err := types.ExpandIPNet(staticAddress.(string))
			if err != nil {
				return err
			}

			v1UpdateRequest.Address = &address
		}
	}

	if _, err := apiV1.UpdateGatewayNetwork(v1UpdateRequest, scw.WithContext(ctx)); err != nil {
		return err
	}

	_, err := waitForVPCGatewayNetwork(ctx, apiV1, zone, id, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	return nil
}

// updateGWNetwork wraps the update process: try v2 update first, then fallback to v1 update on a 412 error
func updateGWNetwork(ctx context.Context, d *schema.ResourceData, api *v2.API, apiV1 *vpcgw.API, zone scw.Zone, id string) error {
	err := updateGWNetworkV2(ctx, d, api, zone, id)
	if err != nil {
		if httperrors.Is412(err) {
			return updateGWNetworkV1(ctx, d, apiV1, zone, id)
		}

		return err
	}

	return nil
}
