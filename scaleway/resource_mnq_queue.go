package scaleway

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	FIFOQueueNameSuffix = ".fifo"
)

const (
	DefaultQueueKMSDataKeyReusePeriodSeconds  = 300
	DefaultQueueMaximumMessageSize            = 262_144 // 256 KiB.
	DefaultQueueMessageRetentionPeriod        = 345_600 // 4 days.
	DefaultQueueReceiveMessageWaitTimeSeconds = 0
	DefaultQueueVisibilityTimeout             = 30
)

var queueSchema = map[string]*schema.Schema{
	"arn": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"namespace_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Namespace associated to",
	},
	"access_key": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Namespace associated to",
	},
	"secret_key": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Namespace associated to",
	},
	"content_based_deduplication": {
		Type:     schema.TypeBool,
		Default:  false,
		Optional: true,
	},
	"fifo_queue": {
		Type:     schema.TypeBool,
		Default:  false,
		ForceNew: true,
		Optional: true,
	},
	"max_message_size": {
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      DefaultQueueMaximumMessageSize,
		ValidateFunc: validation.IntBetween(1024, 262_144),
	},
	"message_retention_seconds": {
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      DefaultQueueMessageRetentionPeriod,
		ValidateFunc: validation.IntBetween(60, 1_209_600),
	},
	"name": {
		Type:          schema.TypeString,
		Optional:      true,
		Computed:      true,
		ForceNew:      true,
		ConflictsWith: []string{"name_prefix"},
	},
	"name_prefix": {
		Type:          schema.TypeString,
		Optional:      true,
		Computed:      true,
		ForceNew:      true,
		ConflictsWith: []string{"name"},
	},
	"receive_wait_time_seconds": {
		Type:     schema.TypeInt,
		Optional: true,
		Default:  DefaultQueueReceiveMessageWaitTimeSeconds,
	},
	"url": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"visibility_timeout_seconds": {
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      DefaultQueueVisibilityTimeout,
		ValidateFunc: validation.IntBetween(0, 43_200),
	},
}

func resourceScalewayMNQQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScalewayMNQQueueCreate,
		ReadContext:          resourceScalewayMNQQueueRead,
		UpdateContext:        resourceScalewayMNQQueueUpdate,
		DeleteContext:        resourceScalewayMNQQueueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.Sequence(
			resourceQueueCustomizeDiff,
		),
		SchemaVersion: 0,
		Schema:        queueSchema,
	}
}

func resourceScalewayMNQQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	var name string
	fifoQueue := d.Get("fifo_queue").(bool)
	if fifoQueue {
		name = NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string), FIFOQueueNameSuffix)
	} else {
		name = Name(d.Get("name").(string), d.Get("name_prefix").(string))
	}

	input := &sqs.CreateQueueInput{
		QueueName: aws.String(name),
	}

	attributes, err := getQueueAttributeMap().ResourceDataToAPIAttributesCreate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	input.Attributes = aws.StringMap(attributes)

	log.Printf("[DEBUG] Creating SQS Queue: %s", input)
	outputRaw, err := retryWhenAWSErrCodeEquals(ctx, queueCreatedTimeout, func() (interface{}, error) {
		return api.CreateQueueWithContext(ctx, input)
	}, sqs.ErrCodeQueueDeletedRecently)

	// Some partitions may not support tag-on-create
	if input.Tags != nil && errorISOUnsupported(api.PartitionID, err) {
		log.Printf("[WARN] failed creating SQS Queue (%s) with tags: %s. Trying create without tags.", name, err)

		input.Tags = nil
		outputRaw, err = retryWhenAWSErrCodeEquals(ctx, queueCreatedTimeout, func() (interface{}, error) {
			return api.CreateQueueWithContext(ctx, input)
		}, sqs.ErrCodeQueueDeletedRecently)
	}

	if err != nil {
		return diag.Errorf("creating SQS Queue (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*sqs.CreateQueueOutput).QueueUrl))

	err = waitQueueAttributesPropagated(ctx, api, d.Id(), attributes)

	if err != nil {
		return diag.Errorf("waiting for SQS Queue (%s) attributes create: %s", d.Id(), err)
	}

	return resourceScalewayMNQQueueRead(ctx, d, meta)
}

func resourceScalewayMNQQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	outputRaw, err := retryWhenNotFound(ctx, queueReadTimeout, func() (interface{}, error) {
		return FindQueueAttributesByURL(ctx, api, d.Id())
	})

	if !d.IsNewResource() && NotFound(err) {
		log.Printf("[WARN] SQS Queue (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading SQS Queue (%s): %s", d.Id(), err)
	}

	name, err := QueueNameFromURL(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	output := outputRaw.(map[string]string)

	err = getQueueAttributeMap().APIAttributesToResourceData(output, d)

	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", name)
	if d.Get("fifo_queue").(bool) {
		_ = d.Set("name_prefix", NamePrefixFromNameWithSuffix(name, FIFOQueueNameSuffix))
	} else {
		_ = d.Set("name_prefix", NamePrefixFromName(name))
	}
	_ = d.Set("url", d.Id())

	return nil
}

func resourceScalewayMNQQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	attributes, err := getQueueAttributeMap().ResourceDataToAPIAttributesUpdate(d)
	if err != nil {
		return diag.FromErr(err)
	}

	input := &sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(attributes),
		QueueUrl:   aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Updating SQS Queue: %s", input)
	_, err = api.SetQueueAttributesWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating SQS Queue (%s) attributes: %s", d.Id(), err)
	}

	err = waitQueueAttributesPropagated(ctx, api, d.Id(), attributes)

	if err != nil {
		return diag.Errorf("waiting for SQS Queue (%s) attributes update: %s", d.Id(), err)
	}
	return resourceScalewayMNQQueueRead(ctx, d, meta)
}

func resourceScalewayMNQQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting SQS Queue: %s", d.Id())
	_, err = api.DeleteQueueWithContext(ctx, &sqs.DeleteQueueInput{
		QueueUrl: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting SQS Queue (%s): %s", d.Id(), err)
	}

	err = waitQueueDeleted(ctx, api, d.Id())

	if err != nil {
		return diag.Errorf("waiting for SQS Queue (%s) delete: %s", d.Id(), err)
	}

	return nil
}
