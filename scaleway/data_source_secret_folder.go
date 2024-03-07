package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
)

func dataSourceScalewaySecretFolder() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewaySecretFolder().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["folder_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the folder",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewaySecretFolderRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewaySecretFolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := secretAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	folderID, folderIDExists := d.GetOk("folder_id")
	if !folderIDExists {
		res, err := api.ListFolders(&secret.ListFoldersRequest{
			Region: region,
			//Name:       expandStringPtr(d.Get("name")), TODO
			ProjectID: expandStringPtr(d.Get("project_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		for _, folder := range res.Folders {
			if folder.Name == d.Get("name").(string) {
				if folderID != "" {
					return diag.Errorf("more than 1 folder found with the same name %s", d.Get("name"))
				}
				folderID = folder.ID
			}
		}
		if folderID == "" {
			return diag.Errorf("no folder found with the name %s", d.Get("name"))
		}
	}

	regionID := datasourceNewRegionalID(folderID, region)
	d.SetId(regionID)
	err = d.Set("folder_id", regionID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewaySecretFolderRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read folder state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("folder (%s) not found", regionID)
	}

	return nil
}
