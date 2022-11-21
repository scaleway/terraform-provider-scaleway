package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayIamApplication() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayIamApplication().Schema)

	addOptionalFieldsToSchema(dsSchema, "name")

	dsSchema["name"].ConflictsWith = []string{"application_id"}
	dsSchema["application_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the IAM application",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUID(),
	}
	dsSchema["organization_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "The organization_id you want to attach the resource to",
		Optional:    true,
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewayIamApplicationRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayIamApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	appID, appIDExists := d.GetOk("application_id")

	if !appIDExists {
		res, err := api.ListApplications(&iam.ListApplicationsRequest{
			OrganizationID: getOrganizationID(meta, d),
			Name:           expandStringPtr(d.Get("name")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, app := range res.Applications {
			if app.Name == d.Get("name").(string) {
				if appID != "" {
					return diag.Errorf("more than 1 application found with the same name %s", d.Get("name"))
				}
				appID = app.ID
			}
		}
		if appID == "" {
			return diag.Errorf("no application found with the name %s", d.Get("name"))
		}
	}

	d.SetId(appID.(string))
	err := d.Set("application_id", appID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewayIamApplicationRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read iam application state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("iam application (%s) not found", appID)
	}

	return nil
}
