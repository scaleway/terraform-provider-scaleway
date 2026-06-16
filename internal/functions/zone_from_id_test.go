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

func TestZoneFromIDFunctionRun(t *testing.T) {
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
		// Test valid ID format - extracts zone from ID
		"valid-id-format": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par-1/1111-1111-1111-1111-1111111111111111"), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("fr-par-1")),
			},
		},
		// Test another valid ID format
		"valid-id-format-amsterdam": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams-1/1111-1111-1111-1111-1111111111111111"), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("nl-ams-1")),
			},
		},
		"valid-id-multi-part": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams-1/foo/bar"), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("nl-ams-1")),
			},
		},
		"unknown-id-valid-format": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("xx-yyy-1/11111111-1111-1111-1111-111111111111"), types.BoolValue(true)}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("xx-yyy-1")),
			},
		},
		// Test invalid format - empty string
		"empty-string": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue(""), types.BoolNull()}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "bad zone format, available zones are: fr-par-1, nl-ams-1, pl-waw-1"),
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

			functions.NewZoneFromID().Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccProviderFunction_Zone_From_ID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Can get the zone from a resource's id in one step
				Config: `
# terraform block required for provider function to be found
erraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
}

resource "scaleway_instance_server" "main" {
	name = "terraform_test_zone_from_id"
	type = "DEV1-S"
	image = "fr-par-1/ubuntu_jammy"
	zone = "fr-par-1"
}

output "zone" {
  value = provider::scaleway::zone_from_id(scaleway_instance_server.main.id)
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("zone", "fr-par-1"),
				),
			},
		},
	})
}
