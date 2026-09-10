package messageq

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func DataSourceNodeType() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceNodeTypeRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The node type name",
			},
			"region": regional.Schema(),
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the node type",
			},
			"vcpus": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of vCPUs available",
			},
			"memory_size_in_gb": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Amount of memory available in GB",
			},
			"stock_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Stock status of the node type",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the node type is disabled",
			},
			"beta": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the node type is in beta",
			},
			"instance_range": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Instance range associated with the node type offer",
			},
			"available_volume_types": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Available storage options for the node type",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Volume type",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Volume type description",
						},
						"min_size_in_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Minimum volume size in GB",
						},
						"max_size_in_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum volume size in GB",
						},
						"chunk_size_in_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Volume size increment in GB",
						},
					},
				},
			},
		},
	}
}

func DataSourceNodeTypeRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)

	res, err := api.ListNodeTypes(&messageqapi.ListNodeTypesRequest{
		Region: region,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	var found *messageqapi.NodeType

	for _, nodeType := range res.NodeTypes {
		if nodeType.Name == name {
			found = nodeType

			break
		}
	}

	if found == nil {
		return diag.FromErr(fmt.Errorf("messageq node type %q not found", name))
	}

	d.SetId(regional.NewIDString(region, found.Name))
	_ = d.Set("region", region.String())
	_ = d.Set("name", found.Name)
	_ = d.Set("description", found.Description)
	_ = d.Set("vcpus", int(found.Vcpus))
	_ = d.Set("memory_size_in_gb", flattenVolumeSizeGB(found.MemoryBytes))
	_ = d.Set("stock_status", string(found.StockStatus))
	_ = d.Set("disabled", found.Disabled)
	_ = d.Set("beta", found.Beta)
	_ = d.Set("instance_range", found.InstanceRange)

	volumeTypes := make([]map[string]any, 0, len(found.AvailableVolumeTypes))
	for _, volumeType := range found.AvailableVolumeTypes {
		volumeTypes = append(volumeTypes, map[string]any{
			"type":             string(volumeType.Type),
			"description":      volumeType.Description,
			"min_size_in_gb":   flattenVolumeSizeGB(volumeType.MinSizeBytes),
			"max_size_in_gb":   flattenVolumeSizeGB(volumeType.MaxSizeBytes),
			"chunk_size_in_gb": flattenVolumeSizeGB(volumeType.ChunkSizeBytes),
		})
	}

	_ = d.Set("available_volume_types", volumeTypes)

	return nil
}
