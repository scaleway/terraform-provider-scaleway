package functions

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func NewIDFromRegionalID() function.Function {
	return &idPartFunction{
		name:        "id_from_regional_id",
		summary:     "Extract the ID without region from a regional ID",
		description: "Given a regional ID string value, returns the ID without the region prefix.",
		paramName:   "regional_id",
		paramDesc:   "regional ID to extract the ID from",
		skipName:    "skip_region_validation",
		skipDesc:    "If true, will skip region validation with the region known by the Scaleway SDK.",
		resolve: func(idParts []string) (string, error) {
			_, err := scw.ParseRegion(idParts[0])

			return strings.Join(idParts[1:], "/"), err
		},
	}
}
