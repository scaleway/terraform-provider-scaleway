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

func CompareRegionalProjectItems(aProject, bProject string, aRegion, bRegion scw.Region, aID, bID string) int {
	if aProject != bProject {
		return strings.Compare(aProject, bProject)
	}

	if aRegion != bRegion {
		return strings.Compare(string(aRegion), string(bRegion))
	}

	return strings.Compare(aID, bID)
}
