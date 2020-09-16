package scaleway

import (
	"context"

	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayK8SCluster() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayK8SCluster().Schema)

	return &schema.Resource{
		ReadContext: dataSourceScalewayK8SClusterRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayK8SClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(*Meta)
	api, region, err := k8sAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID, ok := d.GetOk("cluster_id")
	if !ok {
		res, err := api.ListClusters(&k8s.ListClustersRequest{
			Region: region,
			Name:   scw.StringPtr(d.Get("name").(string)),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Clusters) == 0 {
			return diag.FromErr(fmt.Errorf("no clusters found with the name %s", d.Get("name")))
		}
		if len(res.Clusters) > 1 {
			return diag.FromErr(fmt.Errorf("%d clusters found with the same name %s", len(res.Clusters), d.Get("name")))
		}
		clusterID = res.Clusters[0].ID
	}

	regionalID := datasourceNewRegionalizedID(clusterID, region)
	d.SetId(regionalID)
	_ = d.Set("cluster_id", regionalID)

	return resourceScalewayK8SClusterRead(ctx, d, m)
}
