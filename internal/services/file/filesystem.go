package file

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceFileSystem() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceFileSystemCreate,
		ReadContext:   ResourceFileSystemRead,
		UpdateContext: ResourceFileSystemUpdate,
		DeleteContext: ResourceFileSystemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultFileSystemTimeout),
			Read:    schema.DefaultTimeout(defaultFileSystemTimeout),
			Delete:  schema.DefaultTimeout(defaultFileSystemTimeout),
			Default: schema.DefaultTimeout(defaultFileSystemTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the filesystem",
			},
			"size": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The Filesystem size in bytes, with a granularity of 100 GB (10^11 bytes). Must be compliant with the minimum (100 GB) and maximum (10 TB) allowed size.",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The list of tags assigned to the filesystem",
			},
			"project_id":      account.ProjectIDSchema(),
			"organization_id": account.OrganizationIDSchema(),
			"region":          regional.Schema(),
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Current status of the filesystem (e.g. creating, available, ...)",
			},
			"number_of_attachments": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The current number of attachments (mounts) that the filesystem has",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation date of the filesystem",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update date of the properties of the filesystem",
			},
		},
	}
}

func ResourceFileSystemCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := fileSystemAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &file.CreateFileSystemRequest{
		Region:    region,
		Name:      types.ExpandOrGenerateString(d.Get("name").(string), "file"),
		ProjectID: d.Get("project_id").(string),
		Size:      *types.ExpandUint64Ptr(d.Get("size")),
		Tags:      types.ExpandStrings(d.Get("tags")),
	}

	file, err := api.CreateFileSystem(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, file.ID))

	_, err = waitForFileSystem(ctx, api, region, file.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceFileSystemRead(ctx, d, m)
}

func ResourceFileSystemRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	fileSystem, err := waitForFileSystem(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", fileSystem.Name)
	_ = d.Set("project_id", fileSystem.ProjectID)
	_ = d.Set("region", fileSystem.Region)
	_ = d.Set("organization_id", fileSystem.OrganizationID)
	_ = d.Set("status", fileSystem.Status)
	_ = d.Set("size", int64(fileSystem.Size))
	_ = d.Set("tags", fileSystem.Tags)
	_ = d.Set("created_at", fileSystem.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", fileSystem.UpdatedAt.Format(time.RFC3339))
	_ = d.Set("number_of_attachments", int64(fileSystem.NumberOfAttachments))

	return nil
}

func ResourceFileSystemUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	fileSystem, err := waitForFileSystem(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	req := &file.UpdateFileSystemRequest{
		Region:       region,
		FilesystemID: fileSystem.ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("size") {
		req.Size = types.ExpandUint64Ptr(d.Get("size"))
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandStringsPtr(d.Get("tags"))
	}

	if _, err := api.UpdateFileSystem(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceFileSystemRead(ctx, d, m)
}

func ResourceFileSystemDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFileSystem(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	err = api.DeleteFileSystem(&file.DeleteFileSystemRequest{
		Region:       region,
		FilesystemID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFileSystem(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
