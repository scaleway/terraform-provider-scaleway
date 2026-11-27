package functions

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// idPartFunction is a function that splits an ID on "/", validates the first
// part against the Scaleway SDK (unless skip validation is requested) and
// returns a value computed from the ID parts.
type idPartFunction struct {
	resolve     func(idParts []string) (string, error)
	name        string
	summary     string
	description string
	paramName   string
	paramDesc   string
	skipName    string
	skipDesc    string
}

func (f *idPartFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f *idPartFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     f.summary,
		Description: f.description,
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        f.paramName,
				Description: f.paramDesc,
			},
			function.BoolParameter{
				Name:           f.skipName,
				Description:    f.skipDesc,
				AllowNullValue: true,
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *idPartFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var (
		id   types.String
		skip types.Bool
	)

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &id, &skip))

	if id.IsNull() || id.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, id))

		return
	}

	if skip.IsNull() || skip.IsUnknown() {
		skip = basetypes.NewBoolValue(false)
	}

	idParts := strings.Split(id.ValueString(), "/")
	if len(idParts) < 2 {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "cannot parse ID: invalid format"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	value, err := f.resolve(idParts)
	if err != nil && !skip.ValueBool() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(value)))
}
