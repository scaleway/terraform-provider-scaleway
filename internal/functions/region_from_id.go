package functions

import (
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func NewRegionFromID() function.Function {
	return &idPartFunction{
		name:        "region_from_id",
		summary:     "Extract a region from the ID",
		description: "Given an ID string value, returns the region contained in the ID.",
		paramName:   "id",
		paramDesc:   "id to extract the region from",
		skipName:    "skip_region_validation",
		skipDesc:    "If true, will skip region validation with the region known by the Scaleway SDK.",
		resolve: func(idParts []string) (string, error) {
			region, err := scw.ParseRegion(idParts[0])

			return region.String(), err
		},
	}
}
