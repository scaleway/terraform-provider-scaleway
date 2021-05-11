package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceScalewayVPCPublicGatewayIP() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayVPCPublicGatewayIP().Schema)

	dsSchema["ip_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the IP",
		ValidateFunc: validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCPublicGatewayIPRead,
	}
}

func dataSourceScalewayVPCPublicGatewayIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	_, zone, err := vpcgwAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, _ := d.GetOk("ip_id")

	zonedID := datasourceNewZonedID(ipID, zone)
	d.SetId(zonedID)
	_ = d.Set("ip_id", zonedID)
	return resourceScalewayVPCPublicGatewayIPRead(ctx, d, meta)
}
