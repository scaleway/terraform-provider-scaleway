package verify

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

func IsEmail() schema.SchemaValidateDiagFunc {
	return func(value interface{}, path cty.Path) diag.Diagnostics {
		email, isString := value.(string)
		if !isString {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: path,
				Summary:       "invalid input, expected a string",
				Detail:        "got " + email,
			}}
		}

		if !validation.IsEmail(email) {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: path,
				Summary:       "invalid input, expected a valid email",
				Detail:        "got " + email,
			}}
		}

		return nil
	}
}

func IsEmailList() schema.SchemaValidateDiagFunc {
	return func(value interface{}, path cty.Path) diag.Diagnostics {
		list, ok := value.([]interface{})
		if !ok {
			return diag.Diagnostics{diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: path,
				Summary:       "invalid type, expecting a list of strings",
			}}
		}

		for _, li := range list {
			email, isString := li.(string)
			if !isString {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					AttributePath: path,
					Summary:       "invalid type, each item must be a string",
				}}
			}

			if emailDiags := IsEmail()(email, path); len(emailDiags) > 0 {
				return emailDiags
			}
		}

		return nil
	}
}
