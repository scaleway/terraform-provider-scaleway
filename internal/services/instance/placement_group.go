package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	instanceSDK "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourcePlacementGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstancePlacementGroupCreate,
		ReadContext:   ResourceInstancePlacementGroupRead,
		UpdateContext: ResourceInstancePlacementGroupUpdate,
		DeleteContext: ResourceInstancePlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultInstancePlacementGroupTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the placement group",
			},
			"policy_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          instanceSDK.PlacementGroupPolicyTypeMaxAvailability.String(),
				Description:      "The operating mode is selected by a policy_type",
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.PlacementGroupPolicyType](),
			},
			"policy_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          instanceSDK.PlacementGroupPolicyModeOptional,
				Description:      "One of the two policy_mode may be selected: enforced or optional.",
				ValidateDiagFunc: verify.ValidateEnum[instanceSDK.PlacementGroupPolicyMode](),
			},
			"policy_respected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Is true when the policy is respected.",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The tags associated with the placement group",
			},
			"zone":            zonal.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func ResourceInstancePlacementGroupCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.CreatePlacementGroup(&instanceSDK.CreatePlacementGroupRequest{
		Zone:       zone,
		Name:       types.ExpandOrGenerateString(d.Get("name"), "pg"),
		Project:    types.ExpandStringPtr(d.Get("project_id")),
		PolicyMode: instanceSDK.PlacementGroupPolicyMode(d.Get("policy_mode").(string)),
		PolicyType: instanceSDK.PlacementGroupPolicyType(d.Get("policy_type").(string)),
		Tags:       types.ExpandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, res.PlacementGroup.ID))

	return ResourceInstancePlacementGroupRead(ctx, d, m)
}

func ResourceInstancePlacementGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.GetPlacementGroup(&instanceSDK.GetPlacementGroupRequest{
		Zone:             zone,
		PlacementGroupID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", res.PlacementGroup.Name)
	_ = d.Set("zone", string(zone))
	_ = d.Set("organization_id", res.PlacementGroup.Organization)
	_ = d.Set("project_id", res.PlacementGroup.Project)
	_ = d.Set("policy_mode", res.PlacementGroup.PolicyMode.String())
	_ = d.Set("policy_type", res.PlacementGroup.PolicyType.String())
	_ = d.Set("policy_respected", res.PlacementGroup.PolicyRespected)
	_ = d.Set("tags", res.PlacementGroup.Tags)

	return nil
}

func ResourceInstancePlacementGroupUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &instanceSDK.UpdatePlacementGroupRequest{
		Zone:             zone,
		PlacementGroupID: ID,
		Tags:             scw.StringsPtr([]string{}),
	}

	hasChanged := false

	if d.HasChange("name") {
		req.Name = types.ExpandStringPtr(d.Get("name").(string))
		hasChanged = true
	}

	if d.HasChange("policy_mode") {
		policyMode := instanceSDK.PlacementGroupPolicyMode(d.Get("policy_mode").(string))
		req.PolicyMode = &policyMode
		hasChanged = true
	}

	if d.HasChange("policy_type") {
		policyType := instanceSDK.PlacementGroupPolicyType(d.Get("policy_type").(string))
		req.PolicyType = &policyType
		hasChanged = true
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
		hasChanged = true
	}

	if hasChanged {
		_, err = instanceAPI.UpdatePlacementGroup(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceInstancePlacementGroupRead(ctx, d, m)
}

func ResourceInstancePlacementGroupDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = instanceAPI.DeletePlacementGroup(&instanceSDK.DeletePlacementGroupRequest{
		Zone:             zone,
		PlacementGroupID: ID,
	}, scw.WithContext(ctx))

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
