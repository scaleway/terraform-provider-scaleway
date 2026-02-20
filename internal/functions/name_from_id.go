package functions

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ function.Function = &NameFromID{}

type NameFromID struct{}

func NewNameFromID() function.Function {
	return &NameFromID{}
}

func (f *NameFromID) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "name_from_id"
}

func (f *NameFromID) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Extract a name from the ID",
		Description: "Given an ID string value, returns the name contained in the ID.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "id",
				Description: "id to extract the name from",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *NameFromID) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
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
	if len(idParts) < 2 {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "cannot parse ID: invalid format"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(idParts[len(idParts)-1])))
}
