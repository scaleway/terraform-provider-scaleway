package sdb

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdbSDK "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/versions_data_source.md
var versionsDataSourceDescription string

func DataSourceVersions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVersionsRead,
		SchemaFunc:  dataSourceVersionsSchema,
		Description: versionsDataSourceDescription,
	}
}

func dataSourceVersionsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"region": regional.Schema(),
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter Serverless SQL Database versions that match a given name pattern (e.g. `16`).",
		},
		"versions": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available Serverless SQL Database versions.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The major version of the PostgreSQL engine.",
					},
					"end_of_life_at": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "End of life date of the version (RFC3339).",
					},
					"srn": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The Scaleway Resource Name (SRN) of the version.",
					},
				},
			},
		},
	}
}

func dataSourceVersionsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	versions, err := api.ListVersions(&sdbSDK.ListVersionsRequest{
		Region:  region,
		Version: types.ExpandStringPtr(d.Get("name")),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(region.String())
	_ = d.Set("region", region.String())
	_ = d.Set("versions", flattenSDBVersions(versions.Versions))

	return nil
}

func flattenSDBVersions(versions []*sdbSDK.Version) []any {
	result := make([]any, 0, len(versions))

	for _, version := range versions {
		result = append(result, map[string]any{
			"name":           version.Name,
			"end_of_life_at": types.FlattenTime(version.EndOfLifeAt),
			"srn":            version.Srn,
		})
	}

	return result
}
