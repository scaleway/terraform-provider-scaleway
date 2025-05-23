package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceIPs() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		ReadContext:                       DataSourceLbIPsRead,
		Schema: map[string]*schema.Schema{
			"ip_cidr_range": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsCIDR,
				Description:  "IPs within a CIDR block like it are listed.",
			},
			"ip_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  lb.ListIPsRequestIPTypeAll.String(),
				ValidateFunc: validation.StringInSlice([]string{
					lb.ListIPsRequestIPTypeIPv4.String(),
					lb.ListIPsRequestIPTypeIPv6.String(),
					lb.ListIPsRequestIPTypeAll.String(),
				}, false),
				Description: "IP type to filter for",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IPs with these exact tags are listed",
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
						"tags": {
							Computed: true,
							Type:     schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"zone":            zonal.Schema(),
						"organization_id": account.OrganizationIDSchema(),
						"project_id":      account.ProjectIDSchema(),
					},
				},
			},
			"zone":            zonal.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func DataSourceLbIPsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := lbAPI.ListIPs(&lb.ZonedAPIListIPsRequest{
		Zone:      zone,
		ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		Tags:      types.ExpandStrings(d.Get("tags")),
		IPType:    lb.ListIPsRequestIPType(d.Get("ip_type").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var filteredList []*lb.IP

	if cidrRange, ok := d.GetOk("ip_cidr_range"); ok {
		for i := range res.IPs {
			if ipv4Match(cidrRange.(string), res.IPs[i].IPAddress) {
				filteredList = append(filteredList, res.IPs[i])
			}
		}
	} else {
		filteredList = res.IPs
	}

	ips := []interface{}(nil)

	for _, ip := range filteredList {
		rawIP := make(map[string]interface{})
		rawIP["id"] = zonal.NewID(ip.Zone, ip.ID).String()
		rawIP["ip_address"] = ip.IPAddress
		rawIP["lb_id"] = types.FlattenStringPtr(ip.LBID)
		rawIP["reverse"] = ip.Reverse
		rawIP["zone"] = string(zone)
		rawIP["organization_id"] = ip.OrganizationID
		rawIP["project_id"] = ip.ProjectID

		if len(ip.Tags) > 0 {
			rawIP["tags"] = ip.Tags
		}

		ips = append(ips, rawIP)
	}

	d.SetId(zone.String())
	_ = d.Set("ips", ips)

	return nil
}
