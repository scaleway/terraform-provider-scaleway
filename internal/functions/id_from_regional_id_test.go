package functions_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/functions"
)

func TestIDFromRegionalIDFunctionRun(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expected function.RunResponse
		request  function.RunRequest
	}{
		// The example implementation uses the Go built-in string type, however
		// if AllowNullValue was enabled and *string or types.String was used,
		// this test case shows how the function would be expected to behave.
		"null": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringNull(), types.BoolNull()}),
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
				Arguments: function.NewArgumentsData([]attr.Value{types.StringUnknown(), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		// Test valid ID format - extracts ID without region
		"valid-regional-id-format": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par/1111-1111-1111-1111-1111111111111111"), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("1111-1111-1111-1111-1111111111111111")),
			},
		},
		// Test another valid regional ID format
		"valid-regional-id-format-amsterdam": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams/1111-1111-1111-1111-1111111111111111"), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("1111-1111-1111-1111-1111111111111111")),
			},
		},
		"valid-regional-id-multi-part": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams/foo/bar"), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("foo/bar")),
			},
		},
		"unknown-region-valid-format": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("xx-yyy/11111111-1111-1111-1111-111111111111"), types.BoolValue(true)}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("11111111-1111-1111-1111-111111111111")),
			},
		},
		// Test invalid format - empty string
		"empty-string": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue(""), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "bad regional ID format, expected format is: region/id"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		// Test invalid format - malformed ID
		"malformed-id": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("invalid-format"), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "cannot parse ID: invalid format"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			}

			functions.NewIDFromRegionalID().Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccProviderFunction_ID_From_Regional_ID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Can get the ID without region from a resource's regional id in one step
				Config: `
# terraform block required for provider function to be found
terraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
}

resource "scaleway_secret" "main" {
	name = "terraform_test_id_from_regional_id"
}

output "id" {
  value = provider::scaleway::id_from_regional_id(scaleway_secret.main.id)
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput("id", acctest.UUIDRegex),
				),
			},
		},
	})
}
