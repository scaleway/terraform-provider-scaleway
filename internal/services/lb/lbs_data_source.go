package lb

import (
	"context"
	"fmt"

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
				Type:        schema.TypeList,
				Description: "List of LBs",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed:    true,
							Description: "UUID of the load balancer",
							Type:        schema.TypeString,
						},
						"description": {
							Computed:    true,
							Description: "Description of the load balancer",
							Type:        schema.TypeString,
						},
						"status": {
							Computed:    true,
							Description: "Status of the load balancer",
							Type:        schema.TypeString,
						},
						"name": {
							Computed:    true,
							Description: "Name of the load balancer",
							Type:        schema.TypeString,
						},
						"type": {
							Computed:    true,
							Description: "Type of the load balancer",
							Type:        schema.TypeString,
						},
						"tags": {
							Computed:    true,
							Description: "List of tags assigned to the load balancer",
							Type:        schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"instances": {
							Type:        schema.TypeList,
							Description: "List of instances assigned to the load balancer",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Description: "UUID of the instance",
										Computed:    true,
									},
									"status": {
										Type:        schema.TypeString,
										Description: "Status of the instance",
										Computed:    true,
									},
									"ip_address": {
										Type:        schema.TypeString,
										Description: "IP address of the instance",
										Computed:    true,
									},
									"created_at": {
										Computed:    true,
										Description: "Date and time when the instance was created (RFC3339)",
										Type:        schema.TypeString,
									},
									"updated_at": {
										Computed:    true,
										Description: "Date and time when the instance was updated (RFC3339)",
										Type:        schema.TypeString,
									},
									"zone": zonal.Schema(),
								},
							},
						},
						"ips": {
							Type:        schema.TypeList,
							Description: "List of IPs assigned to the load balancer",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Description: "UUID of the IP",
										Computed:    true,
									},
									"ip_address": {
										Type:        schema.TypeString,
										Description: "IP address",
										Computed:    true,
									},
									"reverse": {
										Type:        schema.TypeString,
										Description: "Reverse DNS attached to the IP",
										Computed:    true,
									},
									"lb_id": {
										Computed:    true,
										Description: "UUID of the load balancer attached to the IP",
										Type:        schema.TypeString,
									},
									"project_id":      account.ProjectIDSchema(),
									"organization_id": account.OrganizationIDSchema(),
									"zone":            zonal.Schema(),
								},
							},
						},
						"frontend_count": {
							Computed:    true,
							Description: "number of frontends the Load Balancer has.",
							Type:        schema.TypeInt,
						},
						"backend_count": {
							Computed:    true,
							Description: "number of backends the Load Balancer has.",
							Type:        schema.TypeInt,
						},
						"private_network_count": {
							Computed:    true,
							Description: "number of Private Networks attached to the Load Balancer.",
							Type:        schema.TypeInt,
						},
						"route_count": {
							Computed:    true,
							Description: "number of routes configured on the Load Balancer.",
							Type:        schema.TypeInt,
						},
						"subscriber": {
							Computed:    true,
							Description: "Subscriber information.",
							Type:        schema.TypeString,
						},
						"ssl_compatibility_level": {
							Computed: true,
							Description: func() string {
								var t lb.SSLCompatibilityLevel

								values := t.Values()

								return fmt.Sprintf("SSL compatibility level possible values are %s", values)
							}(),
							Type: schema.TypeString,
						},
						"created_at": {
							Computed:    true,
							Description: "Date and time when the load balancer was created (RFC3339)",
							Type:        schema.TypeString,
						},
						"updated_at": {
							Computed:    true,
							Description: "Date and time when the load balancer was created (RFC3339)",
							Type:        schema.TypeString,
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
