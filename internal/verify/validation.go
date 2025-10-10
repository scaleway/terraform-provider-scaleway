package verify

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ValidateStringInSliceWithWarning helps to only returns warnings in case we got a non-public locality passed
func ValidateStringInSliceWithWarning(correctValues []string, field string) schema.SchemaValidateDiagFunc {
	return func(i any, path cty.Path) diag.Diagnostics {
		_, rawErr := validation.StringInSlice(correctValues, true)(i, field)

		var res diag.Diagnostics

		for _, e := range rawErr {
			res = append(res, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       e.Error(),
				AttributePath: path,
			})
		}

		return res
	}
}

type StructWithValues[T fmt.Stringer] interface {
	Values() []T
}

func ValidatorFromEnum[T fmt.Stringer](enum StructWithValues[T]) validator.String {
	enumValues := enum.Values()

	enumStringValues := make([]string, 0, len(enumValues))
	for _, enumValue := range enumValues {
		enumStringValues = append(enumStringValues, enumValue.String())
	}

	return stringvalidator.OneOf(enumStringValues...)
}
