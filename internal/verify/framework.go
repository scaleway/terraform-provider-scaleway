package verify

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// StringAlsoRequiresOneOf validates that if the string attribute has a value,
// at least one of the specified sibling attributes must also have a value.
func StringAlsoRequiresOneOf(paths ...path.Expression) validator.String {
	return stringAlsoRequiresOneOfValidator{
		pathExpressions: paths,
	}
}

type stringAlsoRequiresOneOfValidator struct {
	pathExpressions path.Expressions
}

func (v stringAlsoRequiresOneOfValidator) Description(ctx context.Context) string {
	return "ensures at least one of the specified attributes is set when this attribute has a value"
}

func (v stringAlsoRequiresOneOfValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringAlsoRequiresOneOfValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Merge the path expressions with the current path expression
	expressions := req.PathExpression.MergeExpressions(v.pathExpressions...)

	// Check if at least one of the specified paths has a value
	atLeastOneSet := false

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			continue
		}

		for _, mp := range matchedPaths {
			// Skip the current attribute itself
			if mp.Equal(req.Path) {
				continue
			}

			var siblingValue attr.Value

			diags := req.Config.GetAttribute(ctx, mp, &siblingValue)
			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				continue
			}

			// Delay validation until the value is known
			if siblingValue.IsUnknown() {
				return
			}

			if !siblingValue.IsNull() {
				atLeastOneSet = true

				break
			}
		}

		if atLeastOneSet {
			break
		}
	}

	if !atLeastOneSet {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing required attribute",
			fmt.Sprintf("When %s is set, at least one of the identifier attributes (annotation_identifier or description_identifier) must also be set.", req.Path),
		)
	}
}
