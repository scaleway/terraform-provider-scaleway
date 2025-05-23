package rdb

import (
	"context"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceDatabaseBackup() *schema.Resource {
	return &schema.Resource{
		EnableLegacyTypeSystemApplyErrors: true,
		EnableLegacyTypeSystemPlanErrors:  true,
		CreateContext:                     ResourceRdbDatabaseBackupCreate,
		ReadContext:                       ResourceRdbDatabaseBackupRead,
		UpdateContext:                     ResourceRdbDatabaseBackupUpdate,
		DeleteContext:                     ResourceRdbDatabaseBackupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInstanceTimeout),
			Read:    schema.DefaultTimeout(defaultInstanceTimeout),
			Update:  schema.DefaultTimeout(defaultInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceTimeout),
			Default: schema.DefaultTimeout(defaultInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "Instance on which the user is created",
			},
			"database_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the database of this backup.",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the backup.",
				Optional:    true,
				Computed:    true,
			},
			"size": {
				Type:        schema.TypeInt,
				Description: "Size of the backup (in bytes).",
				Computed:    true,
			},
			"instance_name": {
				Type:        schema.TypeString,
				Description: "Name of the instance of the backup.",
				Computed:    true,
			},
			"expires_at": {
				Type:             schema.TypeString,
				Description:      "Expiration date (Format ISO 8601). Cannot be removed.",
				Optional:         true,
				ValidateDiagFunc: verify.IsDate(),
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Creation date (Format ISO 8601).",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Updated date (Format ISO 8601).",
				Computed:    true,
			},
			// Common
			"region": regional.Schema(),
		},
		CustomizeDiff: cdf.LocalityCheck("instance_id"),
	}
}

func ResourceRdbDatabaseBackupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := d.Get("instance_id").(string)

	createReq := &rdb.CreateDatabaseBackupRequest{
		Region:       region,
		InstanceID:   locality.ExpandID(instanceID),
		DatabaseName: d.Get("database_name").(string),
		Name:         types.ExpandOrGenerateString(d.Get("name"), "backup"),
		ExpiresAt:    types.ExpandTimePtr(d.Get("expires_at")),
	}

	dbBackup, err := rdbAPI.CreateDatabaseBackup(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, dbBackup.ID))

	_, err = waitForRDBDatabaseBackup(ctx, rdbAPI, region, dbBackup.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceRdbDatabaseBackupRead(ctx, d, m)
}

func ResourceRdbDatabaseBackupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	dbBackup, err := waitForRDBDatabaseBackup(ctx, rdbAPI, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("instance_id", regional.NewID(region, dbBackup.InstanceID).String())
	_ = d.Set("name", dbBackup.Name)
	_ = d.Set("database_name", dbBackup.DatabaseName)
	_ = d.Set("instance_name", dbBackup.InstanceName)
	_ = d.Set("expires_at", types.FlattenTime(dbBackup.ExpiresAt))
	_ = d.Set("created_at", types.FlattenTime(dbBackup.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(dbBackup.UpdatedAt))
	_ = d.Set("size", types.FlattenSize(dbBackup.Size))
	_ = d.Set("region", dbBackup.Region)

	d.SetId(regional.NewIDString(region, dbBackup.ID))

	return nil
}

func ResourceRdbDatabaseBackupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("expires_at") && d.Get("expires_at").(string) == "" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Invalid expires_at",
				Detail:        "You cannot remove expires_at after it was set.",
				AttributePath: cty.GetAttrPath("expires_at"),
			},
		}
	}

	req := &rdb.UpdateDatabaseBackupRequest{
		Region:           region,
		DatabaseBackupID: id,
		Name:             types.ExpandStringPtr(d.Get("name")),
		ExpiresAt:        types.ExpandTimePtr(d.Get("expires_at")),
	}

	_, err = rdbAPI.UpdateDatabaseBackup(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBDatabaseBackup(ctx, rdbAPI, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceRdbDatabaseBackupRead(ctx, d, m)
}

func ResourceRdbDatabaseBackupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBDatabaseBackup(ctx, rdbAPI, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = rdbAPI.DeleteDatabaseBackup(&rdb.DeleteDatabaseBackupRequest{
		DatabaseBackupID: id,
		Region:           region,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	_, err = waitForRDBDatabaseBackup(ctx, rdbAPI, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
