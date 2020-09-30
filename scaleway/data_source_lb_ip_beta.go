package scaleway

import (
	"fmt"

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
		Read:   dataSourceScalewayLbIPBetaRead,
		Schema: dsSchema,
	}
}

func dataSourceScalewayLbIPBetaRead(d *schema.ResourceData, m interface{}) error {
	api, region, err := lbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	ipID, ok := d.GetOk("ip_id")
	if !ok { // Get IP by region and IP address.
		res, err := api.ListIPs(&lb.ListIPsRequest{
			Region:    region,
			IPAddress: scw.StringPtr(d.Get("ip_address").(string)),
		})
		if err != nil {
			return err
		}
		if len(res.IPs) == 0 {
			return fmt.Errorf("no ips found with the address %s", d.Get("ip_address"))
		}
		if len(res.IPs) > 1 {
			return fmt.Errorf("%d ips found with the same address %s", len(res.IPs), d.Get("ip_address"))
		}
		ipID = res.IPs[0].ID
	}

	regionalID := datasourceNewRegionalizedID(ipID, region)
	d.SetId(regionalID)
	err = d.Set("ip_id", regionalID)
	if err != nil {
		return err
	}
	return resourceScalewayLbIPBetaRead(d, m)
}
