package scaleway

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
)

type testCase struct {
	val         interface{}
	f           schema.SchemaValidateFunc
	expectedErr *regexp.Regexp
}

func runTestCases(t *testing.T, cases []testCase) {
	matchErr := func(errs []error, r *regexp.Regexp) bool {
		// err must match one provided
		for _, err := range errs {
			if r.MatchString(err.Error()) {
				return true
			}
		}

		return false
	}

	for i, tc := range cases {
		_, errs := tc.f(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchErr(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}

func TestValidationStringNotInSlice(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val:         "InvalidValue",
			f:           validationStringNotInSlice([]string{"InvalidValue", "AnotherInvalidValue"}, false),
			expectedErr: regexp.MustCompile("expected [\\w]+ not to be one of \\[InvalidValue AnotherInvalidValue\\], got InvalidValue"),
		},
		// ignore case
		{
			val:         "INVALIDVALUE",
			f:           validationStringNotInSlice([]string{"InvalidValue", "AnotherInvalidValue"}, true),
			expectedErr: regexp.MustCompile("expected [\\w]+ not to be one of \\[InvalidValue AnotherInvalidValue\\], got INVALIDVALUE"),
		},
		{
			val: "VALIDVALUE",
			f:   validationStringNotInSlice([]string{"ValidValue", "AnotherValidValue"}, false),
		},
		{
			val: "ValidValue",
			f:   validationStringNotInSlice([]string{"InvalidValue", "AnotherValidValue"}, false),
		},
		{
			val:         1,
			f:           validationStringNotInSlice([]string{"InvalidValue", "AnotherValidValue"}, false),
			expectedErr: regexp.MustCompile("expected type of [\\w]+ to be string"),
		},
	})
}
