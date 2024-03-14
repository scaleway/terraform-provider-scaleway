package dsf

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
)

// Locality is a SuppressDiffFunc to remove the locality from an ID when checking diff.
// e.g. 2c1a1716-5570-4668-a50a-860c90beabf6 == fr-par-1/2c1a1716-5570-4668-a50a-860c90beabf6
func Locality(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	return locality.ExpandID(oldValue) == locality.ExpandID(newValue)
}
