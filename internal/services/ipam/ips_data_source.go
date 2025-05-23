package ipam

import (
	"context"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceIPs() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		ReadContext:                       DataSourceIPAMIPsRead,
		Schema: map[string]*schema.Schema{
			"private_network_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The private Network to filter for",
			},
			"attached": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Defines whether to filter only for IPs which are attached to a resource",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The tags associated with the IP to filter for",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"resource": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The IP resource to filter for",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ID of the resource to filter for",
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Type of resource to filter for",
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Name of the resource to filter for",
						},
					},
				},
			},
			"mac_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The MAC address to filter for",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP Type (ipv4, ipv6) to filter for",
			},
			"zonal":           zonal.Schema(),
			"region":          regional.Schema(),
			"project_id":      account.ProjectIDSchema(),
			"organization_id": account.OrganizationIDSchema(),
			// Computed
			"ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"address": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"resource": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"mac_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"tags": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"created_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"updated_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"region":     regional.ComputedSchema(),
						"zone":       zonal.ComputedSchema(),
						"project_id": account.ProjectIDSchema(),
					},
				},
			},
		},
	}
}

func DataSourceIPAMIPsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &ipam.ListIPsRequest{
		Region:           region,
		ProjectID:        types.ExpandStringPtr(d.Get("project_id")),
		Zonal:            types.ExpandStringPtr(d.Get("zonal")),
		PrivateNetworkID: types.ExpandStringPtr(d.Get("private_network_id")),
		ResourceID:       types.ExpandStringPtr(expandLastID(d.Get("resource.0.id"))),
		ResourceType:     ipam.ResourceType(d.Get("resource.0.type").(string)),
		ResourceName:     types.ExpandStringPtr(d.Get("resource.0.name")),
		MacAddress:       types.ExpandStringPtr(d.Get("mac_address")),
		Tags:             types.ExpandStrings(d.Get("tags")),
		OrganizationID:   types.ExpandStringPtr(d.Get("organization_id")),
	}

	attached, attachedExists := d.GetOk("attached")
	if attachedExists {
		req.Attached = types.ExpandBoolPtr(attached)
	}

	ipType, ipTypeExist := d.GetOk("type")
	if ipTypeExist {
		switch ipType.(string) {
		case "ipv4":
			req.IsIPv6 = scw.BoolPtr(false)
		case "ipv6":
			req.IsIPv6 = scw.BoolPtr(true)
		default:
			return diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid IP Type",
				Detail:        "Expected ipv4 or ipv6",
				AttributePath: cty.GetAttrPath("type"),
			}}
		}
	}

	resp, err := api.ListIPs(req, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	ips := []interface{}(nil)

	for _, ip := range resp.IPs {
		address, err := types.FlattenIPNet(ip.Address)
		if err != nil {
			return diag.FromErr(err)
		}

		rawIP := make(map[string]interface{})
		rawIP["id"] = regional.NewIDString(region, ip.ID)
		rawIP["address"] = address
		rawIP["resource"] = flattenIPResource(ip.Resource)
		rawIP["tags"] = ip.Tags
		rawIP["created_at"] = types.FlattenTime(ip.CreatedAt)
		rawIP["updated_at"] = types.FlattenTime(ip.UpdatedAt)
		rawIP["region"] = ip.Region.String()
		rawIP["project_id"] = ip.ProjectID

		if ip.Zone != nil {
			rawIP["zone"] = ip.Zone.String()
		}

		ips = append(ips, rawIP)
	}

	d.SetId(region.String())
	_ = d.Set("ips", ips)

	return nil
}
