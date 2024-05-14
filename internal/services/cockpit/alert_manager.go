package cockpit

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCockpitAlertManager() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCockpitAlertManagerCreate,
		ReadContext:   ResourceCockpitAlertManagerRead,
		UpdateContext: ResourceCockpitAlertManagerUpdate,
		DeleteContext: ResourceCockpitAlertManagerDelete,
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Read:    schema.DefaultTimeout(DefaultCockpitTimeout),
			Delete:  schema.DefaultTimeout(DefaultCockpitTimeout),
			Default: schema.DefaultTimeout(DefaultCockpitTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"project_id": account.ProjectIDSchema(),
			"enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"emails": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: verify.IsEmail()},
				Optional:    true,
				Description: "A list of email addresses for the alert receivers",
			},
		},
	}
}

func ResourceCockpitAlertManagerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	enable := d.Get("enable").(bool)
	emails := d.Get("emails").([]interface{})

	_, err = api.EnableAlertManager(&cockpit.RegionalAPIEnableAlertManagerRequest{
		ProjectID: projectID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if enable {
		_, err = api.EnableManagedAlerts(&cockpit.RegionalAPIEnableManagedAlertsRequest{
			ProjectID: projectID,
		})
	} else {
		_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
			ProjectID: projectID,
		}, scw.WithContext(ctx))
	}

	if len(emails) > 0 {
		for _, email := range emails {
			emailStr, ok := email.(string)
			if !ok {
				return diag.FromErr(errors.New("invalid email format"))
			}
			emailCP := &cockpit.ContactPointEmail{
				To: emailStr,
			}
			_, err := api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
				ProjectID: projectID,
				Email:     emailCP,
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(ResourceCockpitAlertManagerID(region, projectID))
	return ResourceCockpitAlertManagerRead(ctx, d, meta)
}

func ResourceCockpitAlertManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	_ = d.Set("enable", alertManager.ManagedAlertsEnabled)

	contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
		Region:    region,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var emails []string
	for _, cp := range contactPoints.ContactPoints {
		if cp.Email != nil {
			emails = append(emails, cp.Email.To)
		}
	}
	_ = d.Set("emails", emails)
	return nil
}

func ResourceCockpitAlertManagerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	projectID := d.Get("project_id").(string)

	if d.HasChange("enable") {
		enable := d.Get("enable").(bool)
		if enable {
			_, err = api.EnableManagedAlerts(&cockpit.RegionalAPIEnableManagedAlertsRequest{
				ProjectID: projectID,
			})
		} else {
			ar, err := api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
				ProjectID: projectID,
			}, scw.WithContext(ctx))
			_ = ar
			_ = err
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("emails") {
		oldEmailsInterface, newEmailsInterface := d.GetChange("emails")
		oldEmails := convertInterfaceSliceToStringSlice(oldEmailsInterface.([]interface{}))
		newEmails := convertInterfaceSliceToStringSlice(newEmailsInterface.([]interface{}))

		for _, email := range oldEmails {
			if !contains(newEmails, email) {
				err := api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
					ProjectID: projectID,
					Email:     &cockpit.ContactPointEmail{To: email},
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		for _, email := range newEmails {
			if !contains(oldEmails, email) {
				_, err := api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
					ProjectID: projectID,
					Email:     &cockpit.ContactPointEmail{To: email},
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return ResourceCockpitAlertManagerRead(ctx, d, meta)
}

func ResourceCockpitAlertManagerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)

	contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	for _, cp := range contactPoints.ContactPoints {
		if cp.Email != nil {
			err := api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
				ProjectID: projectID,
				Email:     &cockpit.ContactPointEmail{To: cp.Email.To},
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	_, err = api.DisableManagedAlerts(&cockpit.RegionalAPIDisableManagedAlertsRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = api.DisableAlertManager(&cockpit.RegionalAPIDisableAlertManagerRequest{
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

func convertInterfaceSliceToStringSlice(input []interface{}) []string {
	result := make([]string, 0, len(input))
	for _, v := range input {
		result = append(result, v.(string))
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
