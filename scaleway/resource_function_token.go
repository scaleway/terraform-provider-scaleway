package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func resourceScalewayFunctionToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayFunctionTokenCreate,
		ReadContext:   resourceScalewayFunctionTokenRead,
		DeleteContext: resourceScalewayFunctionTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				ExactlyOneOf:     []string{"namespace_id"},
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"namespace_id": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				ExactlyOneOf:     []string{"function_id"},
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"expires_at": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateDate(),
				DiffSuppressFunc: diffSuppressFuncTimeRFC3339,
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"region": regional.Schema(),
		},
		CustomizeDiff: CustomizeDiffLocalityCheck("function_id", "namespace_id"),
	}
}

func resourceScalewayFunctionTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	token, err := api.CreateToken(&function.CreateTokenRequest{
		Region:      region,
		FunctionID:  types.ExpandStringPtr(locality.ExpandID(d.Get("function_id"))),
		NamespaceID: types.ExpandStringPtr(locality.ExpandID(d.Get("namespace_id"))),
		Description: types.ExpandStringPtr(d.Get("description")),
		ExpiresAt:   expandTimePtr(d.Get("expires_at")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, token.ID))

	_ = d.Set("token", token.Token)

	return resourceScalewayFunctionTokenRead(ctx, d, m)
}

func resourceScalewayFunctionTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, ID, err := FunctionAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	token, err := api.GetToken(&function.GetTokenRequest{
		Region:  region,
		TokenID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("function_id", types.FlattenStringPtr(token.FunctionID))
	_ = d.Set("namespace_id", types.FlattenStringPtr(token.NamespaceID))
	_ = d.Set("description", types.FlattenStringPtr(token.Description))
	_ = d.Set("expires_at", types.FlattenTime(token.ExpiresAt))

	return nil
}

func resourceScalewayFunctionTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, ID, err := ContainerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteToken(&container.DeleteTokenRequest{
		Region:  region,
		TokenID: ID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
