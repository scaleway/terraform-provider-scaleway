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
			"ipam_ip_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The ID of the IPAM IP",
				ValidateFunc:  validationUUIDorUUIDWithLocality(),
				ConflictsWith: []string{"private_network_id", "resource", "mac_address", "type", "tags", "attached"},
			},
			"private_network_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The private Network to filter for",
				ConflictsWith: []string{"ipam_ip_id"},
			},
			"resource": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"ipam_ip_id"},
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
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The MAC address to filter for",
				ConflictsWith: []string{"ipam_ip_id"},
			},
			"type": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "IP Type (ipv4, ipv6)",
				ConflictsWith: []string{"ipam_ip_id"},
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:      true,
				Description:   "The tags associated with the IP",
				ConflictsWith: []string{"ipam_ip_id"},
			},
			"attached": {
				Type:          schema.TypeBool,
				Optional:      true,
				Description:   "Defines whether to filter only for IPs which are attached to a resource",
				ConflictsWith: []string{"ipam_ip_id"},
			},
			"zonal":           zoneSchema(),
			"region":          regionSchema(),
			"project_id":      projectIDSchema(),
			"organization_id": organizationIDSchema(),
			// Computed
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IP address",
			},
			"address_cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IP address with a CIDR notation",
			},
		},
	}
}

func dataSourceScalewayIPAMIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := ipamAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	var address, addressCidr string
	IPID, ok := d.GetOk("ipam_ip_id")
	if !ok {
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
		IPID = ip.ID

		address = ip.Address.IP.String()
		addressCidr, err = flattenIPNet(ip.Address)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		res, err := api.GetIP(&ipam.GetIPRequest{
			Region: region,
			IPID:   expandID(IPID.(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		address = res.Address.IP.String()
		addressCidr, err = flattenIPNet(res.Address)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(IPID.(string))
	_ = d.Set("address", address)
	_ = d.Set("address_cidr", addressCidr)

	return nil
}
