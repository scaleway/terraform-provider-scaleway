package cockpit

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceCockpitContactPoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceCockpitContactPointCreate,
		ReadContext:   ResourceCockpitContactPointRead,
		UpdateContext: ResourceCockpitContactPointUpdate,
		DeleteContext: ResourceCockpitContactPointDelete,
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
			"email": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The email address of the alert receivers",
				ValidateFunc: verify.IsEmail(),
			},
		},
	}
}

func ResourceCockpitContactPointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	email := &cockpit.ContactPointEmail{
		To: d.Get("email").(string),
	}

	_, err = api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
		ProjectID: projectID,
		Email	 n :     email,
	}, scw.WithContext(ctx))

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cockpitIDWithProjectID(projectID, strconv.FormatUint(uint64(0), 10)))

	return ResourceCockpitContactPointRead(ctx, d, meta)
}

func ResourceCockpitContactPointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	api, _, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	email := d.Get("email").(string)

	contactPoints, err := api.ListContactPoints(&cockpit.RegionalAPIListContactPointsRequest{
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var found *cockpit.ContactPoint
	for _, cp := range contactPoints.ContactPoints {
		if cp.Email != nil && cp.Email.To == email {
			found = cp
			break
		}
	}

	if found == nil {
		d.SetId("")
		return nil
	}
	d.SetId(projectID)
	_ = d.Set("email", found.Email.To)
	return nil
}

func ResourceCockpitContactPointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	if d.HasChange("email") {
		oldEmail, newEmail := d.GetChange("email")

		err = api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
			ProjectID: projectID,
			Email:     &cockpit.ContactPointEmail{To: oldEmail.(string)},
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = api.CreateContactPoint(&cockpit.RegionalAPICreateContactPointRequest{
			ProjectID: projectID,
			Email:     &cockpit.ContactPointEmail{To: newEmail.(string)},
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceCockpitContactPointRead(ctx, d, meta)
}

func ResourceCockpitContactPointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := cockpitAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := d.Get("project_id").(string)
	email := &cockpit.ContactPointEmail{
		To: d.Get("email").(string),
	}
	err = api.DeleteContactPoint(&cockpit.RegionalAPIDeleteContactPointRequest{
		ProjectID: projectID,
		Email:     email,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
