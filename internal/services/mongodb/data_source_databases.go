package mongodb

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/databases_datasource.md
var databasesDataSourceDescription string

func DataSourceDatabases() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMongoDBDatabasesRead,
		SchemaFunc:  dataSourceMongoDBDatabasesSchema,
		Description: databasesDataSourceDescription,
	}
}

func dataSourceMongoDBDatabasesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"instance_id": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "MongoDB instance ID. Can be a plain UUID or a regional ID.",
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		},
		"region": regional.Schema(),
		"databases": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of databases on the MongoDB instance.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the database.",
					},
				},
			},
		},
	}
}

func dataSourceMongoDBDatabasesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	mongodbAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID := locality.ExpandID(d.Get("instance_id").(string))
	if parsedRegion, id, parseErr := regional.ParseID(d.Get("instance_id").(string)); parseErr == nil {
		region = parsedRegion
		instanceID = id
	}

	res, err := mongodbAPI.ListDatabases(&mongodb.ListDatabasesRequest{
		Region:     region,
		InstanceID: instanceID,
		OrderBy:    mongodb.ListDatabasesRequestOrderByNameAsc,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, instanceID))
	_ = d.Set("instance_id", regional.NewIDString(region, instanceID))
	_ = d.Set("region", region.String())
	_ = d.Set("databases", flattenMongoDBDatabases(res.Databases))

	return nil
}

func flattenMongoDBDatabases(databases []*mongodb.Database) []any {
	result := make([]any, 0, len(databases))
	for _, database := range databases {
		result = append(result, map[string]any{
			"name": database.Name,
		})
	}

	return result
}
