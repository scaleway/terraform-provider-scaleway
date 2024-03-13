package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func resourceScalewayVPC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayVPCCreate,
		ReadContext:   resourceScalewayVPCRead,
		UpdateContext: resourceScalewayVPCUpdate,
		DeleteContext: resourceScalewayVPCDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the VPC",
				Computed:    true,
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The tags associated with the VPC",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"project_id": projectIDSchema(),
			"region":     regional.Schema(),
			// Computed elements
			"organization_id": organizationIDSchema(),
			"is_default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines whether the VPC is the default one for its Project",
			},
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

func resourceScalewayVPCCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.CreateVPC(&vpc.CreateVPCRequest{
		Name:      types.ExpandOrGenerateString(d.Get("name"), "vpc"),
		Tags:      types.ExpandStrings(d.Get("tags")),
		ProjectID: d.Get("project_id").(string),
		Region:    region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))

	return resourceScalewayVPCRead(ctx, d, m)
}

func resourceScalewayVPCRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.GetVPC(&vpc.GetVPCRequest{
		Region: region,
		VpcID:  ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("is_default", res.IsDefault)
	_ = d.Set("region", region)

	if len(res.Tags) > 0 {
		_ = d.Set("tags", res.Tags)
	}

	return nil
}

func resourceScalewayVPCUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = vpcAPI.UpdateVPC(&vpc.UpdateVPCRequest{
		VpcID:  ID,
		Region: region,
		Name:   scw.StringPtr(d.Get("name").(string)),
		Tags:   types.ExpandUpdatedStringsPtr(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayVPCRead(ctx, d, m)
}

func resourceScalewayVPCDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vpcAPI, region, ID, err := vpcAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = vpcAPI.DeleteVPC(&vpc.DeleteVPCRequest{
		Region: region,
		VpcID:  ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
