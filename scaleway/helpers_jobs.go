package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	jobs "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// jobsAPIWithRegion returns a new jobs API and the region for a Create request
func jobsAPIWithRegion(d *schema.ResourceData, m interface{}) (*jobs.API, scw.Region, error) {
	meta := m.(*Meta)
	jobsAPI := jobs.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return jobsAPI, region, nil
}

// jobsAPIWithRegionalAndID returns a new jobs API with region and ID extracted from the state
func jobsAPIWithRegionAndID(m interface{}, regionalID string) (*jobs.API, scw.Region, string, error) {
	meta := m.(*Meta)
	jobsAPI := jobs.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return jobsAPI, region, ID, nil
}
