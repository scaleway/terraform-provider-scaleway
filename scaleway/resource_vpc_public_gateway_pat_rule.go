package scaleway

import (
	"context"
	"math"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayVPCPublicGatewayPATRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCPublicGatewayPATRuleCreate,
		ReadContext:   resourceScalewayVPCPublicGatewayPATRuleRead,
		UpdateContext: resourceScalewayVPCPublicGatewayPATRuleUpdate,
		DeleteContext: resourceScalewayVPCPublicGatewayPATRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
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
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					vpcgw.PATRuleProtocolTCP.String(),
					vpcgw.PATRuleProtocolUDP.String(),
					vpcgw.PATRuleProtocolBoth.String(),
				}, true),
				Default:     "both",
				Description: "The protocol used in the PAT rule",
			},
			"zone": zoneSchema(),
			// Computed elements
			"organization_id": organizationIDSchema(),
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
	}
}

func resourceScalewayVPCPublicGatewayPATRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	gatewayID := expandZonedID(d.Get("gateway_id").(string)).ID
	retryInterval := retryIntervalVPCPublicGatewayNetwork
	//check gateway is in stable state.
	_, err = vpcgwAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     gatewayID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
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

	res, err := vpcgwAPI.CreatePATRule(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	//check gateway is in stable state.
	_, err = vpcgwAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     res.GatewayID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, res.ID))

	return resourceScalewayVPCPublicGatewayPATRuleRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayPATRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patRules, err := vpcgwAPI.GetPATRule(&vpcgw.GetPATRuleRequest{
		PatRuleID: ID,
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	gatewayID := newZonedID(zone, patRules.GatewayID).String()
	_ = d.Set("created_at", patRules.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", patRules.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("gateway_id", gatewayID)
	_ = d.Set("private_ip", patRules.PrivateIP.String())
	_ = d.Set("private_port", int(patRules.PrivatePort))
	_ = d.Set("public_port", int(patRules.PublicPort))
	_ = d.Set("protocol", patRules.Protocol.String())
	_ = d.Set("zone", zone)

	return nil
}

func resourceScalewayVPCPublicGatewayPATRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patRules, err := vpcgwAPI.GetPATRule(&vpcgw.GetPATRuleRequest{
		PatRuleID: ID,
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	retryInterval := retryIntervalVPCPublicGatewayNetwork
	//check gateway is in stable state.
	_, err = vpcgwAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     patRules.GatewayID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("private_ip", "private_port", "public_port", "protocol") {
		port := uint32(d.Get("public_port").(int))
		privateIP := net.ParseIP(d.Get("private_ip").(string))
		privatePort := uint32(d.Get("private_port").(int))
		_, err = vpcgwAPI.UpdatePATRule(&vpcgw.UpdatePATRuleRequest{
			Zone:        zone,
			PatRuleID:   ID,
			PublicPort:  &port,
			PrivateIP:   &privateIP,
			PrivatePort: &privatePort,
			Protocol:    vpcgw.PATRuleProtocol(d.Get("protocol").(string)),
		}, scw.WithContext(ctx))

		if err != nil {
			if is404Error(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	return resourceScalewayVPCPublicGatewayPATRuleRead(ctx, d, meta)
}

func resourceScalewayVPCPublicGatewayPATRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, ID, err := vpcgwAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	patRules, err := vpcgwAPI.GetPATRule(&vpcgw.GetPATRuleRequest{
		PatRuleID: ID,
		Zone:      zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	retryInterval := retryIntervalVPCPublicGatewayNetwork
	//check gateway is in stable state.
	_, err = vpcgwAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     patRules.GatewayID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	err = vpcgwAPI.DeletePATRule(&vpcgw.DeletePATRuleRequest{
		PatRuleID: ID,
		Zone:      zone,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = vpcgwAPI.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID:     patRules.GatewayID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(gatewayWaitForTimeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
