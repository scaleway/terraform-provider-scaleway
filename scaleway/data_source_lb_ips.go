package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func dataSourceScalewayLbIPs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayLbIPsRead,
		Schema: map[string]*schema.Schema{
			"ip_cidr_range": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsCIDR,
				Description:  "IPs within a CIDR block like it are listed.",
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
						"zone":            zonal.Schema(),
						"organization_id": organizationIDSchema(),
						"project_id":      projectIDSchema(),
					},
				},
			},
			"zone":            zonal.Schema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func dataSourceScalewayLbIPsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := lbAPI.ListIPs(&lb.ZonedAPIListIPsRequest{
		Zone:      zone,
		ProjectID: expandStringPtr(d.Get("project_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var filteredList []*lb.IP
	for i := range res.IPs {
		if ipv4Match(d.Get("ip_cidr_range").(string), res.IPs[i].IPAddress) {
			filteredList = append(filteredList, res.IPs[i])
		}
	}

	ips := []interface{}(nil)
	for _, ip := range filteredList {
		rawIP := make(map[string]interface{})
		rawIP["id"] = zonal.NewID(ip.Zone, ip.ID).String()
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
