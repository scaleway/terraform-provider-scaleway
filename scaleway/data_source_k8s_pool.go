package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
	meta := m.(*Meta)
	k8sAPI, region, err := k8sAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	poolID, ok := d.GetOk("pool_id")
	if !ok {
		clusterID := expandRegionalID(d.Get("cluster_id"))
		res, err := k8sAPI.ListPools(&k8s.ListPoolsRequest{
			Region:    region,
			Name:      expandStringPtr(d.Get("name")),
			ClusterID: clusterID.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		for _, pool := range res.Pools {
			if pool.Name == d.Get("name").(string) {
				if poolID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 pool found with the same name %s", d.Get("name")))
				}
				poolID = pool.ID
			}
		}
		if poolID == "" {
			return diag.FromErr(fmt.Errorf("no pool found with the name %s", d.Get("name")))
		}
	}

	regionalizedID := datasourceNewRegionalizedID(poolID, region)
	d.SetId(regionalizedID)
	_ = d.Set("pool_id", regionalizedID)
	return resourceScalewayK8SPoolRead(ctx, d, m)
}
