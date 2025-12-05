package jobs

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	jobs "github.com/scaleway/scaleway-sdk-go/api/jobs/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceDefinition() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceJobDefinitionCreate,
		ReadContext:   ResourceJobDefinitionRead,
		UpdateContext: ResourceJobDefinitionUpdate,
		DeleteContext: ResourceJobDefinitionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The job name",
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The job description",
				Optional:    true,
			},
			"cpu_limit": {
				Type:        schema.TypeInt,
				Description: "CPU limit of the job",
				Required:    true,
			},
			"memory_limit": {
				Type:        schema.TypeInt,
				Description: "Memory limit of the job",
				Required:    true,
			},
			"image_uri": {
				Type:        schema.TypeString,
				Description: "Image URI to use for the job",
				Optional:    true,
			},
			"command": {
				Type:        schema.TypeString,
				Description: "Command to use for the job",
				Optional:    true,
			},
			"timeout": {
				Type:             schema.TypeString,
				Description:      "Timeout for the job in seconds",
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: dsf.Duration,
			},
			"env": {
				Type:        schema.TypeMap,
				Description: "Environment variables to pass to the job",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 1000),
				},
				ValidateDiagFunc: validation.MapKeyLenBetween(0, 100),
			},
			"cron": {
				Type:        schema.TypeList,
				Description: "Cron expression",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schedule": {
							Type:         schema.TypeString,
							Description:  "UNIX cron schedule to run job",
							Required:     true,
							RequiredWith: []string{"cron.0"},
						},
						"timezone": {
							Type:         schema.TypeString,
							Description:  "Timezone for the cron schedule, in tz database format (e.g., 'Europe/Paris').",
							Required:     true,
							RequiredWith: []string{"cron.0"},
						},
					},
				},
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
			"secret_reference": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A reference to a Secret Manager secret.",
				Set: func(v any) int {
					secret := v.(map[string]any)
					if secret["environment"] != "" {
						return schema.HashString(locality.ExpandID(secret["secret_id"].(string)) + secret["secret_version"].(string) + secret["environment"].(string))
					}

					return schema.HashString(locality.ExpandID(secret["secret_id"].(string)) + secret["secret_version"].(string) + secret["file"].(string))
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret_id": {
							Type:                  schema.TypeString,
							Description:           "The secret unique identifier, it could be formatted as region/UUID or UUID. In case the region is passed, it must be the same as the job definition.",
							Required:              true,
							DiffSuppressOnRefresh: true,
							DiffSuppressFunc:      dsf.Locality,
						},
						"secret_reference_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The secret reference UUID",
						},
						"secret_version": {
							Type:        schema.TypeString,
							Description: "The secret version.",
							Default:     "latest",
							Optional:    true,
						},
						"file": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "The absolute file path where the secret will be mounted.",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(/[^/]+)+$`), "must be an absolute path to the file"),
						},
						"environment": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "An environment variable containing the secret value.",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Z|0-9]+(_[A-Z|0-9]+)*$`), "environment variable must be composed of uppercase letters separated by an underscore"),
						},
					},
				},
			},
		},
	}
}

func ResourceJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &jobs.CreateJobDefinitionRequest{
		Region:               region,
		Name:                 types.ExpandOrGenerateString(d.Get("name").(string), "job"),
		CPULimit:             uint32(d.Get("cpu_limit").(int)),
		MemoryLimit:          uint32(d.Get("memory_limit").(int)),
		ImageURI:             d.Get("image_uri").(string),
		Command:              d.Get("command").(string),
		ProjectID:            d.Get("project_id").(string),
		EnvironmentVariables: types.ExpandMapStringString(d.Get("env")),
		Description:          d.Get("description").(string),
		CronSchedule:         expandJobDefinitionCron(d.Get("cron")).ToCreateRequest(),
	}

	if timeoutSeconds, ok := d.GetOk("timeout"); ok {
		duration, err := time.ParseDuration(timeoutSeconds.(string))
		if err != nil {
			return diag.Diagnostics{{
				Severity:      diag.Error,
				Summary:       "Invalid timeout, expected Go duration format",
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("timeout"),
			}}
		}

		req.JobTimeout = scw.NewDurationFromTimeDuration(duration)
	}

	definition, err := api.CreateJobDefinition(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if rawSecretReference, ok := d.GetOk("secret_reference"); ok {
		if err := CreateJobDefinitionSecret(ctx, api, expandJobDefinitionSecret(rawSecretReference), region, definition.ID); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(regional.NewIDString(region, definition.ID))

	return ResourceJobDefinitionRead(ctx, d, m)
}

func ResourceJobDefinitionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	definition, err := api.GetJobDefinition(&jobs.GetJobDefinitionRequest{
		JobDefinitionID: id,
		Region:          region,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	rawSecretRefs, err := api.ListJobDefinitionSecrets(&jobs.ListJobDefinitionSecretsRequest{
		Region:          region,
		JobDefinitionID: id,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", definition.Name)
	_ = d.Set("cpu_limit", int(definition.CPULimit))
	_ = d.Set("memory_limit", int(definition.MemoryLimit))
	_ = d.Set("image_uri", definition.ImageURI)
	_ = d.Set("command", definition.Command)
	_ = d.Set("env", types.FlattenMap(definition.EnvironmentVariables))
	_ = d.Set("description", definition.Description)
	_ = d.Set("timeout", definition.JobTimeout.ToTimeDuration().String())
	_ = d.Set("cron", flattenJobDefinitionCron(definition.CronSchedule))
	_ = d.Set("region", definition.Region)
	_ = d.Set("project_id", definition.ProjectID)
	_ = d.Set("secret_reference", flattenJobDefinitionSecret(rawSecretRefs.Secrets))

	return nil
}

func ResourceJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &jobs.UpdateJobDefinitionRequest{
		Region:          region,
		JobDefinitionID: id,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("cpu_limit") {
		req.CPULimit = types.ExpandUint32Ptr(d.Get("cpu_limit"))
	}

	if d.HasChange("memory_limit") {
		req.MemoryLimit = types.ExpandUint32Ptr(d.Get("memory_limit"))
	}

	if d.HasChange("image_uri") {
		req.ImageURI = types.ExpandUpdatedStringPtr(d.Get("image_uri"))
	}

	if d.HasChange("command") {
		req.Command = types.ExpandUpdatedStringPtr(d.Get("command"))
	}

	if d.HasChange("env") {
		req.EnvironmentVariables = types.ExpandMapPtrStringString(d.Get("env"))
	}

	if d.HasChange("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}

	if d.HasChange("timeout") {
		if timeoutSeconds, ok := d.GetOk("timeout"); ok {
			duration, err := time.ParseDuration(timeoutSeconds.(string))
			if err != nil {
				return diag.Diagnostics{{
					Severity:      diag.Error,
					Summary:       "Invalid timeout, expected Go duration format",
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("timeout"),
				}}
			}

			req.JobTimeout = scw.NewDurationFromTimeDuration(duration)
		}
	}

	if d.HasChange("cron") {
		req.CronSchedule = expandJobDefinitionCron(d.Get("cron")).ToUpdateRequest()
	}

	if d.HasChange("secret_reference") {
		oldRawSecretRefs, newRawSecretRefs := d.GetChange("secret_reference")

		oldSecretRefs := expandJobDefinitionSecret(oldRawSecretRefs)
		newSecretRefs := expandJobDefinitionSecret(newRawSecretRefs)

		toCreate, toDelete, err := DiffJobDefinitionSecrets(oldSecretRefs, newSecretRefs)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, secret := range toDelete {
			deleteReq := &jobs.DeleteJobDefinitionSecretRequest{
				Region:          region,
				JobDefinitionID: id,
				SecretID:        secret.SecretReferenceID,
			}
			if err := api.DeleteJobDefinitionSecret(deleteReq, scw.WithContext(ctx)); err != nil {
				return diag.FromErr(err)
			}
		}

		if len(toCreate) > 0 {
			if err := CreateJobDefinitionSecret(ctx, api, toCreate, region, id); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if _, err := api.UpdateJobDefinition(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceJobDefinitionRead(ctx, d, m)
}

func ResourceJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Try to delete the job definition first
	err = api.DeleteJobDefinition(&jobs.DeleteJobDefinitionRequest{
		Region:          region,
		JobDefinitionID: id,
	}, scw.WithContext(ctx))

	// If deletion fails with 403 (job runs still running), clean them up first
	if err != nil && httperrors.Is403(err) {
		// List all job runs for this job definition
		jobRuns, listErr := api.ListJobRuns(&jobs.ListJobRunsRequest{
			Region:          region,
			JobDefinitionID: scw.StringPtr(id),
		}, scw.WithContext(ctx))
		if listErr != nil {
			return diag.FromErr(fmt.Errorf("failed to list job runs before cleanup: %w", listErr))
		}

		// Stop all running or queued job runs
		var jobRunIDsToWait []string
		for _, jobRun := range jobRuns.JobRuns {
			if jobRun.State == jobs.JobRunStateQueued || jobRun.State == jobs.JobRunStateRunning {
				_, stopErr := api.StopJobRun(&jobs.StopJobRunRequest{
					JobRunID: jobRun.ID,
					Region:   region,
				}, scw.WithContext(ctx))
				if stopErr != nil && !httperrors.Is404(stopErr) {
					return diag.FromErr(fmt.Errorf("failed to stop job run %s: %w", jobRun.ID, stopErr))
				}
				jobRunIDsToWait = append(jobRunIDsToWait, jobRun.ID)
			}
		}

		// Wait for all stopped job runs to terminate
		for _, jobRunID := range jobRunIDsToWait {
			_, waitErr := api.WaitForJobRun(&jobs.WaitForJobRunRequest{
				JobRunID: jobRunID,
				Region:   region,
			}, scw.WithContext(ctx))
			if waitErr != nil && !httperrors.Is404(waitErr) {
				return diag.FromErr(fmt.Errorf("failed to wait for job run %s: %w", jobRunID, waitErr))
			}
		}

		// Retry deletion after cleanup
		err = api.DeleteJobDefinition(&jobs.DeleteJobDefinitionRequest{
			Region:          region,
			JobDefinitionID: id,
		}, scw.WithContext(ctx))
	}

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
