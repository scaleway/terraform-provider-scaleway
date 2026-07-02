package keymanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeyRead,
		SchemaFunc:  dataSourceKeySchema,
	}
}

func dataSourceKeySchema() map[string]*schema.Schema {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceKeyManagerKey().SchemaFunc())

	datasource.AddOptionalFieldsToSchema(dsSchema, "region")

	dsSchema["key_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		Description:      "The ID of the key",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return dsSchema
}

func dataSourceKeyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	keyID := d.Get("key_id")

	fallbackRegion, err := meta.ExtractRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := datasource.NewRegionalID(keyID, fallbackRegion)
	d.SetId(regionalID)

	err = d.Set("key_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	client, region, parsedKeyID, err := NewKeyManagerAPIWithRegionAndID(m, regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	key, err := client.GetKey(&key_manager.GetKeyRequest{
		Region: region,
		KeyID:  parsedKeyID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	setKeyState(d, key)

	return nil
}
