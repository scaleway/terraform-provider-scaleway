package vpc

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

var (
	ErrVPCRoutingCannotBeDisabled = errors.New("routing cannot be disabled on this VPC")
)

func ResourceVPC() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceVPCCreate,
		ReadContext:   ResourceVPCRead,
		UpdateContext: ResourceVPCUpdate,
		DeleteContext: ResourceVPCDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		SchemaFunc:    vpcSchema,
		Identity:      identity.DefaultRegional(),
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
			before, after := diff.GetChange("enable_routing")
			if before != nil && before.(bool) && after != nil && !after.(bool) {
				return ErrVPCRoutingCannotBeDisabled
			}

			return nil
		},
	}
}

func vpcSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		"enable_routing": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Enable routing between Private Networks in the VPC",
		},
		"enable_custom_routes_propagation": {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Defines whether the VPC advertises custom routes between its Private Networks",
		},
		"project_id": account.ProjectIDSchema(),
		"region":     regional.Schema(),
		// Computed elements
		"organization_id": account.OrganizationIDSchema(),
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
	}
}

func ResourceVPCCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := vpcAPI.CreateVPC(&vpc.CreateVPCRequest{
		Name:          types.ExpandOrGenerateString(d.Get("name"), "vpc"),
		Tags:          types.ExpandStrings(d.Get("tags")),
		EnableRouting: d.Get("enable_routing").(bool),
		ProjectID:     d.Get("project_id").(string),
		Region:        region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("enable_custom_routes_propagation"); ok {
		_, err = vpcAPI.EnableCustomRoutesPropagation(&vpc.EnableCustomRoutesPropagationRequest{
			Region: region,
			VpcID:  res.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = identity.SetRegionalIdentity(d, region, res.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceVPCRead(ctx, d, m)
}

func ResourceVPCRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
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

	diags := setVPCState(d, res, region)

	err = identity.SetRegionalIdentity(d, region, ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func setVPCState(d *schema.ResourceData, res *vpc.VPC, region scw.Region) diag.Diagnostics {
	_ = d.Set("name", res.Name)
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("project_id", res.ProjectID)
	_ = d.Set("created_at", types.FlattenTime(res.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(res.UpdatedAt))
	_ = d.Set("is_default", res.IsDefault)
	_ = d.Set("enable_routing", res.RoutingEnabled)
	_ = d.Set("enable_custom_routes_propagation", res.CustomRoutesPropagationEnabled)
	_ = d.Set("region", region)

	if len(res.Tags) > 0 {
		_ = d.Set("tags", res.Tags)
	}

	return nil
}

func ResourceVPCUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	hasChanged := false

	updateRequest := &vpc.UpdateVPCRequest{
		Region: region,
		VpcID:  ID,
	}

	if d.HasChange("name") {
		updateRequest.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
		hasChanged = true
	}

	if d.HasChange("tags") {
		updateRequest.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err = vpcAPI.UpdateVPC(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("enable_routing") {
		enableRouting := d.Get("enable_routing").(bool)
		if enableRouting {
			_, err = vpcAPI.EnableRouting(&vpc.EnableRoutingRequest{
				Region: region,
				VpcID:  ID,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("enable_custom_routes_propagation") {
		enableCustomRoutesPropagation := d.Get("enable_custom_routes_propagation").(bool)
		if enableCustomRoutesPropagation {
			_, err = vpcAPI.EnableCustomRoutesPropagation(&vpc.EnableCustomRoutesPropagationRequest{
				Region: region,
				VpcID:  ID,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return ResourceVPCRead(ctx, d, m)
}

func ResourceVPCDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	vpcAPI, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
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
