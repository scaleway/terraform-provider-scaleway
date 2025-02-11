package verify

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type EnumValues[T ~string] interface {
	~string
	Values() []T
}

// ValidateEnum creates a schema validation function for the provided enum type
func ValidateEnum[T EnumValues[T]]() schema.SchemaValidateDiagFunc {
	values := filterUnknownValues(getValues[T]())

	return validation.ToDiagFunc(validation.StringInSlice(values, false))
}

// ValidateEnumIgnoreCase creates a schema validation function for the provided enum type with case-insensitive validation
func ValidateEnumIgnoreCase[T EnumValues[T]]() schema.SchemaValidateDiagFunc {
	values := filterUnknownValues(getValues[T]())

	return validation.ToDiagFunc(validation.StringInSlice(values, true))
}

func getValues[T EnumValues[T]]() []string {
	var t T
	values := t.Values()
	result := make([]string, len(values))

	for i, v := range values {
		result[i] = string(v)
	}

	return result
}

// filterUnknownValues removes "unknown" and "unknown_*" values from the slice
func filterUnknownValues(values []string) []string {
	filtered := make([]string, 0, len(values))

	for _, v := range values {
		if v == "unknown" || strings.HasPrefix(v, "unknown_") {
			continue
		}

		filtered = append(filtered, v)
	}

	return filtered
}
