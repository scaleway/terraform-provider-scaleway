package container

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/container/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceTrigger() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceContainerTriggerCreate,
		ReadContext:   ResourceContainerTriggerRead,
		UpdateContext: ResourceContainerTriggerUpdate,
		DeleteContext: ResourceContainerTriggerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultTriggerTimeout),
			Read:    schema.DefaultTimeout(defaultTriggerTimeout),
			Update:  schema.DefaultTimeout(defaultTriggerTimeout),
			Delete:  schema.DefaultTimeout(defaultTriggerTimeout),
			Create:  schema.DefaultTimeout(defaultTriggerTimeout),
		},
		SchemaVersion: 0,
		SchemaFunc:    triggerSchema,
		CustomizeDiff: customdiff.All(
			cdf.LocalityCheck("container_id"),
			forceNewOnSourceChange("sqs", "nats", "cron"),
		),
	}
}

func triggerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"container_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			Description:      "The ID of the container to create a trigger for",
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The trigger name",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The trigger description",
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "List of tags [\"tag1\", \"tag2\", ...] attached to the container trigger",
		},
		"destination_config": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Required:    true,
			Description: "Configuration of the destination to trigger.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"http_path": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The HTTP path to send the request to (e.g., \"/my-webhook-endpoint\").",
					},
					"http_method": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      "The HTTP method to use when sending the request (e.g., get, post, put, patch, delete).",
						ValidateDiagFunc: verify.ValidateEnum[container.CreateTriggerRequestDestinationConfigHTTPMethod](),
					},
				},
			},
		},
		"sqs": {
			Type:          schema.TypeList,
			MaxItems:      1,
			Description:   "Config for sqs based trigger using scaleway mnq",
			Optional:      true,
			ConflictsWith: []string{"nats", "cron"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"queue": {
						Optional:    true,
						ForceNew:    true,
						Type:        schema.TypeString,
						Description: "Name of the queue",
						Deprecated:  "This field is no longer supported, please use queue_url instead to identify the queue.",
					},
					"project_id": {
						Computed:    true,
						Optional:    true,
						ForceNew:    true,
						Type:        schema.TypeString,
						Description: "Project ID of the project where the mnq sqs exists, defaults to provider project_id",
					},
					"region": {
						Computed:    true,
						Optional:    true,
						ForceNew:    true,
						Type:        schema.TypeString,
						Description: "The region where the SQS queue is hosted, defaults to function's region",
					},
					"endpoint": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Endpoint URL to use to access SQS (e.g., \"https://sqs.mnq.fr-par.scaleway.com\").",
					},
					"queue_url": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The URL of the SQS queue to monitor for messages.",
					},
					"access_key": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "The access key for accessing the SQS queue.",
					},
					"secret_key": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "The secret key for accessing the SQS queue.",
					},
				},
			},
		},
		"nats": {
			Type:          schema.TypeList,
			MaxItems:      1,
			Description:   "Config for nats based trigger using scaleway mnq",
			Optional:      true,
			ConflictsWith: []string{"sqs", "cron"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"account_id": {
						Optional:         true,
						Type:             schema.TypeString,
						Description:      "ID of the mnq nats account",
						DiffSuppressFunc: dsf.Locality,
					},
					"subject": {
						Required:    true,
						Type:        schema.TypeString,
						Description: "NATS subject to subscribe to (e.g., \"my-subject\").",
					},
					"project_id": {
						Computed:    true,
						Optional:    true,
						ForceNew:    true,
						Type:        schema.TypeString,
						Description: "Project ID of the project where the mnq nats exists, defaults to provider project_id",
					},
					"region": {
						Computed:    true,
						Optional:    true,
						ForceNew:    true,
						Type:        schema.TypeString,
						Description: "Region where the mnq nats exists, defaults to function's region",
					},
					"server_urls": {
						Type:        schema.TypeList,
						MaxItems:    5,
						MinItems:    1,
						Required:    true,
						Description: "The URLs of the NATS server (e.g., \"nats://nats.mnq.fr-par.scaleway.com:4222\").",
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"credentials_file_content": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The content of the NATS credentials file that will be used to authenticate with the NATS server and subscribe to the specified subject.",
					},
				},
			},
		},
		"cron": {
			Type:          schema.TypeList,
			MaxItems:      1,
			Description:   "Config for cron based trigger",
			Optional:      true,
			ConflictsWith: []string{"sqs", "nats"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"schedule": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: verify.ValidateCronExpression(),
						Description:      "UNIX cron schedule to run job (e.g., \"* * * * *\").",
					},
					"timezone": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Timezone for the cron schedule, in tz database format (e.g., \"Europe/Paris\").",
					},
					"body": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Body to send to the container when the trigger is invoked.",
					},
					"headers": {
						Type: schema.TypeMap,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional:    true,
						Description: "Additional headers to send to the container when the trigger is invoked.",
					},
				},
			},
		},
		"region": regional.Schema(),
	}
}

func ResourceContainerTriggerCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &container.CreateTriggerRequest{
		Region:      region,
		Name:        types.ExpandOrGenerateString(d.Get("name").(string), "trigger"),
		ContainerID: locality.ExpandID(d.Get("container_id")),
		Description: types.ExpandStringPtr(d.Get("description")),
	}

	if tags, ok := d.GetOk("tags"); ok {
		req.Tags = types.ExpandStrings(tags)
	}

	if destConf, err := expandDestinationConfig(d.Get("destination_config.0")); err != nil {
		return diag.FromErr(err)
	} else {
		req.DestinationConfig = destConf
	}

	if scwSqs, isScwSqs := d.GetOk("sqs.0"); isScwSqs {
		req.SqsConfig = expandContainerTriggerSqsCreationConfig(scwSqs, region)
		// Sensitive values like access_key and secret_key will no longer be accessible after creation,
		// so we need to store them in the state now.
		_ = d.Set("sqs", []any{scwSqs})
	}

	if scwNats, isScwNats := d.GetOk("nats.0"); isScwNats {
		req.NatsConfig = expandContainerTriggerNatsCreationConfig(scwNats)
		// Sensitive values like credentials_file_content will no longer be accessible after creation,
		// so we need to store them in the state now.
		_ = d.Set("nats", []any{scwNats})
	}

	if cron, isCron := d.GetOk("cron.0"); isCron {
		req.CronConfig = expandContainerTriggerCronCreationConfig(cron)
	}

	trigger, err := api.CreateTrigger(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, trigger.ID))

	_, err = waitForTrigger(ctx, api, trigger.ID, region, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceContainerTriggerRead(ctx, d, m)
}

func ResourceContainerTriggerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	trigger, err := waitForTrigger(ctx, api, id, region, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", trigger.Name)
	_ = d.Set("description", trigger.Description)
	_ = d.Set("tags", types.FlattenSliceString(trigger.Tags))
	_ = d.Set("destination_config", flattenDestinationConfig(trigger.DestinationConfig))
	_ = d.Set("sqs", flattenTriggerSqs(d, trigger.SqsConfig))
	_ = d.Set("nats", flattenTriggerNats(d, trigger.NatsConfig))
	_ = d.Set("cron", flattenTriggerCron(trigger.CronConfig))

	diags := diag.Diagnostics(nil)

	if trigger.Status == container.TriggerStatusError {
		errMsg := ""
		if trigger.ErrorMessage != nil {
			errMsg = *trigger.ErrorMessage
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Trigger in error state",
			Detail:   errMsg,
		})
	}

	return diags
}

func ResourceContainerTriggerUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	trigger, err := waitForTrigger(ctx, api, id, region, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	req := &container.UpdateTriggerRequest{
		Region:    region,
		TriggerID: trigger.ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}

	if d.HasChange("tags") {
		req.Tags = types.ExpandUpdatedStringsPtr(d.Get("tags"))
	}

	if d.HasChange("destination_config") {
		req.DestinationConfig, err = updateDestinationConfig(d.Get("destination_config.0"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("sqs") {
		req.SqsConfig = updateSqsConfig(d.Get("sqs.0"))
	}

	if d.HasChange("nats") {
		req.NatsConfig = updateNatsConfig(d.Get("nats.0"))
	}

	if d.HasChange("cron") {
		req.CronConfig = updateCronConfig(d.Get("cron.0"))
	}

	if _, err := api.UpdateTrigger(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return ResourceContainerTriggerRead(ctx, d, m)
}

func ResourceContainerTriggerDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForTrigger(ctx, api, id, region, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteTrigger(&container.DeleteTriggerRequest{
		Region:    region,
		TriggerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForTrigger(ctx, api, id, region, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
