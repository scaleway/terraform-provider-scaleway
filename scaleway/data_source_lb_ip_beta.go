package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayLbIPBeta() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayLbIPBeta().Schema)

	dsSchema["ip_address"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The IP address",
		ConflictsWith: []string{"ip_id"},
	}
	dsSchema["ip_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IP address",
		ConflictsWith: []string{"ip_address"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayLbIPBetaRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayLbIPBetaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := lbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, ok := d.GetOk("ip_id")
	if !ok { // Get IP by region and IP address.
		res, err := api.ListIPs(&lb.ListIPsRequest{
			Region:    region,
			IPAddress: expandStringPtr(d.Get("ip_address")),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.IPs) == 0 {
			return diag.FromErr(fmt.Errorf("no ips found with the address %s", d.Get("ip_address")))
		}
		if len(res.IPs) > 1 {
			return diag.FromErr(fmt.Errorf("%d ips found with the same address %s", len(res.IPs), d.Get("ip_address")))
		}
		ipID = res.IPs[0].ID
	}

	regionalID := datasourceNewRegionalizedID(ipID, region)
	d.SetId(regionalID)
	err = d.Set("ip_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayLbIPBetaRead(ctx, d, m)
}
