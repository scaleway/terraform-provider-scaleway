package scaleway

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewaySecretVersion() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewaySecretVersion().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "region", "revision")
	dsSchema["secret_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the secret",
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
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
		Description: "The payload of the secret version",
	}

	return &schema.Resource{
		ReadContext: datasourceSchemaFromResourceVersionSchema,
		Schema:      dsSchema,
	}
}

func datasourceSchemaFromResourceVersionSchema(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := secretAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	var secretVersionIDStr string
	var payloadSecretRaw []byte

	secretID, existSecretID := d.GetOk("secret_id")
	if !existSecretID {
		request := &secret.AccessSecretVersionByNameRequest{
			Region:     region,
			SecretName: d.Get("secret_name").(string),
			Revision:   d.Get("revision").(string),
		}

		res, err := api.AccessSecretVersionByName(request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		secretVersionIDStr = newRegionalIDString(region, fmt.Sprintf("%s/%d", res.SecretID, res.Revision))
		_ = d.Set("secret_id", newRegionalIDString(region, res.SecretID))
		payloadSecretRaw = res.Data
	} else {
		request := &secret.AccessSecretVersionRequest{
			Region:   region,
			SecretID: expandID(secretID),
			Revision: d.Get("revision").(string),
		}

		res, err := api.AccessSecretVersion(request, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		secretVersionIDStr = newRegionalIDString(region, fmt.Sprintf("%s/%d", res.SecretID, res.Revision))
		payloadSecretRaw = res.Data
	}

	d.SetId(secretVersionIDStr)
	err = d.Set("data", base64.StdEncoding.EncodeToString(payloadSecretRaw))
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceScalewaySecretVersionRead(ctx, d, meta)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read secret version")...)
	}

	if d.Id() == "" {
		return diag.Errorf("secret version (%s) not found", secretVersionIDStr)
	}

	return nil
}
