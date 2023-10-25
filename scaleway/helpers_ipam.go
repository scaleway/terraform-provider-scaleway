package scaleway

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
)

// ipamAPIWithRegion returns a new ipam API and the region
func ipamAPIWithRegion(d *schema.ResourceData, m interface{}) (*ipam.API, scw.Region, error) {
	meta := m.(*Meta)
	ipamAPI := ipam.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return ipamAPI, region, nil
}

// expandLastID expand the last ID in a potential composed ID
// region/id1/id2 -> id2
// region/id1 -> id1
// region/id1/invalid -> id1
// id1 -> id1
// invalid -> invalid
func expandLastID(i interface{}) string {
	composedID := i.(string)
	elems := strings.Split(composedID, "/")
	for i := len(elems) - 1; i >= 0; i-- {
		if validation.IsUUID(elems[i]) {
			return elems[i]
		}
	}

	return composedID
}
