package redis

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

//go:embed descriptions/node_types_data_source.md
var nodeTypesDataSourceDescription string

func DataSourceNodeTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNodeTypesRead,
		SchemaFunc:  dataSourceNodeTypesSchema,
		Description: nodeTypesDataSourceDescription,
	}
}

func dataSourceNodeTypesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"zone": zonal.Schema(),
		"include_disabled_types": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to include disabled node types in the list.",
		},
		"node_types": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available Redis™ node types.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the node type.",
					},
					"stock_status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Current stock status of the node type.",
					},
					"description": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Current specifications of the offer.",
					},
					"vcpus": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Number of virtual CPUs.",
					},
					"memory_size_in_gb": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Amount of memory available in GB.",
					},
					"disabled": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether the node type is currently disabled.",
					},
					"beta": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether the node type is currently in beta.",
					},
				},
			},
		},
	}
}

func dataSourceNodeTypesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.ListNodeTypes(&redis.ListNodeTypesRequest{
		Zone:                 zone,
		IncludeDisabledTypes: d.Get("include_disabled_types").(bool),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zone.String())
	_ = d.Set("zone", zone.String())
	_ = d.Set("node_types", flattenRedisNodeTypes(res.NodeTypes))

	return nil
}

func flattenRedisNodeTypes(nodeTypes []*redis.NodeType) []any {
	result := make([]any, 0, len(nodeTypes))

	for _, nodeType := range nodeTypes {
		result = append(result, map[string]any{
			"name":              nodeType.Name,
			"stock_status":      nodeType.StockStatus.String(),
			"description":       nodeType.Description,
			"vcpus":             int(nodeType.Vcpus),
			"memory_size_in_gb": int(nodeType.Memory / scw.GB),
			"disabled":          nodeType.Disabled,
			"beta":              nodeType.Beta,
		})
	}

	return result
}
