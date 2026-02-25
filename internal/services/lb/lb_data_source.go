package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceLb() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceLb().SchemaFunc())

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["name"].ConflictsWith = []string{"lb_id"}
	dsSchema["lb_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the load-balancer",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}
	dsSchema["release_ip"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Release the IPs related to this load-balancer",
	}

	return &schema.Resource{
		ReadContext: DataSourceLbRead,
		Schema:      dsSchema,
	}
}

func DataSourceLbRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	lbID, ok := d.GetOk("lb_id")
	if !ok { // Get LB by name.
		lbName := d.Get("name").(string)

		res, err := api.ListLBs(&lbSDK.ZonedAPIListLBsRequest{
			Zone:      zone,
			Name:      types.ExpandStringPtr(lbName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundLB, err := datasource.FindExact(
			res.LBs,
			func(s *lbSDK.LB) bool { return s.Name == lbName },
			lbName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		lbID = foundLB.ID
	}

	err = d.Set("release_ip", false)
	if err != nil {
		return diag.FromErr(err)
	}

	zonedID := datasource.NewZonedID(lbID, zone)
	d.SetId(zonedID)

	err = d.Set("lb_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	lb, err := waitForInstances(ctx, api, zone, locality.ExpandID(lbID.(string)), d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	region, err := zone.Region()
	if err != nil {
		return diag.FromErr(err)
	}

	privateNetworks, err := waitForPrivateNetworks(ctx, api, zone, lb.ID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	allPrivateIPs, err := getLBPrivateIPs(ctx, m, region, lb.ID, privateNetworks)
	if err != nil {
		return diag.FromErr(err)
	}

	return setLBState(d, lb, zone, region, privateNetworks, allPrivateIPs, false)
}
