package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
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
		secretName := d.Get("name").(string)
		request := &secret.ListSecretsRequest{
			Region:         region,
			Name:           expandStringPtr(secretName),
			ProjectID:      expandStringPtr(projectID),
			OrganizationID: expandStringPtr(d.Get("organization_id")),
		}

		rawPath, pathExist := d.GetOk("path")
		if pathExist {
			request.Path = expandStringPtr(rawPath)
		}

		res, err := api.ListSecrets(request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundSecret, err := findExact(
			res.Secrets,
			func(s *secret.Secret) bool { return s.Name == secretName },
			secretName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		secretID = foundSecret.ID
	}

	regionalID := datasourceNewRegionalID(secretID, region)
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
