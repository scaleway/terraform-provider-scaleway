package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
)

func dataSourceScalewayRDBDatabaseBackup() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRdbDatabaseBackup().Schema)

	addOptionalFieldsToSchema(dsSchema, "name", "region", "instance_id", "project_id")

	dsSchema["instance_id"].RequiredWith = []string{"name"}
	dsSchema["backup_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Backup",
		ConflictsWith: []string{"name", "instance_id"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayRDBDatabaseBackupRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRDBDatabaseBackupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := rdbAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	backupID, backupIDExists := d.GetOk("backup_id")
	if !backupIDExists {
		backupName := d.Get("name").(string)
		res, err := api.ListDatabaseBackups(&rdb.ListDatabaseBackupsRequest{
			Region:     region,
			Name:       expandStringPtr(backupName),
			InstanceID: expandStringPtr(expandID(d.Get("instance_id"))),
			ProjectID:  expandStringPtr(d.Get("project_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		foundBackup, err := findExact(
			res.DatabaseBackups,
			func(s *rdb.DatabaseBackup) bool { return s.Name == backupName },
			backupName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		backupID = foundBackup.ID
	}

	regionID := datasourceNewRegionalID(backupID, region)
	d.SetId(regionID)
	err = d.Set("backup_id", regionID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayRdbDatabaseBackupRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read database backup state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("database backup (%s) not found", regionID)
	}

	return nil
}
