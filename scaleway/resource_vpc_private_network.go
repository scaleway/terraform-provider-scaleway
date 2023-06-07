package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	v1 "github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	v2 "github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayVPCPrivateNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCPrivateNetworkCreate,
		ReadContext:   resourceScalewayVPCPrivateNetworkRead,
		UpdateContext: resourceScalewayVPCPrivateNetworkUpdate,
		DeleteContext: resourceScalewayVPCPrivateNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the private network",
				Computed:    true,
			},
			"ipv4_subnet": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The IPv4 subnet associated with the private network",
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							Description:  "The subnet CIDR",
							ValidateFunc: validation.IsCIDRNetwork(0, 32),
						},
						// computed
						"id": {
							Type:        schema.TypeString,
							Description: "The subnet ID",
							Computed:    true,
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the creation of the subnet",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the last update of the subnet",
						},
					},
				},
			},
			"ipv6_subnets": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The IPv6 subnet associated with the private network",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							Description:  "The subnet CIDR",
							ValidateFunc: validation.IsCIDRNetwork(0, 128),
						},
						// computed
						"id": {
							Type:        schema.TypeString,
							Description: "The subnet ID",
							Computed:    true,
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the creation of the subnet",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date and time of the last update of the subnet",
						},
					},
				},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The tags associated with private network",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_regional": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Defines whether the private network is Regional. By default, it will be Zonal",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The VPC in which to create the private network",
			},
			"project_id": projectIDSchema(),
			"zone":       zoneSchema(),
			"region":     regionSchema(),
			// Computed elements
			"organization_id": organizationIDSchema(),
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the private network",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the private network",
			},
		},
	}
}

func resourceScalewayVPCPrivateNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("is_regional").(bool) {
		return resourceScalewayVPCPrivateNetworkRegionalCreate(ctx, d, meta)
	}
	return resourceScalewayVPCPrivateNetworkZonalCreate(ctx, d, meta)
}

func resourceScalewayVPCPrivateNetworkZonalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, zone, err := vpcAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	ipv4Subnets, ipv6Subnets, err := expandSubnets(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &v1.CreatePrivateNetworkRequest{
		Name:      expandOrGenerateString(d.Get("name"), "pn"),
		Tags:      expandStrings(d.Get("tags")),
		ProjectID: d.Get("project_id").(string),
		Zone:      zone,
	}

	if ipv4Subnets != nil {
		req.Subnets = append(req.Subnets, ipv4Subnets...)
	}

	if ipv6Subnets != nil {
		req.Subnets = append(req.Subnets, ipv6Subnets...)
	}

	pn, err := vpcAPI.CreatePrivateNetwork(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, pn.ID))

	return resourceScalewayVPCPrivateNetworkZonalRead(ctx, d, meta)
}

func resourceScalewayVPCPrivateNetworkRegionalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	ipv4Subnets, ipv6Subnets, err := expandSubnets(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &v2.CreatePrivateNetworkRequest{
		Name:      expandOrGenerateString(d.Get("name"), "pn"),
		Tags:      expandStrings(d.Get("tags")),
		ProjectID: d.Get("project_id").(string),
		Region:    region,
	}

	if _, ok := d.GetOk("vpc_id"); ok {
		vpcID := expandRegionalID(d.Get("vpc_id").(string)).ID
		req.VpcID = expandUpdatedStringPtr(vpcID)
	}

	if ipv4Subnets != nil {
		req.Subnets = append(req.Subnets, ipv4Subnets...)
	}

	if ipv6Subnets != nil {
		req.Subnets = append(req.Subnets, ipv6Subnets...)
	}

	pn, err := vpcAPI.CreatePrivateNetwork(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, pn.ID))

	return resourceScalewayVPCPrivateNetworkRegionalRead(ctx, d, meta)
}

func resourceScalewayVPCPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("is_regional").(bool) {
		return resourceScalewayVPCPrivateNetworkRegionalRead(ctx, d, meta)
	}
	return resourceScalewayVPCPrivateNetworkZonalRead(ctx, d, meta)
}

func resourceScalewayVPCPrivateNetworkZonalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, zone, ID, err := vpcAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pn, err := vpcAPI.GetPrivateNetwork(&v1.GetPrivateNetworkRequest{
		PrivateNetworkID: ID,
		Zone:             zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", pn.Name)
	_ = d.Set("organization_id", pn.OrganizationID)
	_ = d.Set("project_id", pn.ProjectID)
	_ = d.Set("created_at", pn.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", pn.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("tags", pn.Tags)
	_ = d.Set("zone", zone)
	_ = d.Set("is_regional", false)

	ipv4Subnet, ipv6Subnets := flattenAndSortSubnets(pn.Subnets)
	_ = d.Set("ipv4_subnet", ipv4Subnet)
	_ = d.Set("ipv6_subnets", ipv6Subnets)

	return nil
}

func resourceScalewayVPCPrivateNetworkRegionalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pn, err := vpcAPI.GetPrivateNetwork(&v2.GetPrivateNetworkRequest{
		PrivateNetworkID: ID,
		Region:           region,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", pn.Name)
	_ = d.Set("vpc_id", newRegionalIDString(region, pn.VpcID))
	_ = d.Set("organization_id", pn.OrganizationID)
	_ = d.Set("project_id", pn.ProjectID)
	_ = d.Set("created_at", flattenTime(pn.CreatedAt))
	_ = d.Set("updated_at", flattenTime(pn.UpdatedAt))
	_ = d.Set("tags", pn.Tags)
	_ = d.Set("region", region)
	_ = d.Set("is_regional", true)

	ipv4Subnet, ipv6Subnets := flattenAndSortSubnets(pn.Subnets)
	_ = d.Set("ipv4_subnet", ipv4Subnet)
	_ = d.Set("ipv6_subnets", ipv6Subnets)

	return nil
}

func resourceScalewayVPCPrivateNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("is_regional").(bool) {
		return resourceScalewayVPCPrivateNetworkRegionalUpdate(ctx, d, meta)
	}
	return resourceScalewayVPCPrivateNetworkZonalUpdate(ctx, d, meta)
}

func resourceScalewayVPCPrivateNetworkZonalUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, zone, ID, err := vpcAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("vpc_id"); ok {
		if !d.Get("is_regional").(bool) {
			return diag.Errorf("vpc_id can only be set if is_regional is set to true")
		}
	}

	_, err = vpcAPI.UpdatePrivateNetwork(&v1.UpdatePrivateNetworkRequest{
		PrivateNetworkID: ID,
		Zone:             zone,
		Name:             scw.StringPtr(d.Get("name").(string)),
		Tags:             expandUpdatedStringsPtr(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCPrivateNetworkZonalRead(ctx, d, meta)
}

func resourceScalewayVPCPrivateNetworkRegionalUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = vpcAPI.UpdatePrivateNetwork(&v2.UpdatePrivateNetworkRequest{
		PrivateNetworkID: ID,
		Region:           region,
		Name:             scw.StringPtr(d.Get("name").(string)),
		Tags:             expandUpdatedStringsPtr(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCPrivateNetworkRegionalRead(ctx, d, meta)
}

func resourceScalewayVPCPrivateNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("is_regional").(bool) {
		return resourceScalewayVPCPrivateNetworkRegionalDelete(ctx, d, meta)
	}
	return resourceScalewayVPCPrivateNetworkZonalDelete(ctx, d, meta)
}

func resourceScalewayVPCPrivateNetworkZonalDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, zone, ID, err := vpcAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var warnings diag.Diagnostics
	err = vpcAPI.DeletePrivateNetwork(&v1.DeletePrivateNetworkRequest{
		PrivateNetworkID: ID,
		Zone:             zone,
	}, scw.WithContext(ctx))
	if err != nil {
		if is409Error(err) || is412Error(err) || is404Error(err) {
			return append(warnings, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  err.Error(),
			})
		}
		return diag.FromErr(err)
	}

	return nil
}

func resourceScalewayVPCPrivateNetworkRegionalDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var warnings diag.Diagnostics
	err = vpcAPI.DeletePrivateNetwork(&v2.DeletePrivateNetworkRequest{
		PrivateNetworkID: ID,
		Region:           region,
	}, scw.WithContext(ctx))
	if err != nil {
		if is409Error(err) || is412Error(err) || is404Error(err) {
			return append(warnings, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  err.Error(),
			})
		}
		return diag.FromErr(err)
	}

	return nil
}
