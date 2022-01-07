package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayLb() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayLb().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["name"].ConflictsWith = []string{"lb_id"}
	dsSchema["lb_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the load-balancer",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}
	dsSchema["release_ip"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Release the IPs related to this load-balancer",
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	lbID, ok := d.GetOk("lb_id")
	if !ok { // Get LB by name.
		res, err := api.ListLBs(&lb.ZonedAPIListLBsRequest{
			Zone:      zone,
			Name:      expandStringPtr(d.Get("name")),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.LBs) == 0 {
			return diag.FromErr(fmt.Errorf("no lbs found with the name %s", d.Get("name")))
		}
		if len(res.LBs) > 1 {
			return diag.FromErr(fmt.Errorf("%d lbs found with the same name %s", len(res.LBs), d.Get("name")))
		}
		lbID = res.LBs[0].ID
	}

	err = d.Set("release_ip", false)
	if err != nil {
		return diag.FromErr(err)
	}
	zonedID := datasourceNewZonedID(lbID, zone)
	d.SetId(zonedID)
	err = d.Set("lb_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbRead(ctx, d, meta)
}
