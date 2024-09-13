package secret

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceSecret() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceSecret().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "region", "path")

	dsSchema["name"].ConflictsWith = []string{"secret_id"}
	dsSchema["path"].ConflictsWith = []string{"secret_id"}
	dsSchema["secret_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the secret",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    []string{"name", "path"},
	}
	dsSchema["organization_id"] = account.OrganizationIDOptionalSchema()
	dsSchema["project_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The project ID the resource is associated to",
		ValidateDiagFunc: verify.IsUUID(),
	}

	return &schema.Resource{
		ReadContext: DataSourceSecretRead,
		Schema:      dsSchema,
	}
}

func DataSourceSecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, projectID, err := newAPIWithRegionAndProjectID(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	secretID, ok := d.GetOk("secret_id")
	if !ok {
		secretName := d.Get("name").(string)
		request := &secret.ListSecretsRequest{
			Region:         region,
			Name:           types.ExpandStringPtr(secretName),
			ProjectID:      types.ExpandStringPtr(projectID),
			OrganizationID: types.ExpandStringPtr(d.Get("organization_id")),
			Path:           types.ExpandStringPtr(d.Get("path")),
		}

		res, err := api.ListSecrets(request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundSecret, err := datasource.FindExact(
			res.Secrets,
			func(s *secret.Secret) bool { return s.Name == secretName },
			secretName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		secretID = foundSecret.ID
	}

	regionalID := datasource.NewRegionalID(secretID, region)
	d.SetId(regionalID)
	err = d.Set("secret_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := ResourceSecretRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read secret")...)
	}

	if d.Id() == "" {
		return diag.Errorf("secret (%s) not found", regionalID)
	}

	return nil
}
