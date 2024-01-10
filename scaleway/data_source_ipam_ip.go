package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayIPAMIP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayIPAMIPRead,
		Schema: map[string]*schema.Schema{
			// Input
			"private_network_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The private Network to filter for",
			},
			"resource": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
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
				Required:    true,
				Description: "IP Type (ipv4, ipv6)",
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					switch i.(string) {
					case "ipv4":
						return nil
					case "ipv6":
						return nil
					default:
						return diag.Diagnostics{{
							Severity:      diag.Error,
							Summary:       "Invalid IP Type",
							Detail:        "Expected ipv4 or ipv6",
							AttributePath: cty.GetAttrPath("type"),
						}}
					}
				},
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the IP",
			},
			"attached": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Defines whether to filter only for IPs which are attached to a resource",
			},
			"zonal":           zoneSchema(),
			"region":          regionSchema(),
			"project_id":      projectIDSchema(),
			"organization_id": organizationIDSchema(),
			// Computed
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceScalewayIPAMIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := ipamAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	resources, resourcesOk := d.GetOk("resource")
	if resourcesOk {
		resourceList := resources.([]interface{})
		if len(resourceList) > 0 {
			resourceMap := resourceList[0].(map[string]interface{})
			id, idExists := resourceMap["id"].(string)
			name, nameExists := resourceMap["name"].(string)

			if (idExists && id == "") && (nameExists && name == "") {
				return diag.Diagnostics{{
					Severity: diag.Error,
					Summary:  "Missing field",
					Detail:   "Either 'id' or 'name' must be provided in 'resource'",
				}}
			}
		}
	}

	req := &ipam.ListIPsRequest{
		Region:           region,
		ProjectID:        expandStringPtr(d.Get("project_id")),
		Zonal:            expandStringPtr(d.Get("zonal")),
		ResourceID:       expandStringPtr(expandLastID(d.Get("resource.0.id"))),
		ResourceType:     ipam.ResourceType(d.Get("resource.0.type").(string)),
		ResourceName:     expandStringPtr(d.Get("resource.0.name").(string)),
		MacAddress:       expandStringPtr(d.Get("mac_address")),
		Tags:             expandStrings(d.Get("tags")),
		OrganizationID:   expandStringPtr(d.Get("organization_id")),
		PrivateNetworkID: expandStringPtr(d.Get("private_network_id")),
	}

	switch d.Get("type").(string) {
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

	attached, attachedExists := d.GetOk("attached")
	if attachedExists {
		req.Attached = expandBoolPtr(attached)
	}

	resp, err := api.ListIPs(req, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(resp.IPs) == 0 {
		return diag.FromErr(fmt.Errorf("no ip found with given filters"))
	}
	if len(resp.IPs) > 1 {
		return diag.FromErr(fmt.Errorf("more than one ip found with given filter"))
	}

	ip := resp.IPs[0]

	d.SetId(ip.ID)
	_ = d.Set("address", ip.Address.IP.String())

	return nil
}
