package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceVPCPublicGateway() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePublicGateway().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["name"].ConflictsWith = []string{"public_gateway_id"}
	dsSchema["public_gateway_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the public gateway",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceVPCPublicGatewayRead,
	}
}

func DataSourceVPCPublicGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZoneV2(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.GetOk("zone"); ok {
		zone = scw.Zone(v.(string))
	}

	publicGatewayID, ok := d.GetOk("public_gateway_id")
	if !ok {
		gwName := d.Get("name").(string)

		res, err := api.ListGateways(
			&vpcgw.ListGatewaysRequest{
				Name:          types.ExpandStringPtr(gwName),
				Zone:          zone,
				ProjectID:     types.ExpandStringPtr(d.Get("project_id")),
				IncludeLegacy: scw.BoolPtr(true),
			}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundGW, err := datasource.FindExact(
			res.Gateways,
			func(s *vpcgw.Gateway) bool { return s.Name == gwName },
			gwName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		publicGatewayID = foundGW.ID
	}

	zonedID := datasource.NewZonedID(publicGatewayID, zone)
	d.SetId(zonedID)
	_ = d.Set("public_gateway_id", zonedID)

	return ResourceVPCPublicGatewayRead(ctx, d, m)
}
