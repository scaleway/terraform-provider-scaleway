package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func dataSourceScalewayK8SCluster() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayK8SCluster().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")
	delete(dsSchema, "delete_additional_resources")

	dsSchema["name"].ConflictsWith = []string{"cluster_id"}
	dsSchema["cluster_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the cluster",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayK8SClusterRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewayK8SClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	k8sAPI, region, err := k8sAPIWithRegion(d, m.(*meta.Meta))
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID, ok := d.GetOk("cluster_id")
	if !ok {
		clusterName := d.Get("name").(string)
		res, err := k8sAPI.ListClusters(&k8s.ListClustersRequest{
			Region:    region,
			Name:      expandStringPtr(clusterName),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundCluster, err := findExact(
			res.Clusters,
			func(s *k8s.Cluster) bool { return s.Name == clusterName },
			clusterName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		clusterID = foundCluster.ID
	}

	regionalizedID := datasourceNewRegionalID(clusterID, region)
	d.SetId(regionalizedID)
	_ = d.Set("cluster_id", regionalizedID)
	return resourceScalewayK8SClusterRead(ctx, d, m)
}
