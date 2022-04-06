package scaleway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const createDatabaseTimeout = 2 * time.Minute

func resourceScalewayRdbDatabase() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayRdbDatabaseCreate,
		ReadContext:   resourceScalewayRdbDatabaseRead,
		DeleteContext: resourceScalewayRdbDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultRdbInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDWithLocality(),
				Description:  "Instance on which the database is created",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Database name",
				Required:    true,
				ForceNew:    true,
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
		},
	}
}

func resourceScalewayRdbDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI := newRdbAPI(meta)
	region, instanceID, err := parseRegionalID(d.Get("instance_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
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
	err = resource.RetryContext(context.Background(), createDatabaseTimeout, func() *resource.RetryError {
		currentDB, errCreateDB := rdbAPI.CreateDatabase(createReq, scw.WithContext(ctx))
		if errCreateDB != nil {
			// WIP: Issue on creation/write database. Need a database stable status
			if is409Error(errCreateDB) {
				return resource.RetryableError(errCreateDB)
			}
			return resource.NonRetryableError(errCreateDB)
		}
		// set database information
		db = currentDB
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceScalewayRdbDatabaseID(region, instanceID, db.Name))

	return resourceScalewayRdbDatabaseRead(ctx, d, meta)
}

func getDatabase(ctx context.Context, api *rdb.API, r scw.Region, instanceID, dbName string) (*rdb.Database, error) {
	res, err := api.ListDatabases(&rdb.ListDatabasesRequest{
		Region:     r,
		InstanceID: instanceID,
		Name:       &dbName,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			return nil, nil
		}
		return nil, err
	}

	return res.Databases[0], nil
}

func resourceScalewayRdbDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI := newRdbAPI(meta)
	region, instanceID, databaseName, err := resourceScalewayRdbDatabaseParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutRead))
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := getDatabase(ctx, rdbAPI, region, instanceID, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceScalewayRdbDatabaseID(region, instanceID, database.Name))
	_ = d.Set("instance_id", newRegionalID(region, instanceID).String())
	_ = d.Set("name", database.Name)
	_ = d.Set("owner", database.Owner)
	_ = d.Set("managed", database.Managed)
	_ = d.Set("size", database.Size.String())

	return nil
}

func resourceScalewayRdbDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI := newRdbAPI(meta)
	region, instanceID, databaseName, err := resourceScalewayRdbDatabaseParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	err = rdbAPI.DeleteDatabase(&rdb.DeleteDatabaseRequest{
		Region:     region,
		InstanceID: instanceID,
		Name:       databaseName,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

// Build the resource identifier
// The resource identifier format is "Region/InstanceId/DatabaseName"
func resourceScalewayRdbDatabaseID(region scw.Region, instanceID string, databaseName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, instanceID, databaseName)
}

// Extract instance ID and database from the resource identifier.
// The resource identifier format is "Region/InstanceId/DatabaseId"
func resourceScalewayRdbDatabaseParseID(resourceID string) (region scw.Region, instanceID string, database string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}
	return scw.Region(idParts[0]), idParts[1], idParts[2], nil
}
