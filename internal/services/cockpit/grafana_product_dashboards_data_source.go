package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCockpitGrafanaProductDashboards() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCockpitGrafanaProductDashboardsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "The ID of the project to list Grafana product dashboards for",
				ValidateDiagFunc: verify.IsUUID(),
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Filter dashboards by tags (e.g. rdb, lb)",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"dashboards": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Grafana product dashboards",
				Elem: &schema.Resource{
					Schema: grafanaProductDashboardNestedSchema(),
				},
			},
			"names": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of dashboard names",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func grafanaProductDashboardNestedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Dashboard name (use with scaleway_cockpit_grafana_product_dashboard)",
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
	}
}

func dataSourceCockpitGrafanaProductDashboardsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID, _, err := meta.ExtractProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &cockpit.GlobalAPIListGrafanaProductDashboardsRequest{
		ProjectID: projectID,
	}

	if tags, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(tags)
	}

	resp, err := retryOn403Value(ctx, func() (*cockpit.ListGrafanaProductDashboardsResponse, error) {
		return api.ListGrafanaProductDashboards(req, scw.WithContext(ctx), scw.WithAllPages())
	})
	if err != nil {
		return diag.FromErr(err)
	}

	dashboards := make([]map[string]any, 0, len(resp.Dashboards))
	names := make([]string, 0, len(resp.Dashboards))

	for _, dashboard := range resp.Dashboards {
		dashboards = append(dashboards, flattenGrafanaProductDashboard(dashboard))
		names = append(names, dashboard.Name)
	}

	d.SetId(projectID)
	_ = d.Set("project_id", projectID)
	_ = d.Set("dashboards", dashboards)
	_ = d.Set("names", types.FlattenSliceString(names))

	return nil
}

func flattenGrafanaProductDashboard(dashboard *cockpit.GrafanaProductDashboard) map[string]any {
	if dashboard == nil {
		return nil
	}

	return map[string]any{
		"name":      dashboard.Name,
		"title":     dashboard.Title,
		"url":       dashboard.URL,
		"tags":      types.FlattenSliceString(dashboard.Tags),
		"variables": types.FlattenSliceString(dashboard.Variables),
	}
}
