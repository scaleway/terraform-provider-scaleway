package jobs

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	jobs "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// newAPIWithRegion returns a new jobs API and the region for a Create request
func newAPIWithRegion(d *schema.ResourceData, m interface{}) (*jobs.API, scw.Region, error) {
	jobsAPI := jobs.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return jobsAPI, region, nil
}

// NewAPIWithRegionAndID returns a new jobs API with region and ID extracted from the state
func NewAPIWithRegionAndID(m interface{}, regionalID string) (*jobs.API, scw.Region, string, error) {
	jobsAPI := jobs.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return jobsAPI, region, ID, nil
}

type JobDefinitionCron struct {
	Schedule string
	Timezone string
}

func (c *JobDefinitionCron) ToCreateRequest() *jobs.CreateJobDefinitionRequestCronScheduleConfig {
	if c == nil {
		return nil
	}

	return &jobs.CreateJobDefinitionRequestCronScheduleConfig{
		Schedule: c.Schedule,
		Timezone: c.Timezone,
	}
}

func (c *JobDefinitionCron) ToUpdateRequest() *jobs.UpdateJobDefinitionRequestCronScheduleConfig {
	if c == nil {
		return &jobs.UpdateJobDefinitionRequestCronScheduleConfig{
			Schedule: nil,
			Timezone: nil,
		} // Send an empty update request to delete cron
	}

	return &jobs.UpdateJobDefinitionRequestCronScheduleConfig{
		Schedule: &c.Schedule,
		Timezone: &c.Timezone,
	}
}

func expandJobDefinitionCron(i any) *JobDefinitionCron {
	rawList := i.([]any)
	if len(rawList) == 0 {
		return nil
	}
	rawCron := rawList[0].(map[string]any)

	return &JobDefinitionCron{
		Schedule: rawCron["schedule"].(string),
		Timezone: rawCron["timezone"].(string),
	}
}

func flattenJobDefinitionCron(cron *jobs.CronSchedule) []any {
	if cron == nil {
		return []any{}
	}

	return []any{
		map[string]any{
			"schedule": cron.Schedule,
			"timezone": cron.Timezone,
		},
	}
}
