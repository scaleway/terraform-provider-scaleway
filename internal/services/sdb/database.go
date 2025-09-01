package sdb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdbSDK "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceDatabaseCreate,
		ReadContext:   ResourceDatabaseRead,
		UpdateContext: ResourceDatabaseUpdate,
		DeleteContext: ResourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultTimeout),
			Read:    schema.DefaultTimeout(defaultTimeout),
			Update:  schema.DefaultTimeout(defaultTimeout),
			Delete:  schema.DefaultTimeout(defaultTimeout),
			Default: schema.DefaultTimeout(defaultTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The database name",
			},
			"max_cpu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     15,
				Description: "The maximum number of CPU units for your Serverless SQL Database",
			},
			"min_cpu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The minimum number of CPU units for your Serverless SQL Database",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "endpoint of the database",
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := api.CreateDatabase(&sdbSDK.CreateDatabaseRequest{
		Region:       region,
		ProjectID:    d.Get("project_id").(string),
		Name:         d.Get("name").(string),
		CPUMin:       uint32(d.Get("min_cpu").(int)),
		CPUMax:       uint32(d.Get("max_cpu").(int)),
		FromBackupID: nil,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, database.ID))

	_, err = waitForDatabase(ctx, api, region, database.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceDatabaseRead(ctx, d, m)
}

func ResourceDatabaseRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := waitForDatabase(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", database.Name)
	_ = d.Set("max_cpu", int(database.CPUMax))
	_ = d.Set("min_cpu", int(database.CPUMin))
	_ = d.Set("endpoint", database.Endpoint)
	_ = d.Set("region", database.Region)
	_ = d.Set("project_id", database.ProjectID)

	return nil
}

func ResourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := waitForDatabase(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	req := &sdbSDK.UpdateDatabaseRequest{
		Region:     region,
		DatabaseID: database.ID,
	}

	if d.HasChange("max_cpu") {
		req.CPUMax = types.ExpandUint32Ptr(d.Get("max_cpu"))
	}

	if d.HasChange("min_cpu") {
		req.CPUMin = types.ExpandUint32Ptr(d.Get("min_cpu"))
	}

	if _, err := api.UpdateDatabase(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceDatabaseRead(ctx, d, m)
}

func ResourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDatabase(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteDatabase(&sdbSDK.DeleteDatabaseRequest{
		Region:     region,
		DatabaseID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForDatabase(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is403(err) {
		return diag.FromErr(err)
	}

	return nil
}
