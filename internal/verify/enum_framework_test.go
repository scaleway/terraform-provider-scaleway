package verify_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

type testEnum string

const (
	testEnumUnknownValue testEnum = "unknown_value"
	testEnumAlpha        testEnum = "alpha"
	testEnumBeta         testEnum = "beta"
)

func (e testEnum) Values() []testEnum {
	return []testEnum{
		testEnumUnknownValue,
		testEnumAlpha,
		testEnumBeta,
	}
}

func TestFrameworkValidateEnum(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	v := verify.FrameworkValidateEnum[testEnum]()

	testCases := map[string]struct {
		value   string
		wantErr bool
	}{
		"valid value alpha": {
			value:   "alpha",
			wantErr: false,
		},
		"valid value beta": {
			value:   "beta",
			wantErr: false,
		},
		"invalid value": {
			value:   "gamma",
			wantErr: true,
		},
		"unknown value is filtered and rejected": {
			value:   "unknown_value",
			wantErr: true,
		},
		"case mismatch is rejected": {
			value:   "Alpha",
			wantErr: true,
		},
		"empty string is rejected": {
			value:   "",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				ConfigValue: types.StringValue(tc.value),
			}

			resp := validator.StringResponse{}
			v.ValidateString(ctx, req, &resp)

			if tc.wantErr && !resp.Diagnostics.HasError() {
				t.Fatalf("expected error for value %q, got none", tc.value)
			}

			if !tc.wantErr && resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error for value %q: %s", tc.value, resp.Diagnostics[0].Summary())
			}
		})
	}
}

func TestFrameworkValidateEnumIgnoreCase(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	v := verify.FrameworkValidateEnumIgnoreCase[testEnum]()

	testCases := map[string]struct {
		value   string
		wantErr bool
	}{
		"valid value alpha": {
			value:   "alpha",
			wantErr: false,
		},
		"valid value beta": {
			value:   "beta",
			wantErr: false,
		},
		"mixed case Alpha accepted": {
			value:   "Alpha",
			wantErr: false,
		},
		"upper case BETA accepted": {
			value:   "BETA",
			wantErr: false,
		},
		"invalid value": {
			value:   "gamma",
			wantErr: true,
		},
		"unknown value is filtered and rejected": {
			value:   "unknown_value",
			wantErr: true,
		},
		"unknown value case-insensitive is rejected": {
			value:   "UNKNOWN_VALUE",
			wantErr: true,
		},
		"empty string is rejected": {
			value:   "",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := validator.StringRequest{
				ConfigValue: types.StringValue(tc.value),
			}

			resp := validator.StringResponse{}
			v.ValidateString(ctx, req, &resp)

			if tc.wantErr && !resp.Diagnostics.HasError() {
				t.Fatalf("expected error for value %q, got none", tc.value)
			}

			if !tc.wantErr && resp.Diagnostics.HasError() {
				t.Fatalf("unexpected error for value %q: %s", tc.value, resp.Diagnostics[0].Summary())
			}
		})
	}
}
