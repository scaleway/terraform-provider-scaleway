package regional

import "github.com/scaleway/scaleway-sdk-go/scw"

type Model interface {
	GetRegion() string
}

// RegionsToQuery determines regions to query
func RegionsToQuery(model Model) []scw.Region {
	var regionsToQuery []scw.Region
	if model.GetRegion() == "all" {
		regionsToQuery = scw.AllRegions
	} else {
		regionsToQuery = []scw.Region{scw.Region(model.GetRegion())}
	}

	return regionsToQuery
}
