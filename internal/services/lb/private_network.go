package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourcePrivateNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLbPrivateNetworkCreate,
		ReadContext:   resourceLbPrivateNetworkRead,
		UpdateContext: resourceLbPrivateNetworkUpdate,
		DeleteContext: resourceLbPrivateNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Read:    schema.DefaultTimeout(defaultLbLbTimeout),
			Update:  schema.DefaultTimeout(defaultLbLbTimeout),
			Delete:  schema.DefaultTimeout(defaultLbLbTimeout),
			Default: schema.DefaultTimeout(defaultLbLbTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
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
		},
		CustomizeDiff: cdf.LocalityCheck("lb_id", "private_network_id"),
	}
}

func resourceLbPrivateNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, zonal.ExpandID(d.Get("lb_id").(string)).ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	attach, err := lbAPI.AttachPrivateNetwork(&lb.ZonedAPIAttachPrivateNetworkRequest{
		Zone:             zone,
		LBID:             zonal.ExpandID(d.Get("lb_id").(string)).ID,
		PrivateNetworkID: regional.ExpandID(d.Get("private_network_id").(string)).ID,
		IpamIDs:          types.ExpandStrings(d.Get("ipam_ip_ids")),
	}, scw.WithContext(ctx))
	if err != nil {
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

func resourceLbPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
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

	_ = d.Set("private_network_id", foundPN.PrivateNetworkID)
	_ = d.Set("lb_id", foundPN.LB.ID)
	_ = d.Set("ipam_ip_ids", foundPN.IpamIDs)
	_ = d.Set("status", foundPN.Status)
	_ = d.Set("created_at", foundPN.CreatedAt)
	_ = d.Set("updated_at", foundPN.UpdatedAt)
	_ = d.Set("zone", zone)
	_ = d.Set("project_id", foundPN.LB.ProjectID)

	return nil
}

func resourceLbPrivateNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceLbPrivateNetworkRead(ctx, d, m)
}

func resourceLbPrivateNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
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

	return nil
}
