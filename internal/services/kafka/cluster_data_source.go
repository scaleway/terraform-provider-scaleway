package kafka

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kafkaapi "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCluster() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCluster().SchemaFunc())
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "project_id")

	dsSchema["name"].ConflictsWith = []string{"cluster_id"}
	dsSchema["cluster_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the Kafka cluster",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceClusterRead,
		Schema:      dsSchema,
	}
}

func DataSourceClusterRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID, ok := d.GetOk("cluster_id")
	if !ok {
		clusterName := d.Get("name").(string)

		res, err := api.ListClusters(&kafkaapi.ListClustersRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(clusterName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundCluster, err := datasource.FindExact(
			res.Clusters,
			func(s *kafkaapi.Cluster) bool { return s.Name == clusterName },
			clusterName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		clusterID = foundCluster.ID
	}

	regionalID := datasource.NewRegionalID(clusterID, region)
	d.SetId(regionalID)

	err = d.Set("cluster_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.GetCluster(&kafkaapi.GetClusterRequest{
		Region:    region,
		ClusterID: locality.ExpandID(clusterID.(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("no clusters found with the id %s", clusterID))
	}

	return readClusterIntoState(ctx, d, m)
}
