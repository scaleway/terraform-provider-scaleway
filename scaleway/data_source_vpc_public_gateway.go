package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vpcgw "github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayVPCPublicGateway() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayVPCPublicGateway().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name")

	dsSchema["name"].ConflictsWith = []string{"public_gateway_id"}
	dsSchema["public_gateway_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the public gateway",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCPublicGatewayRead,
	}
}

func dataSourceScalewayVPCPublicGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcgwAPI, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	publicGatewayID, ok := d.GetOk("public_gateway_id")
	if !ok {
		res, err := vpcgwAPI.ListGateways(
			&vpcgw.ListGatewaysRequest{
				Name: expandStringPtr(d.Get("name").(string)),
				Zone: zone,
			}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if res.TotalCount == 0 {
			return diag.FromErr(
				fmt.Errorf(
					"no public gateway found with the name %s",
					d.Get("name"),
				),
			)
		}
		if res.TotalCount > 1 {
			return diag.FromErr(
				fmt.Errorf(
					"%d public gateways found with the name %s",
					res.TotalCount,
					d.Get("name"),
				),
			)
		}
		publicGatewayID = res.Gateways[0].ID
	}

	zonedID := datasourceNewZonedID(publicGatewayID, zone)
	d.SetId(zonedID)
	_ = d.Set("public_gateway_id", zonedID)
	return resourceScalewayVPCPublicGatewayRead(ctx, d, meta)
}
