package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nats-io/nats.go"
)

func resourceScalewayMNQNatsQueue() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayMNQNatsQueueCreate,
		ReadContext:   resourceScalewayMNQNatsQueueRead,
		UpdateContext: resourceScalewayMNQNatsQueueUpdate,
		DeleteContext: resourceScalewayMNQNatsQueueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "The Nats queue name",
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "MNQ Nats endpoint",
			},
			"credentials": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "MNQ Nats credentials",
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayMNQNatsQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, region, err := NATSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	stream, err := client.AddStream(&nats.StreamConfig{
		Name: expandOrGenerateString(d.Get("name"), "nats-queue"),
		// MaxAge: time.Duration(maxAge) * time.Second,
		// MaxBytes: int64(maxSize),
		Retention: nats.WorkQueuePolicy, // TODO
	}, nats.Context(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(composeMNQQueueID(region, stream.Config.Name))

	return resourceScalewayMNQNatsQueueRead(ctx, d, meta)
}

func resourceScalewayMNQNatsQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, region, err := NATSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	region, queueName, err := decomposeMNQQueueID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	stream, err := client.StreamInfo(queueName, nats.Context(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("name", stream.Config.Name)
	_ = d.Set("region", region)

	return nil
}

func resourceScalewayMNQNatsQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _, err := NATSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, queueName, err := decomposeMNQQueueID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.UpdateStream(&nats.StreamConfig{
		Name:      queueName,
		Retention: nats.WorkQueuePolicy, // TODO
	}, nats.Context(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScalewayMNQNatsQueueRead(ctx, d, meta)
}

func resourceScalewayMNQNatsQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, _, err := NATSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, queueName, err := decomposeMNQQueueID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.DeleteStream(queueName, nats.Context(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
