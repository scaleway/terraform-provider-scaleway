package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCockpitGrafanaProductDashboard() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCockpitGrafanaProductDashboardRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "The ID of the project the dashboard belongs to",
				ValidateDiagFunc: verify.IsUUID(),
			},
			"dashboard_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Grafana product dashboard to retrieve",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Dashboard name",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Human-readable dashboard title",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL to open the dashboard in Grafana",
			},
			"tags": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Dashboard tags",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"variables": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Dashboard variables",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceCockpitGrafanaProductDashboardRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	dashboardName := d.Get("dashboard_name").(string)

	dashboard, err := retryOn403Value(ctx, func() (*cockpit.GrafanaProductDashboard, error) {
		return api.GetGrafanaProductDashboard(&cockpit.GlobalAPIGetGrafanaProductDashboardRequest{
			ProjectID:     projectID,
			DashboardName: dashboardName,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			return diag.Errorf("Grafana product dashboard %q not found for project %s", dashboardName, projectID)
		}

		return diag.FromErr(err)
	}

	d.SetId(projectID + "/" + dashboard.Name)
	_ = d.Set("project_id", projectID)
	_ = d.Set("dashboard_name", dashboard.Name)
	_ = d.Set("name", dashboard.Name)
	_ = d.Set("title", dashboard.Title)
	_ = d.Set("url", dashboard.URL)
	_ = d.Set("tags", types.FlattenSliceString(dashboard.Tags))
	_ = d.Set("variables", types.FlattenSliceString(dashboard.Variables))

	return nil
}
