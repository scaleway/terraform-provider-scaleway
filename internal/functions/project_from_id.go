package functions

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

var _ function.Function = &ProjectFromID{}

type ProjectFromID struct{}

func NewProjectFromID() function.Function {
	return &ProjectFromID{}
}

func (f *ProjectFromID) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "project_from_id"
}

func (f *ProjectFromID) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Extract a project ID from the ID",
		Description: "Given an ID string value of format `region/project_id/resource_id`, returns the project ID contained in the ID.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "id",
				Description: "id to extract the project from",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *ProjectFromID) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.String

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))

	if input.IsNull() || input.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, input))

		return
	}

	if input.ValueString() == "" {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "cannot parse ID: invalid format"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	idParts := strings.Split(input.ValueString(), "/")
	if len(idParts) < 3 {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "cannot parse ID: expected format is region/project_id/resource_id"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	projectID := idParts[1]
	if !validation.IsUUID(projectID) {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "cannot parse ID: project_id is not a valid UUID"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(projectID)))
}
