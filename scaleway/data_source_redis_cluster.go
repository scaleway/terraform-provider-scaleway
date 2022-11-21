package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayRedisCluster() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayRedisCluster().Schema)
	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "zone")

	dsSchema["name"].ConflictsWith = []string{"cluster_id"}
	dsSchema["cluster_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Redis cluster",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayRedisClusterRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRedisClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, zone, err := redisAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID, ok := d.GetOk("cluster_id")
	if !ok {
		res, err := api.ListClusters(&redis.ListClustersRequest{
			Zone:      zone,
			Name:      expandStringPtr(d.Get("name")),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		for _, cluster := range res.Clusters {
			if cluster.Name == d.Get("name").(string) {
				if clusterID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 cluster found with the same name %s", d.Get("name")))
				}
				clusterID = cluster.ID
			}
		}
		if clusterID == "" {
			return diag.FromErr(fmt.Errorf("no clusters found with the name %s", d.Get("name")))
		}
	}

	zonedID := datasourceNewZonedID(clusterID, zone)
	d.SetId(zonedID)
	err = d.Set("cluster_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if cluster exist as Read will return nil if resource does not exist
	// clusterID may be zoned if using name in data source
	getReq := &redis.GetClusterRequest{
		Zone:      zone,
		ClusterID: expandID(clusterID.(string)),
	}
	_, err = api.GetCluster(getReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("no clusters found with the id %s", clusterID))
	}

	return resourceScalewayRedisClusterRead(ctx, d, meta)
}
