package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCockpit() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceCockpit().Schema)

	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The project_id you want to attach the resource to",
		Optional:     true,
		ValidateFunc: verify.IsUUID(),
	}
	dsSchema["plan"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The current plan of the cockpit project",
	}

	return &schema.Resource{
		ReadContext:        dataSourceCockpitRead,
		Schema:             dsSchema,
		DeprecationMessage: "The 'scaleway_cockpit' data source is deprecated because it duplicates the functionality of the 'scaleway_cockpit' resource. Please use the 'scaleway_cockpit' resource instead.",
	}
}

func dataSourceCockpitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}
	regionalAPI, region, err := cockpitAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)

	res, err := api.GetCurrentPlan(&cockpit.GlobalAPIGetCurrentPlanRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	_ = d.Set("project_id", d.Get("project_id").(string))
	_ = d.Set("plan", res.Name)
	_ = d.Set("plan_id", res.Name)

	dataSourcesRes, err := regionalAPI.ListDataSources(&cockpit.RegionalAPIListDataSourcesRequest{
		Region:    region,
		ProjectID: d.Get("project_id").(string),
		Origin:    "external",
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return diag.FromErr(err)
	}

	grafana, err := api.GetGrafana(&cockpit.GlobalAPIGetGrafanaRequest{
		ProjectID: d.Get("project_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	alertManager, err := regionalAPI.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
		ProjectID: d.Get("project_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	alertManagerURL := ""
	if alertManager.AlertManagerURL != nil {
		alertManagerURL = *alertManager.AlertManagerURL
	}

	endpoints := flattenCockpitEndpoints(dataSourcesRes.DataSources, grafana.GrafanaURL, alertManagerURL)

	_ = d.Set("endpoints", endpoints)
	_ = d.Set("push_url", createCockpitPushURLList(endpoints))
	d.SetId(projectID)
	return nil
}
