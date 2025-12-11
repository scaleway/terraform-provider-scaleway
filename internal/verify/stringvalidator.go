package verify

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Validators for schema.StringAttribute{}

func IsStringUUID() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
		"must be a valid UUID",
	)
}

func IsStringUUIDOrUUIDWithLocality() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^([a-zA-Z]{2}-[a-zA-Z]{3}/)?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
		"must be a valid UUID or UUID with locality prefix (format: aa-aaa-<uuid>)",
	)
}

// IsStringRegion only raises a warning if the region is invalid
func IsStringRegionWithWarning(validRegions []string) validator.String {
	return ErrorToWarningValidator(
		stringvalidator.OneOf(validRegions...),
	)
}

// Converts errors from a validator into warnings
func ErrorToWarningValidator(validator validator.String) validator.String {
	return errorToWarningValidator{validator: validator}
}

type errorToWarningValidator struct {
	validator validator.String
}

func (v errorToWarningValidator) Description(ctx context.Context) string {
	return v.validator.Description(ctx)
}

func (v errorToWarningValidator) MarkdownDescription(ctx context.Context) string {
	return v.validator.MarkdownDescription(ctx)
}

func (v errorToWarningValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Create a new response to capture the original diagnostics
	validationResp := &validator.StringResponse{}

	// Run the original validator
	v.validator.ValidateString(ctx, req, validationResp)

	// Convert any errors to warnings
	for _, d := range validationResp.Diagnostics {
		if d.Severity() == diag.SeverityError {
			// Convert error to warning using the diag.NewWarningDiagnostic function
			warningDiag := diag.NewWarningDiagnostic(
				d.Summary(),
				d.Detail(),
			)
			resp.Diagnostics = append(resp.Diagnostics, warningDiag)
		} else {
			// Keep existing warnings or info
			resp.Diagnostics = append(resp.Diagnostics, d)
		}
	}
}
