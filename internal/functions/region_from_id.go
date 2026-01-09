package functions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
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
		Description: "Given an ID string value, returns the region contain in the ID.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "id",
				Description: "id to extract the region from",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *RegionFromID) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input string

	// Read Terraform argument data into the variable
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))

	region, _, err := regional.ParseID(input)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, region))
}
