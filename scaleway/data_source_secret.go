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

	return &schema.Resource{
		ReadContext: dataSourceScalewaySecretRead,

		Schema: dsSchema,
	}
}

func dataSourceScalewaySecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := secretAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	secretID, ok := d.GetOk("secret_id")
	if !ok {
		res, err := api.ListSecrets(&secret.ListSecretsRequest{
			Region:    region,
			Name:      expandStringPtr(d.Get("name")),
			ProjectID: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		for _, secret := range res.Secrets {

			if secret.Name == d.Get("name").(string) {
				if secretID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 secret found with the same name %s", d.Get("name")))
				}

				secretID = secret.ID
			}
		}

		if secretID == "" {
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
