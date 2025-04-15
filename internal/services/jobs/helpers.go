package jobs

import (
	"bytes"
	"context"
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
	SecretID          regional.ID
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
			SecretID:      regional.ExpandID(secretMap["secret_id"].(string)),
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

func flattenJobDefinitionSecret(jobSecrets []*jobs.Secret) []any {
	secretRefs := make([]interface{}, len(jobSecrets))

	for i, secret := range jobSecrets {
		secretRef := make(map[string]interface{})
		secretRef["secret_id"] = secret.SecretManagerID
		secretRef["secret_reference_id"] = secret.SecretID
		secretRef["secret_version"] = secret.SecretManagerVersion

		if secret.File != nil {
			secretRef["file"] = secret.File.Path
		}

		if secret.EnvVar != nil {
			secretRef["environment"] = secret.EnvVar.Name
		}

		secretRefs[i] = secretRef
	}

	return secretRefs
}

func CreateJobDefinitionSecret(ctx context.Context, api *jobs.API, jobSecrets []JobDefinitionSecret, region scw.Region, jobID string) error {
	secretConfigs := []*jobs.CreateJobDefinitionSecretsRequestSecretConfig{}

	for _, parsedSecretRef := range jobSecrets {
		secretConfig := &jobs.CreateJobDefinitionSecretsRequestSecretConfig{}

		secretConfigs = append(secretConfigs, secretConfig)

		if parsedSecretRef.SecretID.Region.String() != "" {
			if parsedSecretRef.SecretID.Region.String() != region.String() {
				return fmt.Errorf("the secret id %s is in the region %s, expected %s", parsedSecretRef.SecretID, parsedSecretRef.SecretID.Region, region)
			}
		}

		secretConfig.SecretManagerID = parsedSecretRef.SecretID.ID
		secretConfig.SecretManagerVersion = parsedSecretRef.SecretVersion

		if err := validateJobDefinitionSecret(&parsedSecretRef); err != nil {
			return err
		}

		if parsedSecretRef.Environment != "" {
			secretConfig.EnvVarName = &parsedSecretRef.Environment
		}

		if parsedSecretRef.File != "" {
			secretConfig.Path = &parsedSecretRef.File
		}
	}

	_, err := api.CreateJobDefinitionSecrets(&jobs.CreateJobDefinitionSecretsRequest{
		Region:          region,
		JobDefinitionID: jobID,
		Secrets:         secretConfigs,
	}, scw.WithContext(ctx))

	return err
}

func hashJobDefinitionSecret(secret *JobDefinitionSecret) int {
	buf := bytes.NewBufferString(secret.SecretID.String())
	buf.WriteString(secret.SecretVersion)

	if secret.Environment != "" {
		buf.WriteString(secret.Environment)
	}

	if secret.File != "" {
		buf.WriteString(secret.File)
	}

	return schema.HashString(buf.String())
}

func DiffJobDefinitionSecrets(oldSecretRefs, newSecretRefs []JobDefinitionSecret) (toCreate []JobDefinitionSecret, toDelete []JobDefinitionSecret, err error) {
	toCreate = make([]JobDefinitionSecret, 0)
	toDelete = make([]JobDefinitionSecret, 0)

	// hash the new and old secret sets
	oldSecretRefsMap := make(map[int]JobDefinitionSecret, len(oldSecretRefs))
	for _, secret := range oldSecretRefs {
		oldSecretRefsMap[hashJobDefinitionSecret(&secret)] = secret
	}

	newSecretRefsMap := make(map[int]JobDefinitionSecret, len(newSecretRefs))

	for _, secret := range newSecretRefs {
		if err := validateJobDefinitionSecret(&secret); err != nil {
			return toCreate, toDelete, err
		}

		newSecretRefsMap[hashJobDefinitionSecret(&secret)] = secret
	}

	// filter secrets to delete
	for hash, secret := range oldSecretRefsMap {
		if _, ok := newSecretRefsMap[hash]; !ok {
			toDelete = append(toDelete, secret)
		}
	}

	// filter secrets to create
	for hash, secret := range newSecretRefsMap {
		if _, ok := oldSecretRefsMap[hash]; !ok {
			toCreate = append(toCreate, secret)
		}
	}

	return toCreate, toDelete, nil
}

func validateJobDefinitionSecret(secret *JobDefinitionSecret) error {
	if secret == nil {
		return nil
	}

	if secret.Environment != "" && secret.File != "" {
		return fmt.Errorf("the secret id %s must have exactly one mount point: file or environment", secret.SecretID)
	}

	if secret.Environment == "" && secret.File == "" {
		return fmt.Errorf("the secret id %s is missing a mount point: file or environment", secret.SecretID)
	}

	return nil
}
