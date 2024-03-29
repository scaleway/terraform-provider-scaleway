package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/redis/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceScalewayRedisCluster() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceScalewayRedisCluster().Schema)
	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["name"].ConflictsWith = []string{"cluster_id"}
	dsSchema["cluster_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the Redis cluster",
		ConflictsWith: []string{"name"},
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayRedisClusterRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayRedisClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, zone, err := redisAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID, ok := d.GetOk("cluster_id")
	if !ok {
		clusterName := d.Get("name").(string)
		res, err := api.ListClusters(&redis.ListClustersRequest{
			Zone:      zone,
			Name:      types.ExpandStringPtr(clusterName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundCluster, err := findExact(
			res.Clusters,
			func(s *redis.Cluster) bool { return s.Name == clusterName },
			clusterName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		clusterID = foundCluster.ID
	}

	zonedID := datasource.NewZonedID(clusterID, zone)
	d.SetId(zonedID)
	err = d.Set("cluster_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if cluster exist as Read will return nil if resource does not exist
	// clusterID may be zoned if using name in data source
	getReq := &redis.GetClusterRequest{
		Zone:      zone,
		ClusterID: locality.ExpandID(clusterID.(string)),
	}
	_, err = api.GetCluster(getReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("no clusters found with the id %s", clusterID))
	}

	return resourceScalewayRedisClusterRead(ctx, d, m)
}
