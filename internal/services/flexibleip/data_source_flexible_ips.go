package flexibleip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	flexibleip "github.com/scaleway/scaleway-sdk-go/api/flexibleip/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceFlexibleIPs() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceFlexibleIPsRead,
		Schema: map[string]*schema.Schema{
			"server_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Flexible IPs that are attached to these server IDs are listed",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Flexible IPs with these exact tags are listed",
			},
			"ips": {
				Type:        schema.TypeList,
				Description: "List of flexible IPs",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed:    true,
							Description: "ID of the flexible IP",
							Type:        schema.TypeString,
						},
						"description": {
							Computed:    true,
							Description: "Description of the flexible IP",
							Type:        schema.TypeString,
						},
						"tags": {
							Computed:    true,
							Description: "List of flexible IP tags",
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"status": {
							Computed:    true,
							Description: "Status of the flexible IP",
							Type:        schema.TypeString,
						},
						"ip_address": {
							Computed:    true,
							Description: "IP address of the flexible IP",
							Type:        schema.TypeString,
						},
						"reverse": {
							Computed:    true,
							Description: "Reverse DNS of the flexible IP",
							Type:        schema.TypeString,
						},
						"mac_address": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The MAC address of the server associated with this flexible IP",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "MAC address ID",
									},
									"mac_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "MAC address of the Virtual MAC",
									},
									"mac_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Type of virtual MAC",
									},
									"status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Status of virtual MAC",
									},
									"created_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The date on which the virtual MAC was created (RFC 3339 format)",
									},
									"updated_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The date on which the virtual MAC was last updated (RFC 3339 format)",
									},
									"zone": zonal.Schema(),
								},
							},
						},
						"created_at": {
							Computed:    true,
							Description: "Date on which the flexible IP was created (RFC 3339 format)",
							Type:        schema.TypeString,
						},
						"updated_at": {
							Computed:    true,
							Description: "Date on which the flexible IP was last updated (RFC 3339 format)",
							Type:        schema.TypeString,
						},
						"zone":            zonal.ComputedSchema(),
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

func DataSourceFlexibleIPsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	fipAPI, zone, err := fipAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := fipAPI.ListFlexibleIPs(&flexibleip.ListFlexibleIPsRequest{
		Zone:      zone,
		ServerIDs: locality.ExpandIDs(d.Get("server_ids")),
		ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		Tags:      types.ExpandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	fips := []any(nil)

	for _, fip := range res.FlexibleIPs {
		rawFip := make(map[string]any)
		rawFip["id"] = zonal.NewID(fip.Zone, fip.ID).String()
		rawFip["organization_id"] = fip.OrganizationID
		rawFip["project_id"] = fip.ProjectID
		rawFip["description"] = fip.Description

		if len(fip.Tags) > 0 {
			rawFip["tags"] = fip.Tags
		}

		rawFip["created_at"] = types.FlattenTime(fip.CreatedAt)
		rawFip["updated_at"] = types.FlattenTime(fip.UpdatedAt)
		rawFip["status"] = fip.Status

		ip, err := types.FlattenIPNet(fip.IPAddress)
		if err != nil {
			return diag.FromErr(err)
		}

		rawFip["ip_address"] = ip
		if fip.MacAddress != nil {
			rawFip["mac_address"] = flattenFlexibleIPMacAddress(fip.MacAddress)
		}

		rawFip["reverse"] = fip.Reverse
		rawFip["zone"] = string(zone)

		fips = append(fips, rawFip)
	}

	d.SetId(zone.String())
	_ = d.Set("ips", fips)

	return nil
}
