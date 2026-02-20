package functions_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/functions"
)

func TestProjectFromIDFunctionRun(t *testing.T) {
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
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par/11111111-1111-1111-1111-111111111111/queue-name")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("11111111-1111-1111-1111-111111111111")),
			},
		},
		"valid-id-multi-part": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("nl-ams/22222222-2222-2222-2222-222222222222/topic-name/sub-id")}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.StringValue("22222222-2222-2222-2222-222222222222")),
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
		"malformed-id-no-slash": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("invalid-format")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "cannot parse ID: expected format is region/project_id/resource_id"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		"malformed-id-only-two-parts": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par/11111111-1111-1111-1111-111111111111")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "cannot parse ID: expected format is region/project_id/resource_id"),
				Result: function.NewResultData(types.StringUnknown()),
			},
		},
		"invalid-project-uuid": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.StringValue("fr-par/not-a-uuid/queue-name")}),
			},
			expected: function.RunResponse{
				Error:  function.NewArgumentFuncError(0, "cannot parse ID: project_id is not a valid UUID"),
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

			functions.NewProjectFromID().Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccProviderFunction_Project_From_ID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Can get the project from a resource's id in one step
				Config: `
# terraform block required for provider function to be found
terraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
}

resource "scaleway_cockpit_alert_manager" "main" {}

output "project" {
  value = provider::scaleway::project_from_id(scaleway_cockpit_alert_manager.main.id)
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchOutput("project", regexp.MustCompile(`[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`)),
				),
			},
		},
	})
}
