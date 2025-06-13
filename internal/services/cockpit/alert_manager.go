package cockpit

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCockpitAlertManager() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCockpitAlertManagerCreate,
		ReadContext:   ResourceCockpitAlertManagerRead,
		UpdateContext: ResourceCockpitAlertManagerUpdate,
		DeleteContext: ResourceCockpitAlertManagerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": account.ProjectIDSchema(),
			"enable_managed_alerts": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable or disable the alert manager",
			},

			"contact_points": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of contact points",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: verify.IsEmail(),
							Description:      "Email addresses for the alert receivers",
						},
					},
				},
			},
			"region": regional.Schema(),
			"alert_manager_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Alert manager URL",
			},
		},
	}
}

func ResourceCockpitAlertManagerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	contactPoints := d.Get("contact_points").([]any)
	EnableManagedAlerts := d.Get("enable_managed_alerts").(bool)

	_, err = api.EnableAlertManager(&cockpit.RegionalAPIEnableAlertManagerRequest{
		Region:    region,
		ProjectID: projectID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if EnableManagedAlerts {
		_, err = api.EnableManagedAlerts(&cockpit.RegionalAPIEnableManagedAlertsRequest{
			Region:    region,
			ProjectID: projectID,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if len(contactPoints) > 0 {
		for _, cp := range contactPoints {
			cpMap, ok := cp.(map[string]any)
			if !ok {
				return diag.FromErr(errors.New("invalid contact point format"))
			}

			email, ok := cpMap["email"].(string)
			if !ok {
				return diag.FromErr(errors.New("invalid email format"))
			}

			emailCP := &cockpit.ContactPointEmail{
				To: email,
			}

			_, err = api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
				ProjectID: projectID,
				Email:     emailCP,
				Region:    region,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId(ResourceCockpitAlertManagerID(region, projectID))

	return ResourceCockpitAlertManagerRead(ctx, d, meta)
}

func ResourceCockpitAlertManagerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)

	alertManager, err := api.GetAlertManager(&cockpit.RegionalAPIGetAlertManagerRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("enable_managed_alerts", alertManager.ManagedAlertsEnabled)
	_ = d.Set("region", alertManager.Region)
	_ = d.Set("alert_manager_url", alertManager.AlertManagerURL)

	contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var contactPointsList []map[string]any

	for _, cp := range contactPoints.ContactPoints {
		if cp.Email != nil {
			contactPoint := map[string]any{
				"email": cp.Email.To,
			}
			contactPointsList = append(contactPointsList, contactPoint)
		}
	}

	_ = d.Set("contact_points", contactPointsList)

	return nil
}

func ResourceCockpitAlertManagerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)

	if d.HasChange("enable_managed_alerts") {
		enable := d.Get("enable_managed_alerts").(bool)
		if enable {
			_, err = api.EnableManagedAlerts(&cockpit.RegionalAPIEnableManagedAlertsRequest{
				Region:    region,
				ProjectID: projectID,
			})
		} else {
			_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
				Region:    region,
				ProjectID: projectID,
			}, scw.WithContext(ctx))
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("contact_points") {
		oldContactPointsInterface, newContactPointsInterface := d.GetChange("contact_points")
		oldContactPoints := oldContactPointsInterface.([]any)
		newContactPoints := newContactPointsInterface.([]any)

		oldContactMap := make(map[string]map[string]any)

		for _, oldCP := range oldContactPoints {
			cp := oldCP.(map[string]any)
			email := cp["email"].(string)
			oldContactMap[email] = cp
		}

		newContactMap := make(map[string]map[string]any)

		for _, newCP := range newContactPoints {
			cp := newCP.(map[string]any)
			email := cp["email"].(string)
			newContactMap[email] = cp
		}

		for email := range oldContactMap {
			if _, found := newContactMap[email]; !found {
				err := api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
					Region:    region,
					ProjectID: projectID,
					Email:     &cockpit.ContactPointEmail{To: email},
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		for email := range newContactMap {
			if _, found := oldContactMap[email]; !found {
				contactPointEmail := &cockpit.ContactPointEmail{To: email}

				_, err = api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
					Region:    region,
					ProjectID: projectID,
					Email:     contactPointEmail,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return ResourceCockpitAlertManagerRead(ctx, d, meta)
}

func ResourceCockpitAlertManagerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)

	contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	for _, cp := range contactPoints.ContactPoints {
		if cp.Email != nil {
			err = api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
				Region:    region,
				ProjectID: projectID,
				Email:     &cockpit.ContactPointEmail{To: cp.Email.To},
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DisableAlertManager(&cockpit.RegionalAPIDisableAlertManagerRequest{
		Region:    region,
		ProjectID: projectID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func ResourceCockpitAlertManagerID(region scw.Region, projectID string) (resourceID string) {
	return fmt.Sprintf("%s/%s/1", region, projectID)
}
