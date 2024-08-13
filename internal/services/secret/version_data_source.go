package secret

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceVersion() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceVersion().Schema)

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "region", "revision")
	dsSchema["secret_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the secret",
		ValidateFunc:  verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith: []string{"secret_name"},
	}
	dsSchema["secret_name"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The Name of the secret",
		ConflictsWith: []string{"secret_id"},
	}
	dsSchema["data"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "The payload of the secret version",
	}
	dsSchema["organization_id"] = account.OrganizationIDOptionalSchema()
	dsSchema["project_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the project to filter the secret version",
		ValidateFunc: verify.IsUUID(),
	}

	return &schema.Resource{
		ReadContext: datasourceSchemaFromResourceVersionSchema,
		Schema:      dsSchema,
	}
}

func datasourceSchemaFromResourceVersionSchema(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	secretID, existSecretID := d.GetOk("secret_id")
	api, region, projectID, err := newAPIWithRegionProjectIDAndDefault(d, m, regional.ExpandID(secretID).Region)
	if err != nil {
		return diag.FromErr(err)
	}

	var secretVersionIDStr string
	var payloadSecretRaw []byte

	if !existSecretID {
		secretName := d.Get("secret_name").(string)
		secrets, err := api.ListSecrets(&secret.ListSecretsRequest{
			Region:         region,
			Name:           &secretName,
			ProjectID:      types.ExpandStringPtr(projectID),
			OrganizationID: types.ExpandStringPtr(d.Get("organization_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		foundSecret, err := datasource.FindExact(secrets.Secrets,
			func(s *secret.Secret) bool { return s.Name == secretName },
			secretName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		res, err := api.AccessSecretVersion(&secret.AccessSecretVersionRequest{
			Region:   region,
			SecretID: foundSecret.ID,
			Revision: d.Get("revision").(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		secretVersionIDStr = regional.NewIDString(region, fmt.Sprintf("%s/%d", res.SecretID, res.Revision))
		_ = d.Set("secret_id", regional.NewIDString(region, res.SecretID))
		payloadSecretRaw = res.Data
	} else {
		request := &secret.AccessSecretVersionRequest{
			Region:   region,
			SecretID: locality.ExpandID(secretID),
			Revision: d.Get("revision").(string),
		}

		res, err := api.AccessSecretVersion(request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		secretVersionIDStr = regional.NewIDString(region, fmt.Sprintf("%s/%d", res.SecretID, res.Revision))
		payloadSecretRaw = res.Data
	}

	d.SetId(secretVersionIDStr)
	err = d.Set("data", base64.StdEncoding.EncodeToString(payloadSecretRaw))
	if err != nil {
		return diag.FromErr(err)
	}

	diags := ResourceVersionRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read secret version")...)
	}

	if d.Id() == "" {
		return diag.Errorf("secret version (%s) not found", secretVersionIDStr)
	}

	return nil
}
