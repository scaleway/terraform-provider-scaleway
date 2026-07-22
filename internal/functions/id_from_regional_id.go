package functions

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ function.Function = &IDFromRegionalID{}

type IDFromRegionalID struct{}

func NewIDFromRegionalID() function.Function {
	return &IDFromRegionalID{}
}

func (f *IDFromRegionalID) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "id_from_regional_id"
}

func (f *IDFromRegionalID) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Extract the ID without region from a regional ID",
		Description: "Given a regional ID string value, returns the ID without the region prefix.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "regional_id",
				Description: "regional ID to extract the ID from",
			},
			function.BoolParameter{
				Name:           "skip_region_validation",
				Description:    "If true, will skip region validation with the region known by the Scaleway SDK.",
				AllowNullValue: true,
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *IDFromRegionalID) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var (
		regionalID           types.String
		skipRegionValidation types.Bool
	)

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &regionalID, &skipRegionValidation))

	if regionalID.IsNull() || regionalID.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, regionalID))

		return
	}

	if regionalID.ValueString() == "" {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "bad regional ID format, expected format is: region/id"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	if skipRegionValidation.IsNull() || skipRegionValidation.IsUnknown() {
		skipRegionValidation = basetypes.NewBoolValue(false)
	}

	idParts := strings.Split(regionalID.ValueString(), "/")
	if len(idParts) < 2 {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "cannot parse ID: invalid format"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	_, err := scw.ParseRegion(idParts[0])
	if err != nil && !skipRegionValidation.ValueBool() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(strings.Join(idParts[1:], "/"))))
}
