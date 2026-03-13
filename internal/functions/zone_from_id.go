package functions

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ function.Function = &ZoneFromID{}

type ZoneFromID struct{}

func NewZoneFromID() function.Function {
	return &ZoneFromID{}
}

func (f *ZoneFromID) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "zone_from_id"
}

func (f *ZoneFromID) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Extract a zone from the ID",
		Description: "Given an ID string value, returns the zone contained in the ID.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "id",
				Description: "id to extract the zone from",
			},
			function.BoolParameter{
				Name:           "skip_zone_validation",
				Description:    "If true, will skip zone validation with the zone known by the Scaleway SDK.",
				AllowNullValue: true,
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *ZoneFromID) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var (
		input              types.String
		skipZoneValidation types.Bool
	)

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input, &skipZoneValidation))

	if input.IsNull() || input.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, input))

		return
	}

	if input.ValueString() == "" {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "bad zone format, available zones are: fr-par-1, nl-ams-1, pl-waw-1"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	if skipZoneValidation.IsNull() || skipZoneValidation.IsUnknown() {
		skipZoneValidation = basetypes.NewBoolValue(false)
	}

	idParts := strings.Split(input.ValueString(), "/")
	if len(idParts) < 2 {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, "cannot parse ID: invalid format"))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	zone, err := scw.ParseZone(idParts[0])
	if err != nil && !skipZoneValidation.ValueBool() {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewStringUnknown()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.StringValue(zone.String())))
}
