package list

import (
	"strings"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

func RegionalProjectTargets(regions []scw.Region, projects []string) []RegionalFetchTarget {
	targets := make([]RegionalFetchTarget, 0, len(regions)*len(projects))

	for _, r := range regions {
		for _, p := range projects {
			targets = append(targets, RegionalFetchTarget{Region: r, ProjectID: p})
		}
	}

	return targets
}

func ZonalProjectTargets(zones []scw.Zone, projects []string) []ZonalFetchTarget {
	targets := make([]ZonalFetchTarget, 0, len(zones)*len(projects))

	for _, z := range zones {
		for _, p := range projects {
			targets = append(targets, ZonalFetchTarget{Zone: z, ProjectID: p})
		}
	}

	return targets
}

func CompareZonalProjectItems(aProject, bProject string, aZone, bZone scw.Zone, aID, bID string) int {
	if aProject != bProject {
		return strings.Compare(aProject, bProject)
	}

	if aZone != bZone {
		return strings.Compare(string(aZone), string(bZone))
	}

	return strings.Compare(aID, bID)
}

func CompareRegionalProjectItems(aProject, bProject string, aRegion, bRegion scw.Region, aID, bID string) int {
	if aProject != bProject {
		return strings.Compare(aProject, bProject)
	}

	if aRegion != bRegion {
		return strings.Compare(string(aRegion), string(bRegion))
	}

	return strings.Compare(aID, bID)
}
