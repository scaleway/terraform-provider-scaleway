package identity

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func DefaultRegional() *schema.ResourceIdentity {
	return WrapSchemaMap(map[string]*schema.Schema{
		"id": {
			Type:              schema.TypeString,
			Description:       "The id of the resource (UUID format)",
			RequiredForImport: true,
		},
		"region": DefaultRegionAttribute(),
	})
}

func DefaultRegionAttribute() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		Description:       "The region of the resource",
		RequiredForImport: true,
	}
}

func DefaultZonal() *schema.ResourceIdentity {
	return WrapSchemaMap(map[string]*schema.Schema{
		"id": {
			Type:              schema.TypeString,
			Description:       "The id of the resource (UUID format)",
			RequiredForImport: true,
		},
		"zone": {
			Type:              schema.TypeString,
			Description:       "The zone of the resource",
			RequiredForImport: true,
		},
	})
}

func WrapSchemaMap(m map[string]*schema.Schema) *schema.ResourceIdentity {
	return &schema.ResourceIdentity{
		SchemaFunc: func() map[string]*schema.Schema {
			return m
		},
	}
}

func DefaultProjectIDAttribute() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		Description:       "The ID of the project (UUID format)",
		RequiredForImport: true,
	}
}

func DefaultProjectID() *schema.ResourceIdentity {
	return WrapSchemaMap(map[string]*schema.Schema{
		"project_id": DefaultProjectIDAttribute(),
	})
}
