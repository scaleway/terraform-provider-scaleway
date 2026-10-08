package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func DataSourceCockpitRulesCount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCockpitRulesCountRead,
		Schema: map[string]*schema.Schema{
			"project_id": account.ProjectIDSchema(),
			"region":     regional.Schema(),
			"preconfigured_rules_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total count of preconfigured rules.",
			},
			"custom_rules_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total count of custom rules.",
			},
			"rules_count_by_datasource": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Total count of rules grouped by data source.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_source_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the data source.",
						},
						"data_source_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the data source.",
						},
						"rules_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Total count of rules associated with this data source.",
						},
					},
				},
			},
		},
	}
}

func dataSourceCockpitRulesCountRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := retryOn403Value(ctx, func() (*cockpit.GetRulesCountResponse, error) {
		return api.GetRulesCount(&cockpit.RegionalAPIGetRulesCountRequest{
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	rulesByDS := make([]map[string]any, 0, len(resp.RulesCountByDatasource))
	for _, rc := range resp.RulesCountByDatasource {
		if rc == nil {
			continue
		}

		rulesByDS = append(rulesByDS, map[string]any{
			"data_source_id":   rc.DataSourceID,
			"data_source_name": rc.DataSourceName,
			"rules_count":      int(rc.RulesCount),
		})
	}

	d.SetId(regional.NewIDString(region, projectID))
	_ = d.Set("project_id", projectID)
	_ = d.Set("region", region.String())
	_ = d.Set("preconfigured_rules_count", int(resp.PreconfiguredRulesCount))
	_ = d.Set("custom_rules_count", int(resp.CustomRulesCount))
	_ = d.Set("rules_count_by_datasource", rulesByDS)

	return nil
}
