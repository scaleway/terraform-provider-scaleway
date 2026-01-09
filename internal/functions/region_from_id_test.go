package functions_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/functions"
)

func TestRegionFromIDFunctionRun(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  function.RunRequest
		expected function.RunResponse
	}{
		// The example implementation uses the Go built-in string type, however
		// if AllowNullValue was enabled and *string or types.String was used,
		// this test case shows how the function would be expected to behave.
		"null": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringNull()),
			},
		},
		// The example implementation uses the Go built-in string type, however
		// if AllowUnknownValues was enabled and types.String was used,
		// this test case shows how the function would be expected to behave.
		"unknown": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringUnknown()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		// Test valid ID format - extracts region from ID
		"valid-id-format": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par/1111-1111-1111-1111-1111111111111111")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("fr-par")),
			},
		},
		// Test another valid ID format
		"valid-id-format-amsterdam": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams/1111-1111-1111-1111-1111111111111111")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("nl-ams")),
			},
		},
		"valid-id-multi-part": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams/foo/bar")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("nl-ams")),
			},
		},
		// Test invalid format - empty string
		"empty-string": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "cannot parse empty ID"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		// Test invalid format - malformed ID
		"malformed-id": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("invalid-format")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "cannot parse ID: invalid format"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			}

			functions.NewRegionFromID().Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
