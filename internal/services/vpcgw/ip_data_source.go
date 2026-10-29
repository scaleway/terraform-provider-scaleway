package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceIP() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceIP().SchemaFunc())

	dsSchema["ip_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the IP",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		Schema:      dsSchema,
		ReadContext: DataSourceVPCPublicGatewayIPRead,
	}
}

func DataSourceVPCPublicGatewayIPRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, _ := d.GetOk("ip_id")

	zonedID := datasource.NewZonedID(ipID, zone)
	d.SetId(zonedID)
	_ = d.Set("ip_id", zonedID)

	ip, err := api.GetIP(&vpcgw.GetIPRequest{
		IPID: locality.ExpandID(ipID),
		Zone: zone,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setIPState(d, ip)
}
