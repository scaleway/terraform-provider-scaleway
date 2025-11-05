package cockpit

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceCockpitGrafana() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCockpitGrafanaRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				Description:      "The project ID associated with the Grafana instance",
				ValidateDiagFunc: verify.IsUUID(),
			},
			"grafana_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL to access the Grafana dashboard",
			},
		},
	}
}

func dataSourceCockpitGrafanaRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, err := NewGlobalAPI(m)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if projectID == "" {
		defaultProjectID, err := getDefaultProjectID(ctx, m)
		if err != nil {
			return diag.FromErr(err)
		}
		projectID = defaultProjectID
	}

	grafana, err := api.GetGrafana(&cockpit.GlobalAPIGetGrafanaRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return diag.Errorf("Grafana instance not found for project %s. Ensure that Cockpit is activated for this project.", projectID)
		}
		return diag.FromErr(err)
	}

	d.SetId(projectID)
	_ = d.Set("project_id", projectID)
	_ = d.Set("grafana_url", grafana.GrafanaURL)

	return nil
}
