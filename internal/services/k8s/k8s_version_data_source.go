package k8s

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func DataSourceVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceK8SVersionRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Kubernetes version in the form x.y.z",
			},
			"major_minor_only": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the version in the form x.y (ignoring patch version)",
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
			"region": regional.Schema(),
		},
	}
}

func DataSourceK8SVersionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	k8sAPI, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	name, ok := d.GetOk("name")
	if !ok {
		return diag.FromErr(fmt.Errorf("could not find version %q", name))
	}

	var version *k8s.Version

	if name == "latest" {
		res, err := k8sAPI.ListVersions(&k8s.ListVersionsRequest{
			Region: region,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		if len(res.Versions) == 0 {
			return diag.FromErr(errors.New("could not find the latest version"))
		}

		version = res.Versions[0]
	} else {
		res, err := k8sAPI.GetVersion(&k8s.GetVersionRequest{
			Region:      region,
			VersionName: name.(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		version = res
	}

	majorMinor, err := VersionNameWithoutPatch(version.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", region, version.Name))
	_ = d.Set("name", version.Name)
	_ = d.Set("major_minor_only", majorMinor)
	_ = d.Set("available_cnis", version.AvailableCnis)
	_ = d.Set("available_container_runtimes", version.AvailableContainerRuntimes)
	_ = d.Set("available_feature_gates", version.AvailableFeatureGates)
	_ = d.Set("region", region)

	return nil
}

func VersionNameWithoutPatch(version string) (string, error) {
	versionSplit := strings.Split(version, ".")
	if len(versionSplit) != 3 {
		return "", fmt.Errorf("version name must contain 3 parts, got %q", version)
	}

	return strings.Join(versionSplit[:2], "."), nil
}
