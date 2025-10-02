package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	vpcSDK "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/vpc"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceIPAMIPCreate,
		ReadContext:   ResourceIPAMIPRead,
		UpdateContext: ResourceIPAMIPUpdate,
		DeleteContext: ResourceIPAMIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"address": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				Description:      "Request a specific IP in the requested source pool",
				ValidateFunc:     validation.IsIPAddress,
				DiffSuppressFunc: DiffSuppressFuncStandaloneIPandCIDR,
			},
			"source": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The source in which to book the IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zonal": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Zone the IP lives in if the IP is a public zoned one",
						},
						"private_network_id": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							Description:      "Private Network the IP lives in if the IP is a private IP",
							DiffSuppressFunc: dsf.Locality,
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Private Network subnet the IP lives in if the IP is a private IP in a Private Network",
						},
					},
				},
			},
			"custom_resource": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The custom resource in which to book the IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mac_address": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "MAC address of the custom resource",
							ValidateFunc: validation.IsMACAddress,
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "When the resource is in a Private Network, a DNS record is available to resolve the resource name",
						},
					},
				},
			},
			"is_ipv6": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Request an IPv6 instead of an IPv4",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The tags associated with the IP",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_id": account.ProjectIDSchema(),
			"region":     regional.Schema(),
			// Computed elements
			"resource": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The IP resource",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of resource the IP is attached to",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the resource the IP is attached to",
						},
						"mac_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "MAC of the resource the IP is attached to",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the resource the IP is attached to",
						},
					},
				},
			},
			"reverses": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The reverses DNS for this IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The reverse domain name",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP corresponding to the hostname",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the IP",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the IP",
			},
			"zone": zonal.ComputedSchema(),
		},
	}
}

func ResourceIPAMIPCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	ipamAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &ipam.BookIPRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		IsIPv6:    d.Get("is_ipv6").(bool),
		Tags:      types.ExpandStrings(d.Get("tags")),
	}

	address, addressOk := d.GetOk("address")
	if addressOk {
		addressStr := address.(string)

		parsedIP, _, err := net.ParseCIDR(addressStr)
		if err != nil {
			parsedIP = net.ParseIP(addressStr)
			if parsedIP == nil {
				return diag.FromErr(fmt.Errorf("error parsing IP address: %w", err))
			}
		}

		req.Address = scw.IPPtr(parsedIP)
	}

	if source, ok := d.GetOk("source"); ok {
		req.Source = expandIPSource(source)
	}

	if customResource, ok := d.GetOk("custom_resource"); ok {
		req.Resource = expandCustomResource(customResource)
	}

	res, err := ipamAPI.BookIP(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	return ResourceIPAMIPRead(ctx, d, m)
}

func ResourceIPAMIPRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	ipamAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	vpcAPI, err := vpc.NewAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := ipamAPI.GetIP(&ipam.GetIPRequest{
		Region: region,
		IPID:   ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	privateNetworkID := ""

	if source, ok := d.GetOk("source"); ok {
		sourceData := expandIPSource(source)
		if sourceData.PrivateNetworkID != nil {
			pn, err := vpcAPI.GetPrivateNetwork(&vpcSDK.GetPrivateNetworkRequest{
				PrivateNetworkID: *sourceData.PrivateNetworkID,
				Region:           region,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			ipv4Subnets, ipv6Subnets := vpc.FlattenAndSortSubnets(pn.Subnets)

			var found bool

			if d.Get("is_ipv6").(bool) {
				found = checkSubnetIDInFlattenedSubnets(*res.Source.SubnetID, ipv6Subnets)
			} else {
				found = checkSubnetIDInFlattenedSubnets(*res.Source.SubnetID, ipv4Subnets)
			}

			if found {
				privateNetworkID = pn.ID
			}
		}
	}

	address, err := types.FlattenIPNet(res.Address)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("address", address)
	_ = d.Set("source", flattenIPSource(res.Source, privateNetworkID))
	_ = d.Set("resource", flattenIPResource(res.Resource))
	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("is_ipv6", res.IsIPv6)
	_ = d.Set("region", region)

	if res.Zone != nil {
		_ = d.Set("zone", res.Zone.String())
	}

	if len(res.Tags) > 0 {
		_ = d.Set("tags", res.Tags)
	}

	_ = d.Set("reverses", flattenIPReverses(res.Reverses))

	return nil
}

func ResourceIPAMIPUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	ipamAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("custom_resource") {
		oldCustomResourceRaw, newCustomResourceRaw := d.GetChange("custom_resource")
		oldCustomResource := expandCustomResource(oldCustomResourceRaw)
		newCustomResource := expandCustomResource(newCustomResourceRaw)

		_, err = ipamAPI.MoveIP(&ipam.MoveIPRequest{
			Region:       region,
			IPID:         ID,
			FromResource: oldCustomResource,
			ToResource:   newCustomResource,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		_, err = ipamAPI.UpdateIP(&ipam.UpdateIPRequest{
			IPID:   ID,
			Region: region,
			Tags:   types.ExpandUpdatedStringsPtr(d.Get("tags")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceIPAMIPRead(ctx, d, m)
}

func ResourceIPAMIPDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	ipamAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if customResource, ok := d.GetOk("custom_resource"); ok {
		_, err = ipamAPI.DetachIP(&ipam.DetachIPRequest{
			Region:   region,
			IPID:     ID,
			Resource: expandCustomResource(customResource),
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	err = ipamAPI.ReleaseIP(&ipam.ReleaseIPRequest{
		Region: region,
		IPID:   ID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
