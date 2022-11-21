package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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

			"region": regionSchema(),
		},
	}
}

func resourceScalewayFunctionTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	token, err := api.CreateToken(&function.CreateTokenRequest{
		Region:      region,
		FunctionID:  expandStringPtr(expandID(d.Get("function_id"))),
		NamespaceID: expandStringPtr(expandID(d.Get("namespace_id"))),
		Description: expandStringPtr(d.Get("description")),
		ExpiresAt:   expandTimePtr(d.Get("expires_at")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, token.ID))

	return resourceScalewayFunctionTokenRead(ctx, d, meta)
}

func resourceScalewayFunctionTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, ID, err := functionAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	token, err := api.GetToken(&function.GetTokenRequest{
		Region:  region,
		TokenID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("function_id", flattenStringPtr(token.FunctionID))
	_ = d.Set("namespace_id", flattenStringPtr(token.NamespaceID))
	_ = d.Set("description", flattenStringPtr(token.Description))
	_ = d.Set("expires_at", flattenTime(token.ExpiresAt))
	_ = d.Set("token", token.Token)

	return nil
}

func resourceScalewayFunctionTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, ID, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteToken(&container.DeleteTokenRequest{
		Region:  region,
		TokenID: ID,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
