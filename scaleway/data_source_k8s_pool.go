package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func dataSourceScalewayK8SPool() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayK8SPool().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "region", "cluster_id", "size")

	dsSchema["name"].ConflictsWith = []string{"pool_id"}
	dsSchema["cluster_id"].ConflictsWith = []string{"pool_id"}
	dsSchema["cluster_id"].RequiredWith = []string{"name"}
	dsSchema["pool_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the pool",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name", "cluster_id"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayK8SPoolRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewayK8SPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	k8sAPI, region, err := k8sAPIWithRegion(d, m.(*meta.Meta))
	if err != nil {
		return diag.FromErr(err)
	}

	poolID, ok := d.GetOk("pool_id")
	if !ok {
		poolName := d.Get("name").(string)
		clusterID := regional.ExpandID(d.Get("cluster_id"))
		res, err := k8sAPI.ListPools(&k8s.ListPoolsRequest{
			Region:    region,
			Name:      expandStringPtr(poolName),
			ClusterID: clusterID.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundPool, err := findExact(
			res.Pools,
			func(s *k8s.Pool) bool { return s.Name == poolName },
			poolName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		poolID = foundPool.ID
	}

	regionalizedID := datasourceNewRegionalID(poolID, region)
	d.SetId(regionalizedID)
	_ = d.Set("pool_id", regionalizedID)
	return resourceScalewayK8SPoolRead(ctx, d, m)
}
