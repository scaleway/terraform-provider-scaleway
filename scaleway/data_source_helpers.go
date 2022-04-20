package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func datasourceNewZonedID(idI interface{}, fallBackZone scw.Zone) string {
	zone, id, err := parseZonedID(idI.(string))
	if err != nil {
		id = idI.(string)
		zone = fallBackZone
	}

	return newZonedIDString(zone, id)
}

func datasourceNewRegionalizedID(idI interface{}, fallBackRegion scw.Region) string {
	region, id, err := parseRegionalID(idI.(string))
	if err != nil {
		id = idI.(string)
		region = fallBackRegion
	}

	return newRegionalIDString(region, id)
}

////
// The below methods are imported from Google's terraform provider.
// source: https://github.com/terraform-providers/terraform-provider-google/blob/master/google/datasource_helpers.go
////

// datasourceSchemaFromResourceSchema is a recursive func that
// converts an existing Resource schema to a Datasource schema.
// All schema elements are copied, but certain attributes are ignored or changed:
// - all attributes have Computed = true
// - all attributes have ForceNew, Required = false
// - Validation funcs and attributes (e.g. MaxItems) are not copied
func datasourceSchemaFromResourceSchema(rs map[string]*schema.Schema) map[string]*schema.Schema {
	ds := make(map[string]*schema.Schema, len(rs))
	for k, v := range rs {
		dv := &schema.Schema{
			Computed:    true,
			ForceNew:    false,
			Description: v.Description,
			Type:        v.Type,
		}

		switch v.Type {
		case schema.TypeSet:
			dv.Set = v.Set
			fallthrough
		case schema.TypeList:
			// List & Set types are generally used for 2 cases:
			// - a list/set of simple primitive values (e.g. list of strings)
			// - a sub resource
			if elem, ok := v.Elem.(*schema.Resource); ok {
				// handle the case where the Element is a sub-resource
				dv.Elem = &schema.Resource{
					Schema: datasourceSchemaFromResourceSchema(elem.Schema),
				}
			} else {
				// handle simple primitive case
				dv.Elem = v.Elem
			}

		default:
			// Elem of all other types are copied as-is
			dv.Elem = v.Elem
		}
		ds[k] = dv
	}
	return ds
}

// fixDatasourceSchemaFlags is a convenience func that toggles the Computed,
// Optional + Required flags on a schema element. This is useful when the schema
// has been generated (using `datasourceSchemaFromResourceSchema` above for
// example) and therefore the attribute flags were not set appropriately when
// first added to the schema definition. Currently only supports top-level
// schema elements.
func fixDatasourceSchemaFlags(schema map[string]*schema.Schema, required bool, keys ...string) {
	for _, v := range keys {
		schema[v].Computed = false
		schema[v].Optional = !required
		schema[v].Required = required
	}
}

func addOptionalFieldsToSchema(schema map[string]*schema.Schema, keys ...string) {
	fixDatasourceSchemaFlags(schema, false, keys...)
}
