package registry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/registry/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceRegistryImageTag() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceRegistryImageTagRead,

		Schema: map[string]*schema.Schema{
			"tag_id": {
				Type:             schema.TypeString,
				Description:      "The ID of the registry image tag",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			},
			"image_id": {
				Type:             schema.TypeString,
				Description:      "The ID of the registry image",
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the registry image tag",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the registry image tag",
			},
			"digest": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Hash of the tag content. Several tags of a same image may have the same digest",
			},
			"created_at": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Date and time of creation",
			},
			"updated_at": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Date and time of last update",
			},
			"region":          regional.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func DataSourceRegistryImageTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	tagID := d.Get("tag_id").(string)

	res, err := api.GetTag(&registry.GetTagRequest{
		Region: region,
		TagID:  locality.ExpandID(tagID),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	tag := res

	d.SetId(datasource.NewRegionalID(tag.ID, region))
	_ = d.Set("tag_id", tag.ID)
	_ = d.Set("image_id", tag.ImageID)
	_ = d.Set("name", tag.Name)
	_ = d.Set("status", tag.Status.String())
	_ = d.Set("digest", tag.Digest)
	_ = d.Set("created_at", types.FlattenTime(tag.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(tag.UpdatedAt))

	return nil
}
