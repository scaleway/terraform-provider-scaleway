package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	container "github.com/scaleway/scaleway-sdk-go/api/container/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayContainerTrigger() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayContainerTriggerCreate,
		ReadContext:   resourceScalewayContainerTriggerRead,
		UpdateContext: resourceScalewayContainerTriggerUpdate,
		DeleteContext: resourceScalewayContainerTriggerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultFunctionTimeout),
			Read:    schema.DefaultTimeout(defaultFunctionTimeout),
			Update:  schema.DefaultTimeout(defaultFunctionTimeout),
			Delete:  schema.DefaultTimeout(defaultFunctionTimeout),
			Create:  schema.DefaultTimeout(defaultFunctionTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"container_id": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The ID of the container to create a trigger for",
				ValidateFunc: validationUUIDorUUIDWithLocality(),
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
			"sqs": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Description: "Config for sqs based trigger using scaleway mnq",
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace_id": {
							Required:         true,
							Type:             schema.TypeString,
							Description:      "ID of the mnq namespace",
							DiffSuppressFunc: diffSuppressFuncLocality,
						},
						"queue": {
							Required:    true,
							Type:        schema.TypeString,
							Description: "Name of the queue",
						},
						"project_id": {
							Computed:    true,
							Optional:    true,
							Type:        schema.TypeString,
							Description: "Project ID of the project where the mnq sqs exists, defaults to provider project_id",
						},
						"region": {
							Computed:    true,
							Optional:    true,
							Type:        schema.TypeString,
							Description: "Region where the mnq sqs exists, defaults to function's region",
						},
					},
				},
			},
			"region": regionSchema(),
		},
		CustomizeDiff: customizeDiffLocalityCheck("container_id"),
	}
}

func resourceScalewayContainerTriggerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := containerAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &container.CreateTriggerRequest{
		Region:      region,
		Name:        expandOrGenerateString(d.Get("name").(string), "trigger"),
		ContainerID: expandID(d.Get("container_id")),
		Description: expandStringPtr(d.Get("description")),
	}

	if scwSqs, isScwSqs := d.GetOk("sqs.0"); isScwSqs {
		err := completeContainerTriggerMnqSqsCreationConfig(scwSqs, d, meta, region)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to complete sqs config: %w", err))
		}

		_ = d.Set("sqs", []any{scwSqs})
		req.ScwSqsConfig = expandContainerTriggerMnqSqsCreationConfig(scwSqs)
	}

	trigger, err := api.CreateTrigger(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newRegionalIDString(region, trigger.ID))

	_, err = waitForContainerTrigger(ctx, api, region, trigger.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayContainerTriggerRead(ctx, d, meta)
}

func resourceScalewayContainerTriggerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	trigger, err := waitForContainerTrigger(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", trigger.Name)
	_ = d.Set("description", trigger.Description)

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

func resourceScalewayContainerTriggerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	trigger, err := waitForContainerTrigger(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if is404Error(err) {
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
		req.Name = expandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("description") {
		req.Description = expandUpdatedStringPtr(d.Get("description"))
	}

	if _, err := api.UpdateTrigger(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayContainerTriggerRead(ctx, d, meta)
}

func resourceScalewayContainerTriggerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, id, err := containerAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForContainerTrigger(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
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

	_, err = waitForContainerTrigger(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
