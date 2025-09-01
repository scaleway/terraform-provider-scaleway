package k8s

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCluster() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCluster().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")
	delete(dsSchema, "delete_additional_resources")

	dsSchema["name"].ConflictsWith = []string{"cluster_id"}
	dsSchema["cluster_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the cluster",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name"},
	}

	return &schema.Resource{
		ReadContext: DataSourceK8SClusterRead,

		Schema: dsSchema,
	}
}

func DataSourceK8SClusterRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	k8sAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID, ok := d.GetOk("cluster_id")
	if !ok {
		clusterName := d.Get("name").(string)

		res, err := k8sAPI.ListClusters(&k8s.ListClustersRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(clusterName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundCluster, err := datasource.FindExact(
			res.Clusters,
			func(s *k8s.Cluster) bool { return s.Name == clusterName },
			clusterName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		clusterID = foundCluster.ID
	}

	regionalizedID := datasource.NewRegionalID(clusterID, region)
	d.SetId(regionalizedID)
	_ = d.Set("cluster_id", regionalizedID)

	return ResourceK8SClusterRead(ctx, d, m)
}
