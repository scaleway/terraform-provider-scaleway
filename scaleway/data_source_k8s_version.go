package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayK8SVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayK8SVersionRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Kubernetes version",
			},
			"available_cnis": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The list of supported Container Network Interface (CNI) plugins for this version",
			},
			"available_container_runtimes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The list of supported container runtimes for this version",
			},
			"available_feature_gates": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The list of supported feature gates for this version",
			},
			"region": regionSchema(),
		},
	}
}

func dataSourceScalewayK8SVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	k8sAPI, region, err := k8sAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	name, ok := d.GetOk("name")
	if !ok {
		return diag.FromErr(fmt.Errorf("could not find version %q", name))
	}
	res, err := k8sAPI.GetVersion(&k8s.GetVersionRequest{
		Region:      region,
		VersionName: name.(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", region, res.Name))
	_ = d.Set("name", res.Name)
	_ = d.Set("available_cnis", res.AvailableCnis)
	_ = d.Set("available_container_runtimes", res.AvailableContainerRuntimes)
	_ = d.Set("available_feature_gates", res.AvailableFeatureGates)
	_ = d.Set("region", region)

	return nil
}
