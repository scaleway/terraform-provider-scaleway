package identity

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func DefaultRegionAttribute() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		Description:       "The region of the resource",
		RequiredForImport: true,
	}
}

func DefaultZoneAttribute() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		Description:       "The zone of the resource",
		RequiredForImport: true,
	}
}

func DefaultProjectIDAttribute() *schema.Schema {
	return &schema.Schema{
		Type:              schema.TypeString,
		Description:       "The ID of the project (UUID format)",
		RequiredForImport: true,
	}
}
