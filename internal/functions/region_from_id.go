package functions

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ function.Function = &RegionFromID{}

type RegionFromID struct{}

func NewRegionFromID() function.Function {
	return &RegionFromID{}
}

func (f *RegionFromID) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "region_from_id"
}

func (f *RegionFromID) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Extract a region from the ID",
		Description: "Given an ID string value, returns the region contained in the ID.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "id",
				Description: "id to extract the region from",
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

func (f *RegionFromID) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var (
		input                types.String
		skipRegionValidation types.Bool
	)

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &skipRegionValidation))

	if input.IsNull() || input.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, input))

		return
	}

	if input.ValueString() == "" {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "bad region format, available regions are: fr-par, nl-ams, pl-waw"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	if skipRegionValidation.IsNull() || skipRegionValidation.IsUnknown() {
		skipRegionValidation = basetypes.NewBoolValue(false)
	}

	idParts := strings.Split(input.ValueString(), "/")
	if len(idParts) < 2 {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "cannot parse ID: invalid format"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	region, err := scw.ParseRegion(idParts[0])
	if err != nil && !skipRegionValidation.ValueBool() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(region.String())))
}
