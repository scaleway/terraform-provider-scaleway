package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewaySecret() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewaySecret().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "region")

	dsSchema["name"].ConflictsWith = []string{"secret_id"}
	dsSchema["secret_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the secret",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"name"},
	}
	dsSchema["organization_id"] = organizationIDOptionalSchema()
	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The project ID the resource is associated to",
		ValidateFunc: validationUUID(),
	}

	return &schema.Resource{
		ReadContext: dataSourceScalewaySecretRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewaySecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, projectID, err := secretAPIWithRegionAndProjectID(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	secretID, ok := d.GetOk("secret_id")
	if !ok {
		request := &secret.ListSecretsRequest{
			Region: region,
			Name:   scw.StringPtr(d.Get("name").(string)),
		}

		request.ProjectID = scw.StringPtr(projectID)

		if organizationIDRaw, ok := d.GetOk("organization_id"); ok {
			request.OrganizationID = scw.StringPtr(organizationIDRaw.(string))
		}
		res, err := api.ListSecrets(request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, s := range res.Secrets {
			if s.Status == secret.SecretStatusLocked {
				continue
			}

			if s.Name == d.Get("name").(string) {
				if secretID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 secret found with the same name %s", d.Get("name")))
				}

				secretID = newRegionalIDString(region, s.ID)
			}
		}
		if res.TotalCount == 0 {
			return diag.FromErr(fmt.Errorf("no secret found with the name %s", d.Get("name")))
		}
	}

	regionalID := datasourceNewRegionalizedID(secretID, region)
	d.SetId(regionalID)
	err = d.Set("secret_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewaySecretRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read secret")...)
	}

	if d.Id() == "" {
		return diag.Errorf("secret (%s) not found", regionalID)
	}

	return nil
}
