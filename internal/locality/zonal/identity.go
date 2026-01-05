package zonal

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func NestedIdentity(key string) *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"id": {
					Type:              schema.TypeString,
					RequiredForImport: true,
				},
				"zone": {
					Type:              schema.TypeString,
					RequiredForImport: true,
				},
				key: {
					Type:              schema.TypeString,
					RequiredForImport: true,
				},
			}
		},
	}
}
