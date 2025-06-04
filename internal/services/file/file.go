package file

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceFile() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceFileCreate,
		ReadContext:   ResourceFileRead,
		UpdateContext: ResourceFileUpdate,
		DeleteContext: ResourceFileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultFileTimeout),
			Read:    schema.DefaultTimeout(defaultFileTimeout),
			Delete:  schema.DefaultTimeout(defaultFileTimeout),
			Default: schema.DefaultTimeout(defaultFileTimeout),
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
			"number_of_attachements": {
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

func ResourceFileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := fileAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &file.CreateFileSystemRequest{
		Region: region,
		Name:   types.ExpandOrGenerateString(d.Get("name").(string), "file"),
		ProjectID: d.Get("project_id").(string),
		Size: *types.ExpandUint64Ptr(d.Get("size")),
		Tags: types.ExpandStrings(d.Get("tags")),
	}

	file, err := api.CreateFileSystem(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, file.ID))

	//TODO waitForFile

	return ResourceFileRead(ctx, d, m)
}

func ResourceFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	return nil
}

func ResourceFileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	return ResourceFileRead(ctx, d, m)
}

func ResourceFileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	return nil
}
