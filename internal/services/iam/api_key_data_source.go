package iam

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

//go:embed descriptions/iam_api_key_data_source.md
var apikeyDataSourceDescription string

func DataSourceAPIKey() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceAPIKey().SchemaFunc())

	dsSchema["access_key"].Required = true
	dsSchema["access_key"].Computed = false
	delete(dsSchema, "secret_key")

	return &schema.Resource{
		Description: apikeyDataSourceDescription,
		ReadContext: DataSourceIamAPIKeyRead,
		Schema:      dsSchema,
	}
}

func DataSourceIamAPIKeyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	accessKey := d.Get("access_key").(string)

	d.SetId(accessKey)

	iamAPI := NewAPI(m)

	res, err := iamAPI.GetAPIKey(&iam.GetAPIKeyRequest{
		AccessKey: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	setAPIKeyState(d, res)

	return nil
}
