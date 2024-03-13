package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	function "github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func resourceScalewayFunctionTrigger() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayFunctionTriggerCreate,
		ReadContext:   resourceScalewayFunctionTriggerRead,
		UpdateContext: resourceScalewayFunctionTriggerUpdate,
		DeleteContext: resourceScalewayFunctionTriggerDelete,
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
			"function_id": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The ID of the function to create a trigger for",
				ValidateFunc: verify.IsUUIDorUUIDWithLocality(),
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
				Type:          schema.TypeList,
				MaxItems:      1,
				Description:   "Config for sqs based trigger using scaleway mnq",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"nats"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace_id": {
							Optional:         true,
							ForceNew:         true,
							Type:             schema.TypeString,
							Description:      "ID of the mnq namespace",
							DiffSuppressFunc: diffSuppressFuncLocality,
						},
						"queue": {
							Required:    true,
							ForceNew:    true,
							Type:        schema.TypeString,
							Description: "Name of the queue",
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
							Description: "Region where the mnq sqs exists, defaults to function's region",
						},
					},
				},
			},
			"nats": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Description:   "Config for nats based trigger using scaleway mnq",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"sqs"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Optional:         true,
							ForceNew:         true,
							Type:             schema.TypeString,
							Description:      "ID of the mnq nats account",
							DiffSuppressFunc: diffSuppressFuncLocality,
						},
						"subject": {
							Required:    true,
							ForceNew:    true,
							Type:        schema.TypeString,
							Description: "Subject to listen to",
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
							Description: "Region where the mnq sqs exists, defaults to function's region",
						},
					},
				},
			},
			"region": regional.Schema(),
		},
		CustomizeDiff: CustomizeDiffLocalityCheck("function_id"),
	}
}

func resourceScalewayFunctionTriggerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := functionAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &function.CreateTriggerRequest{
		Region:      region,
		Name:        types.ExpandOrGenerateString(d.Get("name").(string), "trigger"),
		FunctionID:  locality.ExpandID(d.Get("function_id")),
		Description: types.ExpandStringPtr(d.Get("description")),
	}

	if scwSqs, isScwSqs := d.GetOk("sqs.0"); isScwSqs {
		err := completeFunctionTriggerMnqCreationConfig(scwSqs, d, m, region)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to complete sqs config: %w", err))
		}

		_ = d.Set("sqs", []any{scwSqs})
		req.ScwSqsConfig = expandFunctionTriggerMnqSqsCreationConfig(scwSqs)
	}

	if scwNats, isScwNats := d.GetOk("nats.0"); isScwNats {
		err := completeFunctionTriggerMnqCreationConfig(scwNats, d, m, region)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to complete nats config: %w", err))
		}

		_ = d.Set("nats", []any{scwNats})
		req.ScwNatsConfig = expandFunctionTriggerMnqNatsCreationConfig(scwNats)
	}

	trigger, err := api.CreateTrigger(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, trigger.ID))

	_, err = waitForFunctionTrigger(ctx, api, region, trigger.ID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayFunctionTriggerRead(ctx, d, m)
}

func resourceScalewayFunctionTriggerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	trigger, err := waitForFunctionTrigger(ctx, api, region, id, d.Timeout(schema.TimeoutRead))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", trigger.Name)
	_ = d.Set("description", trigger.Description)

	diags := diag.Diagnostics(nil)

	if trigger.Status == function.TriggerStatusError {
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

func resourceScalewayFunctionTriggerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	trigger, err := waitForFunctionTrigger(ctx, api, region, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	req := &function.UpdateTriggerRequest{
		Region:    region,
		TriggerID: trigger.ID,
	}

	if d.HasChange("name") {
		req.Name = types.ExpandUpdatedStringPtr(d.Get("name"))
	}

	if d.HasChange("description") {
		req.Description = types.ExpandUpdatedStringPtr(d.Get("description"))
	}

	if _, err := api.UpdateTrigger(req, scw.WithContext(ctx)); err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayFunctionTriggerRead(ctx, d, m)
}

func resourceScalewayFunctionTriggerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := functionAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFunctionTrigger(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.DeleteTrigger(&function.DeleteTriggerRequest{
		Region:    region,
		TriggerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForFunctionTrigger(ctx, api, region, id, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
