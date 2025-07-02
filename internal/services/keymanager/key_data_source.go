package keymanager

import (
	"context"
	"fmt"

	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceKey() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceKey().Schema)
	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "project_id")

	dsSchema["name"].ConflictsWith = []string{"key_id"}
	dsSchema["key_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the Key",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceKeyRead,
		Schema:      dsSchema,
	}
}

func DataSourceKeyRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newKeyManagerAPI(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	keyID, ok := d.GetOk("key_id")
	if !ok {
		keyName := d.Get("name").(string)

		res, err := api.ListKeys(&key_manager.ListKeysRequest{
			Region:    region,
			Name:      types.ExpandStringPtr(keyName),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundKey, err := datasource.FindExact(
			res.Keys,
			func(s *key_manager.Key) bool { return s.Name == keyName },
			keyName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		keyID = foundKey.ID
	}

	regionalID := datasource.NewRegionalID(keyID, region)
	d.SetId(regionalID)

	err = d.Set("key_id", regionalID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if key exist as Read will return nil if resource does not exist
	// keyID may be regional if using name in data source
	getReq := &key_manager.GetKeyRequest{
		Region: region,
		KeyID:  locality.ExpandID(keyID.(string)),
	}

	_, err = api.GetKey(getReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(fmt.Errorf("no key found with the id %s", keyID))
	}

	return resourceKeyRead(ctx, d, m)
}
