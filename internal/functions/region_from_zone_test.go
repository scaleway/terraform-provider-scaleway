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

func TestRegionFromZoneFunctionRun(t *testing.T) {
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
		"valid-zone-fr-par-1": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par-1")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("fr-par")),
			},
		},
		"valid-zone-nl-ams-2": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams-2")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("nl-ams")),
			},
		},
		"valid-zone-pl-waw-1": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("pl-waw-1")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("pl-waw")),
			},
		},
		"empty-string": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "scaleway-sdk-go: bad zone format, available zones are: fr-par-1, fr-par-2, fr-par-3, nl-ams-1, nl-ams-2, nl-ams-3, pl-waw-1, pl-waw-2, pl-waw-3"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		"invalid-zone": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("invalid-zone")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "scaleway-sdk-go: bad zone format, available zones are: fr-par-1, fr-par-2, fr-par-3, nl-ams-1, nl-ams-2, nl-ams-3, pl-waw-1, pl-waw-2, pl-waw-3"),
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

			functions.NewRegionFromZone().Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccProviderFunction_Region_From_Zone(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
# terraform block required for provider function to be found
terraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
}

output "region" {
  value = provider::scaleway::region_from_zone("fr-par-1")
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("region", "fr-par"),
				),
			},
		},
	})
}
