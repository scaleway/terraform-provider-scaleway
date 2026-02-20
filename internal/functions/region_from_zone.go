package functions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ function.Function = &RegionFromZone{}

type RegionFromZone struct{}

func NewRegionFromZone() function.Function {
	return &RegionFromZone{}
}

func (f *RegionFromZone) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "region_from_zone"
}

func (f *RegionFromZone) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Extract a region from a zone",
		Description: "Given a zone string value, returns the region that contains the zone.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "zone",
				Description: "zone to extract the region from (e.g. fr-par-1)",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *RegionFromZone) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input types.String

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))

	if input.IsNull() || input.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, input))

		return
	}

	zone, err := scw.ParseZone(input.ValueString())
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	region, err := zone.Region()
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(region.String())))
}
