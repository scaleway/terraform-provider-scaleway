package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceApplication() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceApplication().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "name")

	dsSchema["name"].ConflictsWith = []string{"application_id"}
	dsSchema["application_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the IAM application",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUID(),
	}
	dsSchema["organization_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The organization_id the application is associated to",
		Optional:    true,
	}

	return &schema.Resource{
		ReadContext: DataSourceIamApplicationRead,
		Schema:      dsSchema,
	}
}

func DataSourceIamApplicationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewAPI(m)

	appID, appIDExists := d.GetOk("application_id")

	if !appIDExists {
		applicationName := d.Get("name").(string)

		res, err := api.ListApplications(&iam.ListApplicationsRequest{
			OrganizationID: types.FlattenStringPtr(account.GetOrganizationID(m, d)).(string),
			Name:           types.ExpandStringPtr(applicationName),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundApp, err := datasource.FindExact(
			res.Applications,
			func(s *iam.Application) bool { return s.Name == applicationName },
			applicationName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		appID = foundApp.ID
	}

	d.SetId(appID.(string))

	err := d.Set("application_id", appID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceIamApplicationRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read iam application state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("iam application (%s) not found", appID)
	}

	return nil
}
