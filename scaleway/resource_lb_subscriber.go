package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayLbSubscriber() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayLbSubscriberCreate,
		ReadContext:   resourceScalewayLbSubscriberRead,
		UpdateContext: resourceScalewayLbSubscriberUpdate,
		DeleteContext: resourceScalewayLbSubscriberDelete,
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Subscriber name.",
				Optional:    true,
				Computed:    true,
			},
			"email_config": {
				ConflictsWith: []string{"webhook_config"},
				MaxItems:      1,
				Description:   "Email address configuration.",
				Type:          schema.TypeList,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Email receiving the alert.",
						},
					},
				},
			},
			"webhook_config": {
				ConflictsWith: []string{"email_config"},
				MaxItems:      1,
				Description:   "WebHook URI configuration.",
				Type:          schema.TypeList,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uri": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "URI receiving the POST request.",
						},
					},
				},
			},
			"project_id": projectIDSchema(),
		},
	}
}
func resourceScalewayLbSubscriberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &lbSDK.ZonedAPICreateSubscriberRequest{
		Zone:          zone,
		ProjectID:     expandStringPtr(d.Get("project_id")),
		Name:          expandOrGenerateString(d.Get("name"), "lb-subscriber"),
		EmailConfig:   expandLbSubscriberEmailConfig(d.Get("email_config")),
		WebhookConfig: expandLbSubscriberWebhookConfig(d.Get("webhook_config")),
	}

	subscriber, err := lbAPI.CreateSubscriber(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(newZonedIDString(zone, subscriber.ID))

	return resourceScalewayLbSubscriberRead(ctx, d, meta)
}

func resourceScalewayLbSubscriberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	subscriber, err := lbAPI.GetSubscriber(&lbSDK.ZonedAPIGetSubscriberRequest{
		Zone:         zone,
		SubscriberID: ID,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", subscriber.Name)
	_ = d.Set("email_config", flattenLbSubscriberURI(subscriber.EmailConfig))
	_ = d.Set("webhook_config", flattenLbSubscriberWebhook(subscriber.WebhookConfig))

	return nil
}

func resourceScalewayLbSubscriberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &lbSDK.ZonedAPIUpdateSubscriberRequest{
		Zone:          zone,
		SubscriberID:  ID,
		Name:          d.Get("name").(string),
		EmailConfig:   expandLbSubscriberEmailConfig(d.Get("email_config")),
		WebhookConfig: expandLbSubscriberWebhookConfig(d.Get("webhook_config")),
	}

	_, err = lbAPI.UpdateSubscriber(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayLbSubscriberRead(ctx, d, meta)
}

func resourceScalewayLbSubscriberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, ID, err := lbAPIWithZoneAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = lbAPI.DeleteSubscriber(&lbSDK.ZonedAPIDeleteSubscriberRequest{
		Zone:         zone,
		SubscriberID: ID,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
