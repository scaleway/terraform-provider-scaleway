package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceLbs() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceLbsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "LBs with a name like it are listed.",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "LBs with these exact tags are listed",
			},
			"lbs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"description": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"status": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"name": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"type": {
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
						"instances": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"status": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"created_at": {
										Computed: true,
										Type:     schema.TypeString,
									},
									"updated_at": {
										Computed: true,
										Type:     schema.TypeString,
									},
									"zone": zonal.Schema(),
								},
							},
						},
						"ips": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"reverse": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"lb_id": {
										Computed: true,
										Type:     schema.TypeString,
									},
									"project_id":      account.ProjectIDSchema(),
									"organization_id": account.OrganizationIDSchema(),
									"zone":            zonal.Schema(),
								},
							},
						},
						"frontend_count": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"backend_count": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"private_network_count": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"route_count": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"subscriber": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"ssl_compatibility_level": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"created_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"updated_at": {
							Computed: true,
							Type:     schema.TypeString,
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

func DataSourceLbsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := lbAPI.ListLBs(&lb.ZonedAPIListLBsRequest{
		Zone:      zone,
		Name:      types.ExpandStringPtr(d.Get("name")),
		ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		Tags:      types.ExpandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	lbs := []any(nil)

	for _, loadbalancer := range res.LBs {
		rawLb := make(map[string]any)
		rawLb["id"] = zonal.NewID(loadbalancer.Zone, loadbalancer.ID).String()
		rawLb["description"] = loadbalancer.Description
		rawLb["zone"] = string(zone)
		rawLb["name"] = loadbalancer.Name
		rawLb["status"] = loadbalancer.Status
		rawLb["type"] = loadbalancer.Type
		rawLb["frontend_count"] = loadbalancer.FrontendCount
		rawLb["backend_count"] = loadbalancer.BackendCount
		rawLb["private_network_count"] = loadbalancer.PrivateNetworkCount
		rawLb["route_count"] = loadbalancer.RouteCount
		rawLb["organization_id"] = loadbalancer.OrganizationID
		rawLb["project_id"] = loadbalancer.ProjectID
		rawLb["instances"] = flattenLbInstances(loadbalancer.Instances)
		rawLb["ips"] = flattenLbIPs(loadbalancer.IP)

		if len(loadbalancer.Tags) > 0 {
			rawLb["tags"] = loadbalancer.Tags
		}

		lbs = append(lbs, rawLb)
	}

	d.SetId(zone.String())
	_ = d.Set("lbs", lbs)

	return nil
}
