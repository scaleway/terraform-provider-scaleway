package scaleway

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewaySecretFolder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewaySecretFolderCreate,
		ReadContext:   resourceScalewaySecretFolderRead,
		DeleteContext: resourceScalewaySecretFolderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The folder name",
				ForceNew:    true,
			},
			"path": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The path where the folder is",
			},
			"region":     regionSchema(),
			"project_id": projectIDSchema(),

			// Computed
			"full_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The full path of the folder",
			},
		},
	}
}

func resourceScalewaySecretFolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := secretAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	folder, err := api.CreateFolder(&secret.CreateFolderRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Name:      expandOrGenerateString(d.Get("name").(string), "folder"),
		Path:      expandStringPtr(d.Get("path")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, folder.ID))

	return resourceScalewaySecretFolderRead(ctx, d, meta)
}

func resourceScalewaySecretFolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := secretAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	folder, err := getSecretFolderByID(ctx, api, region, id)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", folder.Name)
	_ = d.Set("path", folder.Path)
	_ = d.Set("full_path", filepath.Join(folder.Path, folder.Name))
	_ = d.Set("region", region)
	_ = d.Set("project_id", folder.ProjectID)

	return nil
}

func resourceScalewaySecretFolderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := secretAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteFolder(&secret.DeleteFolderRequest{
		Region:   region,
		FolderID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
