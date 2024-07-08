package vpcgw

import (
	"context"
	"math"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourcePATRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceVPCPublicGatewayPATRuleCreate,
		ReadContext:   ResourceVPCPublicGatewayPATRuleRead,
		UpdateContext: ResourceVPCPublicGatewayPATRuleUpdate,
		DeleteContext: ResourceVPCPublicGatewayPATRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultTimeout),
			Update:  schema.DefaultTimeout(defaultTimeout),
			Delete:  schema.DefaultTimeout(defaultTimeout),
			Default: schema.DefaultTimeout(defaultTimeout),
		},
		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:  "The ID of the gateway this PAT rule is applied to",
			},
			"private_ip": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "The private IP used in the PAT rule",
			},
			"public_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
				Description:  "The public port used in the PAT rule",
			},
			"private_port": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: validation.IntBetween(
					0,
					math.MaxUint16,
				),
				Description: "The private port used in the PAT rule",
			},
			"protocol": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: verify.ValidateEnumIgnoreCase[vpcgw.PATRuleProtocol](),
				Default:          "both",
				Description:      "The protocol used in the PAT rule",
			},
			"zone": zonal.Schema(),
			// Computed elements
			"organization_id": account.OrganizationIDSchema(),
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the PAT rule",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the PAT rule",
			},
		},
		CustomizeDiff: cdf.LocalityCheck("gateway_id"),
	}
}

func ResourceVPCPublicGatewayPATRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayID := zonal.ExpandID(d.Get("gateway_id").(string)).ID
	_, err = waitForVPCPublicGateway(ctx, api, zone, gatewayID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpcgw.CreatePATRuleRequest{
		Zone:        zone,
		GatewayID:   gatewayID,
		PublicPort:  uint32(d.Get("public_port").(int)),
		PrivateIP:   net.ParseIP(d.Get("private_ip").(string)),
		PrivatePort: uint32(d.Get("private_port").(int)),
		Protocol:    vpcgw.PATRuleProtocol(d.Get("protocol").(string)),
	}

	_, err = waitForVPCPublicGateway(ctx, api, zone, gatewayID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	patRule, err := api.CreatePATRule(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, patRule.ID))

	_, err = waitForVPCPublicGateway(ctx, api, zone, patRule.GatewayID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCPublicGatewayPATRuleRead(ctx, d, m)
}

func ResourceVPCPublicGatewayPATRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patRule, err := api.GetPATRule(&vpcgw.GetPATRuleRequest{
		PatRuleID: ID,
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	gatewayID := zonal.NewID(zone, patRule.GatewayID).String()
	_ = d.Set("created_at", patRule.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", patRule.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("gateway_id", gatewayID)
	_ = d.Set("private_ip", patRule.PrivateIP.String())
	_ = d.Set("private_port", int(patRule.PrivatePort))
	_ = d.Set("public_port", int(patRule.PublicPort))
	_ = d.Set("protocol", patRule.Protocol.String())
	_ = d.Set("zone", zone)

	return nil
}

func ResourceVPCPublicGatewayPATRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patRule, err := api.GetPATRule(&vpcgw.GetPATRuleRequest{
		PatRuleID: ID,
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// check gateway is in stable state.
	_, err = waitForVPCPublicGateway(ctx, api, zone, patRule.GatewayID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpcgw.UpdatePATRuleRequest{
		Zone:      zone,
		PatRuleID: ID,
		Protocol:  vpcgw.PATRuleProtocol(d.Get("protocol").(string)),
	}

	hasChange := false
	if d.HasChange("private_ip") {
		req.PrivateIP = scw.IPPtr(net.ParseIP(d.Get("private_ip").(string)))
		hasChange = true
	}

	if d.HasChange("private_port") {
		req.PrivatePort = scw.Uint32Ptr(uint32(d.Get("private_port").(int)))
		hasChange = true
	}

	if d.HasChange("public_port") {
		req.PublicPort = scw.Uint32Ptr(uint32(d.Get("public_port").(int)))
		hasChange = true
	}

	if d.HasChange("protocol") {
		req.Protocol = vpcgw.PATRuleProtocol(d.Get("protocol").(string))
		hasChange = true
	}

	if hasChange {
		// check gateway is in stable state.
		_, err = waitForVPCPublicGateway(ctx, api, zone, patRule.GatewayID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		patRule, err = api.UpdatePATRule(req, scw.WithContext(ctx))
		if err != nil {
			if httperrors.Is404(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	// check gateway is in stable state.
	_, err = waitForVPCPublicGateway(ctx, api, zone, patRule.GatewayID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCPublicGatewayPATRuleRead(ctx, d, m)
}

func ResourceVPCPublicGatewayPATRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patRule, err := api.GetPATRule(&vpcgw.GetPATRuleRequest{
		PatRuleID: ID,
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// check gateway is in stable state.
	_, err = waitForVPCPublicGateway(ctx, api, zone, patRule.GatewayID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	err = api.DeletePATRule(&vpcgw.DeletePATRuleRequest{
		PatRuleID: ID,
		Zone:      zone,
	}, scw.WithContext(ctx))

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForVPCPublicGateway(ctx, api, zone, patRule.GatewayID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
