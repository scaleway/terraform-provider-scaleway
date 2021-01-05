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

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "region")
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
	meta := m.(*Meta)
	k8sAPI, region, err := k8sAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID, ok := d.GetOk("cluster_id")
	if !ok {
		res, err := k8sAPI.ListClusters(&k8s.ListClustersRequest{
			Region:    region,
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
			return diag.FromErr(fmt.Errorf("no cluster found with the name %s", d.Get("name")))
		}
	}

	regionalizedID := datasourceNewRegionalizedID(clusterID, region)
	d.SetId(regionalizedID)
	_ = d.Set("cluster_id", regionalizedID)
	return resourceScalewayK8SClusterRead(ctx, d, m)
}
