package rdb

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/database_engines_data_source.md
var databaseEnginesDataSourceDescription string

func DataSourceDatabaseEngines() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDatabaseEnginesRead,
		SchemaFunc:  dataSourceDatabaseEnginesSchema,
		Description: databaseEnginesDataSourceDescription,
	}
}

func dataSourceDatabaseEnginesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region": regional.Schema(),
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter by database engine name (e.g. PostgreSQL, MySQL).",
		},
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter by database engine version (e.g. 16).",
		},
		"engines": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available database engines.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the database engine.",
					},
					"logo_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "URL of the database engine logo.",
					},
					"region": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Region of the database engine.",
					},
					"versions": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Available versions for the database engine.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"version": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Database engine version number.",
								},
								"name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Full engine name including version (e.g. PostgreSQL-16).",
								},
								"end_of_life": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "End of life date of the engine version (RFC3339).",
								},
								"disabled": {
									Type:        schema.TypeBool,
									Computed:    true,
									Description: "Whether the engine version is disabled and cannot be created.",
								},
								"beta": {
									Type:        schema.TypeBool,
									Computed:    true,
									Description: "Whether the engine version is in beta.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceDatabaseEnginesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.ListDatabaseEngines(&rdb.ListDatabaseEnginesRequest{
		Region:  region,
		Name:    types.ExpandStringPtr(d.Get("name")),
		Version: types.ExpandStringPtr(d.Get("version")),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(region.String())
	_ = d.Set("region", region.String())
	_ = d.Set("engines", flattenRDBDatabaseEngines(res.Engines))

	return nil
}

func flattenRDBDatabaseEngines(engines []*rdb.DatabaseEngine) []any {
	result := make([]any, 0, len(engines))

	for _, engine := range engines {
		result = append(result, map[string]any{
			"name":     engine.Name,
			"logo_url": engine.LogoURL,
			"region":   engine.Region.String(),
			"versions": flattenRDBEngineVersions(engine.Versions),
		})
	}

	return result
}

func flattenRDBEngineVersions(versions []*rdb.EngineVersion) []any {
	result := make([]any, 0, len(versions))

	for _, version := range versions {
		result = append(result, map[string]any{
			"version":     version.Version,
			"name":        version.Name,
			"end_of_life": types.FlattenTime(version.EndOfLife),
			"disabled":    version.Disabled,
			"beta":        version.Beta,
		})
	}

	return result
}
