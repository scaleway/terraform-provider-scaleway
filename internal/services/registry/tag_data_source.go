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

func DataSourceImageTag() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceImageTagRead,

		Schema: map[string]*schema.Schema{
			"tag_id": {
				Type:             schema.TypeString,
				Description:      "The ID of the registry image tag",
				Optional:         true,
				ConflictsWith:    []string{"name"},
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			},
			"image_id": {
				Type:             schema.TypeString,
				Description:      "The ID of the registry image",
				Required:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of the registry image tag",
				ConflictsWith: []string{"tag_id"},
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

func DataSourceImageTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := NewAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var tag *registry.Tag
	tagID, tagIDExists := d.GetOk("tag_id")
	imageID := d.Get("image_id").(string)

	if tagIDExists {
		res, err := api.GetTag(&registry.GetTagRequest{
			Region: region,
			TagID:  locality.ExpandID(tagID),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		tag = res
	} else {
		tagName, nameExists := d.GetOk("name")
		if !nameExists {
			return diag.Errorf("either 'tag_id' or 'name' must be provided")
		}

		res, err := api.ListTags(&registry.ListTagsRequest{
			Region:  region,
			ImageID: locality.ExpandID(imageID),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundTag, err := datasource.FindExact(res.Tags, func(s *registry.Tag) bool {
			return s.Name == tagName.(string)
		}, tagName.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		tag = foundTag
	}

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
