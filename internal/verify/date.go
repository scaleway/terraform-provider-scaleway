package verify

import (
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// IsDate will validate that field is a valid ISO 8601
// It is the same as RFC3339
func IsDate() schema.SchemaValidateDiagFunc {
	return func(value any, path cty.Path) diag.Diagnostics {
		date, isStr := value.(string)
		if !isStr {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: path,
				Summary:       "invalid input, expected a string",
			}}
		}

		_, err := time.Parse(time.RFC3339, date)
		if err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: path,
				Summary:       "invalid input, expected a valid RFC3339 date",
			}}
		}

		return nil
	}
}

func IsDuration() schema.SchemaValidateDiagFunc {
	return func(value any, path cty.Path) diag.Diagnostics {
		str, isStr := value.(string)
		if !isStr {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: path,
				Summary:       "invalid input, expected a string",
			}}
		}

		_, err := time.ParseDuration(str)
		if err != nil {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: path,
				Summary:       "invalid input, expected a valid duration",
			}}
		}

		return nil
	}
}
