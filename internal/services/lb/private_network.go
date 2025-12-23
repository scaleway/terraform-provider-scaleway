package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourcePrivateNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLbPrivateNetworkCreate,
		ReadContext:   resourceLbPrivateNetworkRead,
		DeleteContext: resourceLbPrivateNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Read:    schema.DefaultTimeout(defaultLbLbTimeout),
			Delete:  schema.DefaultTimeout(defaultLbLbTimeout),
			Default: schema.DefaultTimeout(defaultLbLbTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    privateNetworkSchema,
		CustomizeDiff: cdf.LocalityCheck("lb_id", "private_network_id"),
		Identity: identity.WrapSchemaMap(map[string]*schema.Schema{
			"zone": identity.DefaultZoneAttribute(),
			"lb_id": {
				RequiredForImport: true,
				Description:       "The ID of the load balancer (UUID format)",
			},
			"private_network_id": {
				RequiredForImport: true,
				Description:       "The ID of the private network (UUID format)",
			},
		}),
	}
}

func privateNetworkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"lb_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The load-balancer ID to attach the private network to",
		},
		"private_network_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The private network ID to attach",
		},
		"ipam_ip_ids": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			MaxItems:    1,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
			Description: "IPAM ID of a pre-reserved IP address to assign to the Load Balancer on this Private Network",
		},
		"zone": zonal.Schema(),
		// Computed
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of private network connection",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the creation of the private network connection",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of the last update of the private network connection",
		},
		"project_id": account.ProjectIDSchema(),
	}
}

func resourceLbPrivateNetworkCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	lbID := zonal.ExpandID(d.Get("lb_id").(string)).ID

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	attach, err := lbAPI.AttachPrivateNetwork(&lb.ZonedAPIAttachPrivateNetworkRequest{
		Zone:             zone,
		LBID:             lbID,
		PrivateNetworkID: regional.ExpandID(d.Get("private_network_id").(string)).ID,
		IpamIDs:          locality.ExpandIDs(d.Get("ipam_ip_ids")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutUpdate))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForPrivateNetworks(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutUpdate))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	d.SetId(
		zonal.NewNestedIDString(
			zone,
			attach.LB.ID,
			attach.PrivateNetworkID,
		),
	)

	return resourceLbPrivateNetworkRead(ctx, d, m)
}

func resourceLbPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, _, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, LBID, PNID, err := ResourceLBPrivateNetworkParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworks, err := waitForPrivateNetworks(ctx, lbAPI, zone, LBID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	var foundPN *lb.PrivateNetwork

	for _, pn := range privateNetworks {
		if pn.PrivateNetworkID == PNID {
			foundPN = pn

			break
		}
	}

	if foundPN == nil {
		d.SetId("")

		return nil
	}

	region, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("private_network_id", regional.NewIDString(region, foundPN.PrivateNetworkID))
	_ = d.Set("lb_id", zonal.NewIDString(zone, foundPN.LB.ID))
	_ = d.Set("ipam_ip_ids", regional.NewIDStrings(region, foundPN.IpamIDs))
	_ = d.Set("status", foundPN.Status.String())
	_ = d.Set("created_at", types.FlattenTime(foundPN.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(foundPN.UpdatedAt))
	_ = d.Set("zone", foundPN.LB.Zone)
	_ = d.Set("project_id", foundPN.LB.ProjectID)

	return nil
}

func resourceLbPrivateNetworkDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, _, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	zone, LBID, PNID, err := ResourceLBPrivateNetworkParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = lbAPI.DetachPrivateNetwork(&lb.ZonedAPIDetachPrivateNetworkRequest{
		Zone:             zone,
		LBID:             LBID,
		PrivateNetworkID: PNID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, LBID, d.Timeout(schema.TimeoutUpdate))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForPrivateNetworks(ctx, lbAPI, zone, LBID, d.Timeout(schema.TimeoutUpdate))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
