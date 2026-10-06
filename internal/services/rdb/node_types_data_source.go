package rdb

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
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
		"region": regional.Schema(),
		"include_disabled_types": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to include disabled node types in the list.",
		},
		"node_types": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of available RDB node types.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name identifier of the node type.",
					},
					"stock_status": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Current stock status for the node type.",
					},
					"description": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Description of the node type offer.",
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
					"is_ha_required": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether the node type can only be used with high availability.",
					},
					"generation": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Generation associated with the node type offer.",
					},
					"instance_range": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Instance range associated with the node type offer.",
					},
					"available_volume_types": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Available storage options for the node type.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"type": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Volume type.",
								},
								"description": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Description of the volume type.",
								},
								"min_size_in_gb": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Minimum volume size in GB.",
								},
								"max_size_in_gb": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Maximum volume size in GB.",
								},
								"chunk_size_in_gb": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "Minimum increment level for a Block Storage volume size in GB.",
								},
								"class": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Storage class of the volume.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceNodeTypesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := api.ListNodeTypes(&rdb.ListNodeTypesRequest{
		Region:               region,
		IncludeDisabledTypes: d.Get("include_disabled_types").(bool),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(region.String())
	_ = d.Set("region", region.String())
	_ = d.Set("node_types", flattenRDBNodeTypes(res.NodeTypes))

	return nil
}

func flattenRDBNodeTypes(nodeTypes []*rdb.NodeType) []any {
	result := make([]any, 0, len(nodeTypes))

	for _, nodeType := range nodeTypes {
		result = append(result, map[string]any{
			"name":                   nodeType.Name,
			"stock_status":           nodeType.StockStatus.String(),
			"description":            nodeType.Description,
			"vcpus":                  int(nodeType.Vcpus),
			"memory_size_in_gb":      int(nodeType.Memory / scw.GB),
			"disabled":               nodeType.Disabled,
			"beta":                   nodeType.Beta,
			"is_ha_required":         nodeType.IsHaRequired,
			"generation":             nodeType.Generation.String(),
			"instance_range":         nodeType.InstanceRange,
			"available_volume_types": flattenRDBNodeTypeVolumeTypes(nodeType.AvailableVolumeTypes),
		})
	}

	return result
}

func flattenRDBNodeTypeVolumeTypes(volumeTypes []*rdb.NodeTypeVolumeType) []any {
	result := make([]any, 0, len(volumeTypes))

	for _, volumeType := range volumeTypes {
		result = append(result, map[string]any{
			"type":             volumeType.Type.String(),
			"description":      volumeType.Description,
			"min_size_in_gb":   int(volumeType.MinSize / scw.GB),
			"max_size_in_gb":   int(volumeType.MaxSize / scw.GB),
			"chunk_size_in_gb": int(volumeType.ChunkSize / scw.GB),
			"class":            volumeType.Class.String(),
		})
	}

	return result
}
