package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceDHCP() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceDHCP().Schema)

	dsSchema["dhcp_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Description:      "The ID of the public gateway DHCP configuration",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceVPCPublicGatewayDHCPRead,
	}
}

func DataSourceVPCPublicGatewayDHCPRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	_, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	dhcpID, _ := d.GetOk("dhcp_id")

	zonedID := datasource.NewZonedID(dhcpID, zone)
	d.SetId(zonedID)
	_ = d.Set("dhcp_id", zonedID)

	return ResourceVPCPublicGatewayDHCPRead(ctx, d, m)
}
