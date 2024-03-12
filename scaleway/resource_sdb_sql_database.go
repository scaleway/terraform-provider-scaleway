package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	serverless_sqldb "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func resourceScalewaySDBSQLDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayServerlessSQLDBDatabaseCreate,
		ReadContext:   resourceScalewayServerlessSQLDBDatabaseRead,
		UpdateContext: resourceScalewayServerlessSQLDBDatabaseUpdate,
		DeleteContext: resourceScalewayServerlessSQLDBDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultSDBSQLTimeout),
			Read:    schema.DefaultTimeout(defaultSDBSQLTimeout),
			Update:  schema.DefaultTimeout(defaultSDBSQLTimeout),
			Delete:  schema.DefaultTimeout(defaultSDBSQLTimeout),
			Default: schema.DefaultTimeout(defaultSDBSQLTimeout),
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
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayServerlessSQLDBDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := serverlessSQLdbAPIWithRegion(d, m.(*meta.Meta))
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := api.CreateDatabase(&serverless_sqldb.CreateDatabaseRequest{
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

	_, err = waitForServerlessSQLDBDatabase(ctx, api, region, database.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayServerlessSQLDBDatabaseRead(ctx, d, m)
}

func resourceScalewayServerlessSQLDBDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := serverlessSQLdbAPIWithRegionAndID(m.(*meta.Meta), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := waitForServerlessSQLDBDatabase(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
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

func resourceScalewayServerlessSQLDBDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := serverlessSQLdbAPIWithRegionAndID(m.(*meta.Meta), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := waitForServerlessSQLDBDatabase(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	req := &serverless_sqldb.UpdateDatabaseRequest{
		Region:     region,
		DatabaseID: database.ID,
	}

	if d.HasChange("max_cpu") {
		req.CPUMax = expandUint32Ptr(d.Get("max_cpu"))
	}
	if d.HasChange("min_cpu") {
		req.CPUMin = expandUint32Ptr(d.Get("min_cpu"))
	}

	if _, err := api.UpdateDatabase(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayServerlessSQLDBDatabaseRead(ctx, d, m)
}

func resourceScalewayServerlessSQLDBDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := serverlessSQLdbAPIWithRegionAndID(m.(*meta.Meta), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForServerlessSQLDBDatabase(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteDatabase(&serverless_sqldb.DeleteDatabaseRequest{
		Region:     region,
		DatabaseID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForServerlessSQLDBDatabase(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is403Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
