package verify

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/robfig/cron/v3"
)

func ValidateCronExpression() schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			diags := diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "expected type string",
				AttributePath: path,
			}}

			return diags
		}

		_, err := cron.ParseStandard(v)
		if err != nil {
			diags := diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "should be an valid Cron expression",
				AttributePath: path,
			}}

			return diags
		}

		return nil
	}
}
