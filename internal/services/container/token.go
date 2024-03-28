package container

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceContainerTokenCreate,
		ReadContext:   ResourceContainerTokenRead,
		DeleteContext: ResourceContainerTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"container_id": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				ExactlyOneOf:     []string{"namespace_id"},
				DiffSuppressFunc: dsf.Locality,
			},
			"namespace_id": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				ExactlyOneOf:     []string{"container_id"},
				DiffSuppressFunc: dsf.Locality,
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
				ValidateDiagFunc: verify.IsDate(),
				DiffSuppressFunc: dsf.TimeRFC3339,
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"region": regional.Schema(),
		},
		CustomizeDiff: cdf.LocalityCheck("container_id", "namespace_id"),
	}
}

func ResourceContainerTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	token, err := api.CreateToken(&container.CreateTokenRequest{
		Region:      region,
		ContainerID: types.ExpandStringPtr(locality.ExpandID(d.Get("container_id"))),
		NamespaceID: types.ExpandStringPtr(locality.ExpandID(d.Get("namespace_id"))),
		Description: types.ExpandStringPtr(d.Get("description")),
		ExpiresAt:   types.ExpandTimePtr(d.Get("expires_at")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, token.ID))

	_ = d.Set("token", token.Token)

	return ResourceContainerTokenRead(ctx, d, m)
}

func ResourceContainerTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	token, err := api.GetToken(&container.GetTokenRequest{
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

	_ = d.Set("container_id", types.FlattenStringPtr(token.ContainerID))
	_ = d.Set("namespace_id", types.FlattenStringPtr(token.NamespaceID))
	_ = d.Set("description", types.FlattenStringPtr(token.Description))
	_ = d.Set("expires_at", types.FlattenTime(token.ExpiresAt))

	return nil
}

func ResourceContainerTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, ID, err := NewAPIWithRegionAndID(m, d.Id())
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
