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
	addOptionalFieldsToSchema(dsSchema, "name", "zone")

	return &schema.Resource{
		ReadContext: dataSourceScalewayK8SPoolRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayK8SPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := m.(*Meta)
	api, region, err := k8sAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	poolID, ok := d.GetOk("pool_id")
	if !ok {
		res, err := api.ListPools(&k8s.ListPoolsRequest{
			Region: region,
			Name:   scw.StringPtr(d.Get("name").(string)),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		if len(res.Pools) == 0 {
			return diag.FromErr(fmt.Errorf("no pools found with the name %s", d.Get("name")))
		}
		if len(res.Pools) > 1 {
			return diag.FromErr(fmt.Errorf("%d pools found with the same name %s", len(res.Pools), d.Get("name")))
		}
		poolID = res.Pools[0].ID
	}

	regionID := datasourceNewRegionalizedID(poolID, region)
	d.SetId(regionID)
	err = d.Set("pool_id", regionID)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayK8SClusterRead(ctx, d, m)
}
