package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func DataSourceCockpitConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCockpitConfigRead,
		Schema: map[string]*schema.Schema{
			"region": regional.Schema(),
			"custom_metrics_retention":  retentionSchema("Retention limits and default for custom metrics data sources."),
			"custom_logs_retention":     retentionSchema("Retention limits and default for custom logs data sources."),
			"custom_traces_retention":   retentionSchema("Retention limits and default for custom traces data sources."),
			"product_metrics_retention": retentionSchema("Retention limits and default for Scaleway product metrics data sources."),
			"product_logs_retention":    retentionSchema("Retention limits and default for Scaleway product logs data sources."),
		},
	}
}

func retentionSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: description,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"min_days": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Minimum retention in days.",
				},
				"max_days": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Maximum retention in days.",
				},
				"default_days": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Default retention in days.",
				},
			},
		},
	}
}

func dataSourceCockpitConfigRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := api.GetConfig(&cockpit.RegionalAPIGetConfigRequest{
		Region: region,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(region.String())
	_ = d.Set("region", region.String())
	_ = d.Set("custom_metrics_retention", flattenRetention(resp.CustomMetricsRetention))
	_ = d.Set("custom_logs_retention", flattenRetention(resp.CustomLogsRetention))
	_ = d.Set("custom_traces_retention", flattenRetention(resp.CustomTracesRetention))
	_ = d.Set("product_metrics_retention", flattenRetention(resp.ProductMetricsRetention))
	_ = d.Set("product_logs_retention", flattenRetention(resp.ProductLogsRetention))

	return nil
}

func flattenRetention(retention *cockpit.GetConfigResponseRetention) []map[string]any {
	if retention == nil {
		return nil
	}

	return []map[string]any{
		{
			"min_days":     int(retention.MinDays),
			"max_days":     int(retention.MaxDays),
			"default_days": int(retention.DefaultDays),
		},
	}
}
