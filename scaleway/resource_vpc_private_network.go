package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/errs"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
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
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{Version: 0, Type: vpcPrivateNetworkUpgradeV1SchemaType(), Upgrade: vpcPrivateNetworkV1SUpgradeFunc},
		},
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
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The network address of the subnet in dotted decimal notation, e.g., '192.168.0.0' for a '192.168.0.0/24' subnet",
						},
						"subnet_mask": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The subnet mask expressed in dotted decimal notation, e.g., '255.255.255.0' for a /24 subnet",
						},
						"prefix_length": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The length of the network prefix, e.g., 24 for a 255.255.255.0 mask",
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
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The network address of the subnet in dotted decimal notation, e.g., '192.168.0.0' for a '192.168.0.0/24' subnet",
						},
						"subnet_mask": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The subnet mask expressed in dotted decimal notation, e.g., '255.255.255.0' for a /24 subnet",
						},
						"prefix_length": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The length of the network prefix, e.g., 24 for a 255.255.255.0 mask",
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
				Deprecated:  "This field is deprecated and will be removed in the next major version",
				Description: "Defines whether the private network is Regional. By default, it will be Zonal",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The VPC in which to create the private network",
			},
			"project_id": projectIDSchema(),
			"zone": {
				Type:             schema.TypeString,
				Description:      "The zone you want to attach the resource to",
				Optional:         true,
				Computed:         true,
				Deprecated:       "This field is deprecated and will be removed in the next major version, please use `region` instead",
				ValidateDiagFunc: locality.ValidateStringInSliceWithWarning(zonal.AllZones(), "zone"),
			},
			"region": regional.Schema(),
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

func resourceScalewayVPCPrivateNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ipv4Subnets, ipv6Subnets, err := expandSubnets(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpc.CreatePrivateNetworkRequest{
		Name:      expandOrGenerateString(d.Get("name"), "pn"),
		Tags:      expandStrings(d.Get("tags")),
		ProjectID: d.Get("project_id").(string),
		Region:    region,
	}

	if _, ok := d.GetOk("vpc_id"); ok {
		vpcID := regional.ExpandID(d.Get("vpc_id").(string)).ID
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

	d.SetId(regional.NewIDString(region, pn.ID))

	return resourceScalewayVPCPrivateNetworkRead(ctx, d, m)
}

func resourceScalewayVPCPrivateNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pn, err := vpcAPI.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
		PrivateNetworkID: ID,
		Region:           region,
	}, scw.WithContext(ctx))
	if err != nil {
		if errs.Is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", pn.Name)
	_ = d.Set("vpc_id", regional.NewIDString(region, pn.VpcID))
	_ = d.Set("organization_id", pn.OrganizationID)
	_ = d.Set("project_id", pn.ProjectID)
	_ = d.Set("created_at", flattenTime(pn.CreatedAt))
	_ = d.Set("updated_at", flattenTime(pn.UpdatedAt))
	_ = d.Set("tags", pn.Tags)
	_ = d.Set("region", region)
	_ = d.Set("is_regional", true)
	_ = d.Set("zone", zone)

	ipv4Subnet, ipv6Subnets := flattenAndSortSubnets(pn.Subnets)
	_ = d.Set("ipv4_subnet", ipv4Subnet)
	_ = d.Set("ipv6_subnets", ipv6Subnets)

	return nil
}

func resourceScalewayVPCPrivateNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = vpcAPI.UpdatePrivateNetwork(&vpc.UpdatePrivateNetworkRequest{
		PrivateNetworkID: ID,
		Region:           region,
		Name:             scw.StringPtr(d.Get("name").(string)),
		Tags:             expandUpdatedStringsPtr(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCPrivateNetworkRead(ctx, d, m)
}

func resourceScalewayVPCPrivateNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = retry.RetryContext(ctx, defaultVPCPrivateNetworkRetryInterval, func() *retry.RetryError {
		err := vpcAPI.DeletePrivateNetwork(&vpc.DeletePrivateNetworkRequest{
			PrivateNetworkID: ID,
			Region:           region,
		}, scw.WithContext(ctx))
		if err != nil {
			if errs.Is412Error(err) {
				return retry.RetryableError(err)
			} else if !errs.Is404Error(err) {
				return retry.NonRetryableError(err)
			}
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
