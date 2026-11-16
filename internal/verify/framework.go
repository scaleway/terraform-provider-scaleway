package verify

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/robfig/cron/v3"
)

// Validators for schema.StringAttribute{}

func IsStringUUID() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
		"must be a valid UUID",
	)
}

func IsStringUUIDOrUUIDWithRegion() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^([a-zA-Z]{2}-[a-zA-Z]{3}/)?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
		"must be a valid UUID or UUID with region prefix (format: aa-aaa/<uuid>)",
	)
}

func IsStringUUIDOrUUIDWithZone() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^([a-zA-Z]{2}-[a-zA-Z]{3}-[0-9]/)?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
		"must be a valid UUID or UUID with zone prefix (format: aa-aaa-00/<uuid>)",
	)
}

func IsStringEmail() validator.String {
	return stringvalidator.RegexMatches(
		regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`),
		"must be a valid email address",
	)
}

// IsStringCronExpression validates a standard 5-field UNIX cron expression.
func IsStringCronExpression() validator.String {
	return stringCronValidator{}
}

type stringCronValidator struct{}

func (stringCronValidator) Description(_ context.Context) string {
	return "must be a valid UNIX cron expression"
}

func (v stringCronValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (stringCronValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if _, err := cron.ParseStandard(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Cron Expression",
			fmt.Sprintf("expected a valid UNIX cron expression: %s", err),
		)
	}
}

// IsStringTimezoneIdentifier validates an IANA timezone name (e.g. Europe/Paris).
func IsStringTimezoneIdentifier() validator.String {
	return stringTimezoneValidator{}
}

type stringTimezoneValidator struct{}

func (stringTimezoneValidator) Description(_ context.Context) string {
	return "must be a valid IANA timezone identifier (e.g. Europe/Paris)"
}

func (v stringTimezoneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (stringTimezoneValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if _, err := time.LoadLocation(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Timezone",
			fmt.Sprintf("expected a valid IANA timezone identifier (e.g. Europe/Paris): %s", err),
		)
	}
}

// MutuallyExclusiveStringConflicts builds a ConflictsWith validator listing every attribute in the group except `self`
func MutuallyExclusiveStringConflicts(self string, group ...string) []validator.String {
	conflicts := make([]path.Expression, 0, len(group))

	for _, name := range group {
		if name == self {
			continue
		}

		conflicts = append(conflicts, path.MatchRoot(name))
	}

	return []validator.String{stringvalidator.ConflictsWith(conflicts...)}
}

// IsStringOneOfWithWarning only raises a warning if the string is not oneOf validValues
func IsStringOneOfWithWarning(validValues []string) validator.String {
	return ErrorToWarningValidator(
		stringvalidator.OneOf(validValues...),
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

// Validators for schema.SetAttribute{}

func SetElemIsStringUUIDOrUUIDWithRegion() validator.Set {
	return setvalidator.ValueStringsAre(IsStringUUIDOrUUIDWithRegion())
}
