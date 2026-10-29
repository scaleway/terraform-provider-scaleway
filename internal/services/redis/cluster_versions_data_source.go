package redis

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/cluster_versions_data_source.md
var clusterVersionsDataSourceDescription string

func DataSourceClusterVersions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterVersionsRead,
		SchemaFunc:  dataSourceClusterVersionsSchema,
		Description: clusterVersionsDataSourceDescription,
	}
}

func dataSourceClusterVersionsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"zone": zonal.Schema(),
		"include_disabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to include disabled Redis™ engine versions.",
		},
		"include_beta": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to include beta Redis™ engine versions.",
		},
		"include_deprecated": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to include deprecated Redis™ engine versions.",
		},
		"version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter Redis™ engine versions that match a given name pattern.",
		},
		"versions": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available Redis™ cluster versions.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"version": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Redis™ engine version.",
					},
					"end_of_life_at": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "End of life date of the version (RFC3339).",
					},
					"released_at": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Release date of the version (RFC3339).",
					},
					"logo_url": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "URL of the Redis™ logo.",
					},
				},
			},
		},
	}
}

func dataSourceClusterVersionsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.ListClusterVersions(&redis.ListClusterVersionsRequest{
		Zone:              zone,
		IncludeDisabled:   d.Get("include_disabled").(bool),
		IncludeBeta:       d.Get("include_beta").(bool),
		IncludeDeprecated: d.Get("include_deprecated").(bool),
		Version:           types.ExpandStringPtr(d.Get("version")),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zone.String())
	_ = d.Set("zone", zone.String())
	_ = d.Set("versions", flattenRedisClusterVersions(res.Versions))

	return nil
}

func flattenRedisClusterVersions(versions []*redis.ClusterVersion) []any {
	result := make([]any, 0, len(versions))

	for _, version := range versions {
		result = append(result, map[string]any{
			"version":        version.Version,
			"end_of_life_at": types.FlattenTime(version.EndOfLifeAt),
			"released_at":    types.FlattenTime(version.ReleasedAt),
			"logo_url":       version.LogoURL,
		})
	}

	return result
}
