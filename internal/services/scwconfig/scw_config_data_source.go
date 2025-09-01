package scwconfig

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func DataSourceConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Description: "Project ID used",
				Computed:    true,
			},
			"project_id_source": {
				Type:        schema.TypeString,
				Description: "Where the project id definition comes from (Environment, configuration file, variable, ...)",
				Computed:    true,
			},
			"access_key": {
				Type:        schema.TypeString,
				Description: "Access Key used",
				Computed:    true,
			},
			"access_key_source": {
				Type:        schema.TypeString,
				Description: "Where the access key definition comes from (Environment, configuration file, variable, ...)",
				Computed:    true,
			},
			"secret_key": {
				Type:        schema.TypeString,
				Description: "Secret Key used",
				Computed:    true,
				Sensitive:   true,
			},
			"secret_key_source": {
				Type:        schema.TypeString,
				Description: "Where the secret key definition comes from (Environment, configuration file, variable, ...)",
				Computed:    true,
			},
			"zone": {
				Type:        schema.TypeString,
				Description: "Zone used",
				Computed:    true,
			},
			"zone_source": {
				Type:        schema.TypeString,
				Description: "Where the zone definition comes from (Environment, configuration file, variable, ...)",
				Computed:    true,
			},
			"region": {
				Type:        schema.TypeString,
				Description: "Region used",
				Computed:    true,
			},
			"region_source": {
				Type:        schema.TypeString,
				Description: "Where the region definition comes from (Environment, configuration file, variable, ...)",
				Computed:    true,
			},
		},
	}
}

func dataSourceConfigRead(_ context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := meta.ExtractScwClient(m)
	providerMeta := m.(*meta.Meta)

	d.SetId("0")

	accessKey, _ := client.GetAccessKey()
	_ = d.Set("access_key", accessKey)
	_ = d.Set("access_key_source", providerMeta.AccessKeySource())

	secretKey, _ := client.GetSecretKey()
	_ = d.Set("secret_key", secretKey)
	_ = d.Set("secret_key_source", providerMeta.SecretKeySource())

	projectID, _ := client.GetDefaultProjectID()
	_ = d.Set("project_id", projectID)
	_ = d.Set("project_id_source", providerMeta.ProjectIDSource())

	zone, _ := client.GetDefaultZone()
	_ = d.Set("zone", zone)
	_ = d.Set("zone_source", providerMeta.ZoneSource())

	region, _ := client.GetDefaultRegion()
	_ = d.Set("region", region)
	_ = d.Set("region_source", providerMeta.RegionSource())

	return nil
}
