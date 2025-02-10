package tem

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceWebhookCreate,
		ReadContext:   ResourceWebhookRead,
		UpdateContext: ResourceWebhookUpdate,
		DeleteContext: ResourceWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The domain id",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(3, 127),
			},
			"event_types": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringNotInSlice([]string{"unknown_type"}, false),
				},
				Description: "List of event types",
				MinItems:    1,
			},
			"sns_arn": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "SNS ARN",
				ValidateFunc: validation.StringLenBetween(3, 127),
			},
			"organization_id": account.OrganizationIDSchema(),
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp",
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := temAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	eventTypes := expandWebhookEventTypes(d.Get("event_types").([]interface{}))

	webhook, err := api.CreateWebhook(&tem.CreateWebhookRequest{
		Region:     region,
		ProjectID:  d.Get("project_id").(string),
		Name:       d.Get("name").(string),
		DomainID:   extractAfterSlash(d.Get("domain_id").(string)),
		SnsArn:     d.Get("sns_arn").(string),
		EventTypes: eventTypes,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, webhook.ID))

	return ResourceWebhookRead(ctx, d, m)
}

func ResourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	webhook, err := api.GetWebhook(&tem.GetWebhookRequest{
		Region:    region,
		WebhookID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("name", webhook.Name)
	_ = d.Set("domain_id", regional.NewIDString(region, webhook.DomainID))
	_ = d.Set("organization_id", webhook.OrganizationID)
	_ = d.Set("project_id", webhook.ProjectID)
	_ = d.Set("event_types", webhook.EventTypes)
	_ = d.Set("sns_arn", webhook.SnsArn)
	_ = d.Set("created_at", types.FlattenTime(webhook.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(webhook.UpdatedAt))

	return nil
}

func ResourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &tem.UpdateWebhookRequest{
		Region:    region,
		WebhookID: id,
	}

	if d.HasChange("name") {
		req.Name = scw.StringPtr(d.Get("name").(string))
	}

	if d.HasChange("event_types") {
		rawEventTypes := d.Get("event_types").([]interface{})
		eventTypes := make([]tem.WebhookEventType, len(rawEventTypes))

		for i, raw := range rawEventTypes {
			eventTypes[i] = tem.WebhookEventType(raw.(string))
		}

		req.EventTypes = eventTypes
	}

	if d.HasChange("sns_arn") {
		req.SnsArn = scw.StringPtr(d.Get("sns_arn").(string))
	}

	_, err = api.UpdateWebhook(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return ResourceWebhookRead(ctx, d, m)
}

func ResourceWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, id, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = api.DeleteWebhook(&tem.DeleteWebhookRequest{
		WebhookID: id,
		Region:    region,
	}, scw.WithContext(ctx))

	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
