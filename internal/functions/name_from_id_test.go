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

func TestNameFromIDFunctionRun(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expected function.RunResponse
		request  function.RunRequest
	}{
		"null": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringNull()),
			},
		},
		"unknown": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringUnknown()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		"valid-id-format": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par/11111111-1111-1111-1111-111111111111")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("11111111-1111-1111-1111-111111111111")),
			},
		},
		"valid-zonal-id": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par-1/11111111-1111-1111-1111-111111111111")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("11111111-1111-1111-1111-111111111111")),
			},
		},
		"valid-id-with-name": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams/11111111-1111-1111-1111-111111111111/my-name")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("my-name")),
			},
		},
		"empty-string": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "cannot parse ID: invalid format"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.RunResponse{
				Result: function.NewResultData(types.StringUnknown()),
			}

			functions.NewNameFromID().Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccProviderFunction_Name_From_ID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Can get the name from a static resource ID
				Config: `
# terraform block required for provider function to be found
terraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
}

output "name" {
  value = provider::scaleway::name_from_id("fr-par/terraform_test_name_from_id")
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("name", "terraform_test_name_from_id"),
				),
			},
		},
	})
}
