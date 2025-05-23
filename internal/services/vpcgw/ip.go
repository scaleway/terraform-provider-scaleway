package vpcgw

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceIP() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		CreateContext:                     ResourceIPCreate,
		ReadContext:                       ResourceIPRead,
		UpdateContext:                     ResourceVPCPublicGatewayIPUpdate,
		DeleteContext:                     ResourceVPCPublicGatewayIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Description: "the IP itself",
				Computed:    true,
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The tags associated with public gateway IP",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_id": account.ProjectIDSchema(),
			"zone":       zonal.Schema(),
			// Computed elements
			"organization_id": account.OrganizationIDSchema(),
			"reverse": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "reverse domain name for the IP address",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the public gateway IP",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the public gateway IP",
			},
		},
	}
}

func ResourceIPCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := newAPIWithZoneV2(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpcgw.CreateIPRequest{
		Tags:      types.ExpandStrings(d.Get("tags")),
		ProjectID: d.Get("project_id").(string),
		Zone:      zone,
	}

	res, err := api.CreateIP(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, res.ID))

	reverse := d.Get("reverse")
	if len(reverse.(string)) > 0 {
		updateRequest := &vpcgw.UpdateIPRequest{
			IPID:    res.ID,
			Zone:    zone,
			Tags:    scw.StringsPtr(types.ExpandStrings(d.Get("tags"))),
			Reverse: types.ExpandStringPtr(reverse.(string)),
		}

		_, err = api.UpdateIP(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceIPRead(ctx, d, m)
}

func ResourceIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ip, err := api.GetIP(&vpcgw.GetIPRequest{
		IPID: ID,
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("organization_id", ip.OrganizationID)
	_ = d.Set("address", ip.Address.String())
	_ = d.Set("project_id", ip.ProjectID)
	_ = d.Set("created_at", ip.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", ip.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("zone", zone)
	_ = d.Set("tags", ip.Tags)
	_ = d.Set("reverse", ip.Reverse)

	return nil
}

func ResourceVPCPublicGatewayIPUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateRequest := &vpcgw.UpdateIPRequest{
		IPID: ID,
		Zone: zone,
	}

	hasChanged := false

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if d.HasChange("reverse") {
		updateRequest.Reverse = types.ExpandStringPtr(d.Get("reverse").(string))
		hasChanged = true
	}

	if hasChanged {
		_, err = api.UpdateIP(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceIPRead(ctx, d, m)
}

func ResourceVPCPublicGatewayIPDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, ID, err := NewAPIWithZoneAndIDv2(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var warnings diag.Diagnostics

	err = api.DeleteIP(&vpcgw.DeleteIPRequest{
		IPID: ID,
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is409(err) || httperrors.Is412(err) || httperrors.Is404(err) {
			return append(warnings, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  err.Error(),
			})
		}

		return diag.FromErr(err)
	}

	return nil
}
