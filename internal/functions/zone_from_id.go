package functions

import (
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func NewZoneFromID() function.Function {
	return &idPartFunction{
		name:        "zone_from_id",
		summary:     "Extract a zone from the ID",
		description: "Given an ID string value, returns the zone contained in the ID.",
		paramName:   "id",
		paramDesc:   "id to extract the zone from",
		skipName:    "skip_zone_validation",
		skipDesc:    "If true, will skip zone validation with the zone known by the Scaleway SDK.",
		resolve: func(idParts []string) (string, error) {
			zone, err := scw.ParseZone(idParts[0])

			return zone.String(), err
		},
	}
}
