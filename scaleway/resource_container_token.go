package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func resourceScalewayContainerToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayContainerTokenCreate,
		ReadContext:   resourceScalewayContainerTokenRead,
		DeleteContext: resourceScalewayContainerTokenDelete,
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
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"namespace_id": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				ExactlyOneOf:     []string{"container_id"},
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
		CustomizeDiff: CustomizeDiffLocalityCheck("container_id", "namespace_id"),
	}
}

func resourceScalewayContainerTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	token, err := api.CreateToken(&container.CreateTokenRequest{
		Region:      region,
		ContainerID: expandStringPtr(locality.ExpandID(d.Get("container_id"))),
		NamespaceID: expandStringPtr(locality.ExpandID(d.Get("namespace_id"))),
		Description: expandStringPtr(d.Get("description")),
		ExpiresAt:   expandTimePtr(d.Get("expires_at")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, token.ID))

	_ = d.Set("token", token.Token)

	return resourceScalewayContainerTokenRead(ctx, d, m)
}

func resourceScalewayContainerTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, ID, err := containerAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	token, err := api.GetToken(&container.GetTokenRequest{
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

	_ = d.Set("container_id", flattenStringPtr(token.ContainerID))
	_ = d.Set("namespace_id", flattenStringPtr(token.NamespaceID))
	_ = d.Set("description", flattenStringPtr(token.Description))
	_ = d.Set("expires_at", flattenTime(token.ExpiresAt))

	return nil
}

func resourceScalewayContainerTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, ID, err := containerAPIWithRegionAndID(m, d.Id())
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
