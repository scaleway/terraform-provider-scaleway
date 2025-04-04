package jobs

import (
	"fmt"

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

type JobDefinitionSecret struct {
	SecretReferenceID string
	SecretID          string
	SecretVersion     string
	File              string
	Environment       string
}

func expandJobDefinitionSecret(i any) []JobDefinitionSecret {
	parsedSecrets := []JobDefinitionSecret{}

	if i == nil {
		return parsedSecrets
	}

	for _, rawSecret := range i.(*schema.Set).List() {
		secretMap := rawSecret.(map[string]interface{})
		env, file := "", ""

		if userEnv, ok := secretMap["environment"].(string); ok {
			env = userEnv
		}

		if userFile, ok := secretMap["file"].(string); ok {
			file = userFile
		}

		secret := JobDefinitionSecret{
			SecretID:      secretMap["secret_id"].(string),
			SecretVersion: secretMap["secret_version"].(string),
			File:          file,
			Environment:   env,
		}
		if v, ok := secretMap["secret_reference_id"]; ok {
			secret.SecretReferenceID = v.(string)
		}

		parsedSecrets = append(parsedSecrets, secret)
	}

	return parsedSecrets
}

func CreateJobDefinitionSecret(rawSecretReference any, api *jobs.API, region scw.Region, jobID string) error {
	parsedSecretReferences := expandJobDefinitionSecret(rawSecretReference)
	secrets := []*jobs.CreateJobDefinitionSecretsRequestSecretConfig{}

	for _, parsedSecretRef := range parsedSecretReferences {
		var secretConfig *jobs.CreateJobDefinitionSecretsRequestSecretConfig

		if parsedSecretRef.Environment != "" {
			secretConfig = &jobs.CreateJobDefinitionSecretsRequestSecretConfig{
				SecretManagerID:      parsedSecretRef.SecretID,
				SecretManagerVersion: parsedSecretRef.SecretVersion,
				EnvVarName:           &parsedSecretRef.Environment,
			}
		}

		if parsedSecretRef.File != "" {
			secretConfig = &jobs.CreateJobDefinitionSecretsRequestSecretConfig{
				SecretManagerID:      parsedSecretRef.SecretID,
				SecretManagerVersion: parsedSecretRef.SecretVersion,
				Path:                 &parsedSecretRef.File,
			}
		}

		if parsedSecretRef.Environment != "" && parsedSecretRef.File != "" {
			return fmt.Errorf("the secret id %s must have exactly one mount point: file or environment", parsedSecretRef.SecretID)
		}

		if parsedSecretRef.Environment == "" && parsedSecretRef.File == "" {
			return fmt.Errorf("the secret id %s is missing a mount point: file or environment", parsedSecretRef.SecretID)
		}

		secrets = append(secrets, secretConfig)
	}

	_, err := api.CreateJobDefinitionSecrets(&jobs.CreateJobDefinitionSecretsRequest{
		Region:          region,
		JobDefinitionID: jobID,
		Secrets:         secrets,
	})

	return err
}
