package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayLbIPs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayLbsRead,
		Schema: map[string]*schema.Schema{
			"ip_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IPs with an address like it are listed.",
			},
			"ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"ip_address": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"lb_id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"reverse": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"zone":            zoneSchema(),
						"organization_id": organizationIDSchema(),
						"project_id":      projectIDSchema(),
					},
				},
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func dataSourceScalewayLbsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := lbAPI.ListIPs(&lb.ZonedAPIListIPsRequest{
		Zone:      zone,
		IPAddress: expandStringPtr(d.Get("ip_address")),
		ProjectID: expandStringPtr(d.Get("project_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	ips := []interface{}(nil)
	for _, ip := range res.IPs {
		rawIP := make(map[string]interface{})
		rawIP["id"] = newZonedID(ip.Zone, ip.ID).String()
		rawIP["ip_address"] = ip.IPAddress
		rawIP["lb_id"] = flattenStringPtr(ip.LBID)
		rawIP["reverse"] = ip.Reverse
		rawIP["zone"] = string(zone)
		rawIP["organization_id"] = ip.OrganizationID
		rawIP["project_id"] = ip.ProjectID

		ips = append(ips, rawIP)
	}

	d.SetId(zone.String())
	_ = d.Set("ips", ips)

	return nil
}
