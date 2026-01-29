package k8s

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/pool_datasource.md
var poolDataSourceDescription string

func DataSourcePool() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourcePool().SchemaFunc())

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "cluster_id", "size")

	dsSchema["name"].ConflictsWith = []string{"pool_id"}
	dsSchema["cluster_id"].ConflictsWith = []string{"pool_id"}
	dsSchema["cluster_id"].RequiredWith = []string{"name"}
	dsSchema["pool_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the pool",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name", "cluster_id"},
	}

	return &schema.Resource{
		ReadContext: DataSourceK8SPoolRead,
		Description: poolDataSourceDescription,
		Schema:      dsSchema,
	}
}

func DataSourceK8SPoolRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	k8sAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	poolID, ok := d.GetOk("pool_id")
	if !ok {
		poolName := d.Get("name").(string)
		clusterID := regional.ExpandID(d.Get("cluster_id"))

		res, err := k8sAPI.ListPools(&k8s.ListPoolsRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(poolName),
			ClusterID: clusterID.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundPool, err := datasource.FindExact(
			res.Pools,
			func(s *k8s.Pool) bool { return s.Name == poolName },
			poolName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		poolID = foundPool.ID
	}

	regionalizedID := datasource.NewRegionalID(poolID, region)
	d.SetId(regionalizedID)
	_ = d.Set("pool_id", regionalizedID)

	req := &k8s.GetPoolRequest{
		Region: region,
		PoolID: poolID.(string),
	}
	pool, err := k8sAPI.GetPool(req, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	nodes, err := getNodes(ctx, k8sAPI, pool)
	if err != nil {
		return diag.FromErr(err)
	}

	return setPoolState(ctx, d, m, pool, k8sAPI, nodes)
}
