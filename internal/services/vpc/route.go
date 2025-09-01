package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceRouteCreate,
		ReadContext:   ResourceRouteRead,
		UpdateContext: ResourceRouteUpdate,
		DeleteContext: ResourceRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "VPC ID the Route belongs to",
				DiffSuppressFunc: dsf.Locality,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The route description",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The tags associated with the Route",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"destination": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The destination IP or IP range of the route",
			},
			"nexthop_resource_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The ID of the nexthop resource",
				DiffSuppressFunc: diffSuppressFuncRouteResourceID,
			},
			"nexthop_private_network_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The ID of the nexthop private network",
				DiffSuppressFunc: dsf.Locality,
			},
			"region": regional.Schema(),
			// Computed elements
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the route",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the route",
			},
		},
	}
}

func ResourceRouteCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	destination, err := types.ExpandIPNet(d.Get("destination").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceID, err := vpcRouteExpandResourceID(d.Get("nexthop_resource_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpc.CreateRouteRequest{
		Description:             d.Get("description").(string),
		Tags:                    types.ExpandStrings(d.Get("tags")),
		VpcID:                   locality.ExpandID(d.Get("vpc_id").(string)),
		NexthopResourceID:       types.ExpandStringPtr(resourceID),
		NexthopPrivateNetworkID: types.ExpandStringPtr(locality.ExpandID(d.Get("nexthop_private_network_id"))),
		Destination:             destination,
		Region:                  region,
	}

	res, err := vpcAPI.CreateRoute(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	return ResourceRouteRead(ctx, d, m)
}

func ResourceRouteRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.GetRoute(&vpc.GetRouteRequest{
		Region:  region,
		RouteID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("description", res.Description)
	_ = d.Set("vpc_id", regional.NewIDString(region, res.VpcID))
	_ = d.Set("nexthop_resource_id", types.FlattenStringPtr(res.NexthopResourceID))
	_ = d.Set("nexthop_private_network_id", regional.NewIDString(region, types.FlattenStringPtr(res.NexthopPrivateNetworkID).(string)))
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("region", region)

	destination, err := types.FlattenIPNet(res.Destination)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("destination", destination)

	if len(res.Tags) > 0 {
		_ = d.Set("tags", res.Tags)
	}

	return nil
}

func ResourceRouteUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	updateRequest := &vpc.UpdateRouteRequest{
		Region:  region,
		RouteID: ID,
	}

	if d.HasChange("description") {
		updateRequest.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if d.HasChange("destination") {
		destination, err := types.ExpandIPNet(d.Get("destination").(string))
		if err != nil {
			return diag.FromErr(err)
		}

		updateRequest.Destination = &destination
		hasChanged = true
	}

	if d.HasChange("nexthop_resource_id") {
		resourceID, err := vpcRouteExpandResourceID(d.Get("nexthop_resource_id").(string))
		if err != nil {
			return diag.FromErr(err)
		}

		updateRequest.NexthopResourceID = types.ExpandUpdatedStringPtr(resourceID)
		hasChanged = true
	}

	if d.HasChange("nexthop_private_network_id") {
		updateRequest.NexthopPrivateNetworkID = types.ExpandUpdatedStringPtr(locality.ExpandID(d.Get("nexthop_private_network_id")))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err = vpcAPI.UpdateRoute(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceRouteRead(ctx, d, m)
}

func ResourceRouteDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// issue with route deletion for now, permissions_denied
	err = vpcAPI.DeleteRoute(&vpc.DeleteRouteRequest{
		Region:  region,
		RouteID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
