package ipam

import (
	"context"
	"errors"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceIP() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceIPAMIPRead,
		Schema: map[string]*schema.Schema{
			// Input
			"ipam_ip_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The ID of the IPAM IP",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				ConflictsWith:    []string{"private_network_id", "resource", "mac_address", "type", "tags", "attached"},
			},
			"private_network_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The private Network to filter for",
				ConflictsWith: []string{"ipam_ip_id"},
			},
			"resource": {
				Type:          schema.TypeList,
				Description:   "The resource to filter for",
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
				ValidateDiagFunc: func(i any, _ cty.Path) diag.Diagnostics {
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
			"zonal":           zonal.Schema(),
			"region":          regional.Schema(),
			"project_id":      account.ProjectIDSchema(),
			"organization_id": account.OrganizationIDSchema(),
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

func DataSourceIPAMIPRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var address, addressCidr string

	var ip scw.IPNet

	IPID, ok := d.GetOk("ipam_ip_id")
	if !ok {
		resources, resourcesOk := d.GetOk("resource")
		if resourcesOk {
			resourceList := resources.([]any)
			if len(resourceList) > 0 {
				resourceMap := resourceList[0].(map[string]any)
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
			ProjectID:        types.ExpandStringPtr(d.Get("project_id")),
			Zonal:            types.ExpandStringPtr(d.Get("zonal")),
			ResourceID:       types.ExpandStringPtr(expandLastID(d.Get("resource.0.id"))),
			ResourceType:     ipam.ResourceType(d.Get("resource.0.type").(string)),
			ResourceName:     types.ExpandStringPtr(d.Get("resource.0.name").(string)),
			MacAddress:       types.ExpandStringPtr(d.Get("mac_address")),
			Tags:             types.ExpandStrings(d.Get("tags")),
			OrganizationID:   types.ExpandStringPtr(d.Get("organization_id")),
			PrivateNetworkID: types.ExpandStringPtr(locality.ExpandID(d.Get("private_network_id"))),
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
			req.Attached = types.ExpandBoolPtr(attached)
		}

		err = retry.RetryContext(ctx, defaultIPRetryInterval, func() *retry.RetryError {
			resp, err := api.ListIPs(req, scw.WithAllPages(), scw.WithContext(ctx))
			if err != nil {
				return retry.NonRetryableError(err)
			}

			if len(resp.IPs) == 0 {
				// Retry if no IPs are found
				return retry.RetryableError(errors.New("no ip found with given filters"))
			}

			if len(resp.IPs) > 1 {
				return retry.NonRetryableError(errors.New("more than one ip found with given filter"))
			}

			ip = resp.IPs[0].Address
			IPID = resp.IPs[0].ID

			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		res, err := api.GetIP(&ipam.GetIPRequest{
			Region: region,
			IPID:   locality.ExpandID(IPID.(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		ip = res.Address
	}

	address = ip.IP.String()

	addressCidr, err = types.FlattenIPNet(ip)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(IPID.(string))
	_ = d.Set("address", address)
	_ = d.Set("address_cidr", addressCidr)

	return nil
}
