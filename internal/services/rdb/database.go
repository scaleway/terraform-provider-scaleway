package rdb

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceRdbDatabaseCreate,
		ReadContext:   ResourceRdbDatabaseRead,
		DeleteContext: ResourceRdbDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultInstanceTimeout),
			Delete:  schema.DefaultTimeout(defaultInstanceTimeout),
			Default: schema.DefaultTimeout(defaultInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: verify.IsUUIDWithLocality(),
				Description:      "Instance on which the database is created",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Database name",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringNotInSlice([]string{
						"information_schema",
						"mysql",
						"performance_schema",
						"postgres",
						"rdb",
						"rdb",
						"sys",
						"template0",
						"template1",
					}, false),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z\d_$-]*$`), "database name must contain only alphanumeric characters, underscores and dashes and it must start with a letter"),
				),
			},
			"managed": {
				Type:        schema.TypeBool,
				Description: "Whether or not the database is managed",
				Computed:    true,
			},
			"owner": {
				Type:        schema.TypeString,
				Description: "User that own the database",
				Computed:    true,
			},
			"size": {
				Type:        schema.TypeString,
				Description: "Size of the database",
				Computed:    true,
			},
			"region": regional.Schema(),
		},
		CustomizeDiff: cdf.LocalityCheck("instance_id"),
	}
}

func ResourceRdbDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI := newAPI(m)
	region, instanceID, err := regional.ParseID(d.Get("instance_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &rdb.CreateDatabaseRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       d.Get("name").(string),
	}

	var db *rdb.Database
	//  wrapper around StateChangeConf that will just retry the database creation
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		currentDB, errCreateDB := rdbAPI.CreateDatabase(createReq, scw.WithContext(ctx))
		if errCreateDB != nil {
			// WIP: Issue on creation/write database. Need a database stable status
			if httperrors.Is409(errCreateDB) {
				return retry.RetryableError(errCreateDB)
			}

			return retry.NonRetryableError(errCreateDB)
		}
		// set database information
		db = currentDB

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ResourceRdbDatabaseID(region, instanceID, db.Name))
	_ = d.Set("region", region)

	return ResourceRdbDatabaseRead(ctx, d, m)
}

func getDatabase(ctx context.Context, api *rdb.API, r scw.Region, instanceID, dbName string) (*rdb.Database, error) {
	res, err := api.ListDatabases(&rdb.ListDatabasesRequest{
		Region:     r,
		InstanceID: instanceID,
		Name:       &dbName,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if len(res.Databases) == 0 {
		return nil, fmt.Errorf("database %s not found", dbName)
	}

	return res.Databases[0], nil
}

func ResourceRdbDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI := newAPI(m)
	region, instanceID, databaseName, err := ResourceRdbDatabaseParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := getDatabase(ctx, rdbAPI, region, instanceID, databaseName)
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	d.SetId(ResourceRdbDatabaseID(region, instanceID, database.Name))
	_ = d.Set("instance_id", regional.NewID(region, instanceID).String())
	_ = d.Set("name", database.Name)
	_ = d.Set("owner", database.Owner)
	_ = d.Set("managed", database.Managed)
	_ = d.Set("size", database.Size.String())
	_ = d.Set("region", region)

	return nil
}

func ResourceRdbDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI := newAPI(m)
	region, instanceID, databaseName, err := ResourceRdbDatabaseParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = rdbAPI.DeleteDatabase(&rdb.DeleteDatabaseRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       databaseName,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

// ResourceRdbDatabaseID builds the resource identifier
// The resource identifier format is "Region/InstanceId/DatabaseName"
func ResourceRdbDatabaseID(region scw.Region, instanceID string, databaseName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, instanceID, databaseName)
}

// ResourceRdbDatabaseParseID extracts instance ID and database from the resource identifier.
// The resource identifier format is "Region/InstanceId/DatabaseId"
func ResourceRdbDatabaseParseID(resourceID string) (region scw.Region, instanceID string, database string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}

	return scw.Region(idParts[0]), idParts[1], idParts[2], nil
}
