package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func dataSourceScalewayVPCPublicGatewayDHCP() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(resourceScalewayVPCPublicGatewayDHCP().Schema)

	dsSchema["dhcp_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "The ID of the public gateway DHCP configuration",
		ValidateFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: dataSourceScalewayVPCPublicGatewayDHCPRead,
	}
}

func dataSourceScalewayVPCPublicGatewayDHCPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, zone, err := vpcgwAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	dhcpID, _ := d.GetOk("dhcp_id")

	zonedID := datasource.NewZonedID(dhcpID, zone)
	d.SetId(zonedID)
	_ = d.Set("dhcp_id", zonedID)
	return resourceScalewayVPCPublicGatewayDHCPRead(ctx, d, m)
}
