package verify_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func TestStringValidatorUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		value   string
		wantErr bool
		errDesc string
	}{
		"valid UUID": {
			value:   "123e4567-e89b-12d3-a456-426614174000",
			wantErr: false,
		},
		"valid UUID with all digits": {
			value:   "01234567-89ab-cdef-0123-456789abcdef",
			wantErr: false,
		},
		"invalid UUID - wrong format": {
			value:   "123e4567-e89b-12d3-a456-42661417400",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid UUID - wrong characters": {
			value:   "123e4567-e89b-12d3-a456-426614174xxx",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid UUID - too short": {
			value:   "123e4567-e89b-12d3-a456-426614174",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid UUID - too long": {
			value:   "123e4567-e89b-12d3-a456-4266141740000",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"empty string": {
			value:   "",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"non-hex characters": {
			value:   "123e4567-e89b-12d3-a456-42661417400g",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tc.value),
			}

			resp := validator.StringResponse{}

			verify.IsStringUUID().ValidateString(ctx, req, &resp)

			if tc.wantErr {
				if !resp.Diagnostics.HasError() {
					t.Fatal("expected error, got none")
				}

				if tc.errDesc != "" {
					errStr := resp.Diagnostics[0].Summary()
					if errStr != tc.errDesc {
						t.Fatalf("expected error description %q, got %q", tc.errDesc, errStr)
					}
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Fatalf("unexpected error: %v", resp.Diagnostics[0].Summary())
				}
			}
		})
	}
}

func TestStringValidatorUUIDOrUUIDWithLocality(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		value   string
		wantErr bool
		errDesc string
	}{
		"valid UUID": {
			value:   "123e4567-e89b-12d3-a456-426614174000",
			wantErr: false,
		},
		"valid UUID with all digits": {
			value:   "01234567-89ab-cdef-0123-456789abcdef",
			wantErr: false,
		},
		"invalid UUID - wrong format": {
			value:   "123e4567-e89b-12d3-a456-42661417400",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid UUID - wrong characters": {
			value:   "123e4567-e89b-12d3-a456-426614174xxx",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid UUID - too short": {
			value:   "123e4567-e89b-12d3-a456-426614174",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid UUID - too long": {
			value:   "123e4567-e89b-12d3-a456-4266141740000",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"empty string": {
			value:   "",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"non-hex characters": {
			value:   "123e4567-e89b-12d3-a456-42661417400g",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"valid UUID with locality": {
			value:   "qw-ert/01234567-89ab-cdef-0123-456789abcdef",
			wantErr: false,
		},
		"valid UUID with uppercase locality": {
			value:   "YU-IOP/123e4567-e89b-12d3-a456-426614174000",
			wantErr: false,
		},
		"invalid - locality with invalid delimiter": {
			value:   "qw/ert/123e4567-e89b-12d3-a456-426614174000",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid - locality with space": {
			value:   "qw ert/123e4567-e89b-12d3-a456-426614174000",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid - missing uuid": {
			value:   "qw-ert/",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid - UUID with locality containing special characters": {
			value:   "qw-@rt/123e4567-e89b-12d3-a456-426614174000",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"invalid - UUID with empty locality": {
			value:   "/123e4567-e89b-12d3-a456-426614174000",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"malformed UUID after valid prefix": {
			value:   "qw-ert/123e4567-e89b-12d3-a456-426614174xxx",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
		"malformed UUID after valid prefix with slash": {
			value:   "qw-ert/123e4567-e89b-12d3-a456-426614174000/extra",
			wantErr: true,
			errDesc: "Invalid Attribute Value Match",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tc.value),
			}

			resp := validator.StringResponse{}

			verify.IsStringUUIDOrUUIDWithLocality().ValidateString(ctx, req, &resp)

			if tc.wantErr {
				if !resp.Diagnostics.HasError() {
					t.Fatal("expected error, got none")
				}

				if tc.errDesc != "" {
					errStr := resp.Diagnostics[0].Summary()
					if errStr != tc.errDesc {
						t.Fatalf("expected error description %q, got %q", tc.errDesc, errStr)
					}
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Fatalf("unexpected error: %v", resp.Diagnostics[0].Summary())
				}
			}
		})
	}
}

func TestStringValidatorRegionWithWarning(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	validRegions := []string{"fr-par", "nl-ams", "pl-waw"}

	testCases := map[string]struct {
		value    string
		hasWarn  bool
		hasError bool
	}{
		"valid region": {
			value:    "fr-par",
			hasWarn:  false,
			hasError: false,
		},
		"invalid region": {
			value:    "qw-ert",
			hasWarn:  true,
			hasError: false, // Important: no error, only warning
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tc.value),
			}

			resp := validator.StringResponse{}

			// This does not have the error to warning func
			verify.IsStringRegionWithWarning(validRegions).ValidateString(ctx, req, &resp)

			var hasWarning bool
			for _, d := range resp.Diagnostics {
				if d.Severity() == diag.SeverityWarning {
					hasWarning = true
					break
				}
			}

			if hasWarning != tc.hasWarn {
				t.Fatalf("expected hasWarn=%v, got %v", tc.hasWarn, hasWarning)
			}

			for _, d := range resp.Diagnostics {
				if d.Severity() == diag.SeverityError {
					t.Fatalf("unexpected error: %s", d.Summary())
				}
			}
		})
	}
}
