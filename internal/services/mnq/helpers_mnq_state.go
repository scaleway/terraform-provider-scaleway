package mnq

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func resourceMNQSQSQueueResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Description:   "The name of the queue. Conflicts with name_prefix.",
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Description:   "Creates a unique name beginning with the specified prefix. Conflicts with name.",
				ConflictsWith: []string{"name"},
			},
			"sqs_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "http://sqs-sns.mnq.{region}.scw.cloud",
				Description: "The sqs endpoint",
			},
			"access_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "SQS access key",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "SQS secret key",
			},
			"fifo_queue": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the queue is a FIFO queue. If true, the queue name must end with .fifo",
			},
			"content_based_deduplication": {
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Description: "Specifies whether to enable content-based deduplication. Allows omitting the deduplication ID",
			},
			"receive_wait_time_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     DefaultQueueReceiveMessageWaitTimeSeconds,
				Description: "The number of seconds to wait for a message to arrive in the queue before returning.",
			},
			"visibility_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      DefaultQueueVisibilityTimeout,
				ValidateFunc: validation.IntBetween(0, 43_200),
				Description:  "The number of seconds a message is hidden from other consumers.",
			},
			"message_max_age": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      DefaultQueueMessageRetentionPeriod,
				ValidateFunc: validation.IntBetween(60, 1_209_600),
				Description:  "The number of seconds the queue retains a message.",
			},
			"message_max_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      DefaultQueueMaximumMessageSize,
				ValidateFunc: validation.IntBetween(1024, 262_144),
				Description:  "The maximum size of a message. Should be in bytes.",
			},
			"region":     regional.Schema(),
			"project_id": account.ProjectIDSchema(),

			// Computed

			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the queue",
			},
		},
	}
}

func resourceMNQSQSQueueStateUpgradeV0(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	rawState["sqs_endpoint"] = rawState["endpoint"]

	return rawState, nil
}
