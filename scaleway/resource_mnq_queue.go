package scaleway

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	DefaultMNQQueueRetryInterval = 5 * time.Second

	DefaultQueueMaximumMessageSize            = 262_144 // 256 KiB.
	DefaultQueueMessageRetentionPeriod        = 345_600 // 4 days.
	DefaultQueueReceiveMessageWaitTimeSeconds = 0
	DefaultQueueVisibilityTimeout             = 30
)

func resourceScalewayMNQQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceScalewayMNQQueueCreate,
		ReadContext:          resourceScalewayMNQQueueRead,
		UpdateContext:        resourceScalewayMNQQueueUpdate,
		DeleteContext:        resourceScalewayMNQQueueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"namespace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the Namespace associated to",
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
			"fifo_queue": {
				Type:     schema.TypeBool,
				Default:  false,
				ForceNew: true,
				Optional: true,
			},
			"sqs": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The SQS attributes of the queue",
				Elem:        resourceScalewayMNQQueueSQS(),
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the queue",
			},
			"region": regionSchema(),
		},
		CustomizeDiff: customdiff.Sequence(
			resourceMNQQueueCustomizeDiff,
		),
	}
}

func resourceScalewayMNQQueueSQS() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The access key of the SQS queue",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The secret key of the SQS queue",
			},
			"content_based_deduplication": {
				Type:     schema.TypeBool,
				Default:  false,
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
			"receive_wait_time_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  DefaultQueueReceiveMessageWaitTimeSeconds,
			},
			"visibility_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      DefaultQueueVisibilityTimeout,
				ValidateFunc: validation.IntBetween(0, 43_200),
			},
		},
	}
}

func resourceScalewayMNQQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := newMNQAPI(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceRegion, namespaceID, err := parseRegionalID(d.Get("namespace_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if region != namespaceRegion {
		return diag.Errorf("region of the queue (%s) and the namespace (%s) must match", region, namespaceRegion)
	}

	namespace, err := api.GetNamespace(&mnq.GetNamespaceRequest{
		Region:      namespaceRegion,
		NamespaceID: namespaceID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	switch namespace.Protocol {
	case mnq.NamespaceProtocolSqsSns:
		return resourceScalewayMNQQueueCreateSQS(ctx, d, meta)
	// case mnq.NamespaceProtocolNats:
	// 	return resourceScalewayMNQQueueCreateNATS(ctx, d, meta)
	default:
		return diag.Errorf("unknown protocol %s", namespace.Protocol)
	}
}

func resourceScalewayMNQQueueCreateSQS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	isFifo := d.Get("fifo_queue").(bool)

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = id.PrefixedUniqueId(v.(string))

		if isFifo {
			name += SQSFIFOQueueNameSuffix
		}
	}

	var attributes map[string]*string
	if v, ok := d.GetOk("sqs"); ok {
		attributes = SQSResourceToAttributes(resourceScalewayMNQQueueSQS().Schema, v.([]interface{})[0].(map[string]interface{}))
	} else {
		attributes = make(map[string]*string)
	}
	attributes[sqs.QueueAttributeNameFifoQueue] = aws.String(strconv.FormatBool(isFifo))

	input := &sqs.CreateQueueInput{
		QueueName:  aws.String(name),
		Attributes: attributes,
	}

	_, err = retryWhenAWSErrCodeEquals(ctx, []string{sqs.ErrCodeQueueDeletedRecently}, &RetryWhenConfig[*sqs.CreateQueueOutput]{
		Timeout:  d.Timeout(schema.TimeoutCreate),
		Interval: DefaultMNQQueueRetryInterval,
		Function: func() (*sqs.CreateQueueOutput, error) {
			return api.CreateQueueWithContext(ctx, input)
		},
	})
	if err != nil {
		return diag.Errorf("failed to create SQS Queue: %s", err)
	}

	namespaceRegion, namespaceID, err := parseRegionalID(d.Get("namespace_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(composeMNQID(namespaceRegion, namespaceID, *input.QueueName))

	return resourceScalewayMNQQueueReadSQS(ctx, d, meta)
}

func resourceScalewayMNQQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	namespace, err := getMNQNamespaceFromComposedID(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	switch namespace.Protocol {
	case mnq.NamespaceProtocolSqsSns:
		return resourceScalewayMNQQueueReadSQS(ctx, d, meta)
	// case mnq.NamespaceProtocolNats:
	// 	return resourceScalewayMNQQueueReadNATS(ctx, d, meta)
	default:
		return diag.Errorf("unknown protocol %s", namespace.Protocol)
	}
}

func resourceScalewayMNQQueueReadSQS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, queueName, err := decomposeMNQID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	queue, err := retryWhenAWSErrCodeEquals(ctx, []string{sqs.ErrCodeQueueDoesNotExist}, &RetryWhenConfig[*sqs.GetQueueUrlOutput]{
		Timeout:  d.Timeout(schema.TimeoutRead),
		Interval: DefaultMNQQueueRetryInterval,
		Function: func() (*sqs.GetQueueUrlOutput, error) {
			return api.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			})
		},
	})
	if err != nil {
		return diag.Errorf("failed to get the SQS Queue URL: %s", err)
	}

	if err != nil {
		return diag.Errorf("reading SQS Queue (%s): %s", d.Id(), err)
	}

	queueAttributes, err := api.GetQueueAttributesWithContext(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: queue.QueueUrl,
		AttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameFifoQueue),
			aws.String(sqs.QueueAttributeNameContentBasedDeduplication),
			aws.String(sqs.QueueAttributeNameMaximumMessageSize),
			aws.String(sqs.QueueAttributeNameMessageRetentionPeriod),
			aws.String(sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds),
			aws.String(sqs.QueueAttributeNameVisibilityTimeout),
		},
	})
	if err != nil {
		return diag.Errorf("failed to get the SQS Queue attributes: %s", err)
	}

	_ = d.Set("name", queueName)
	_ = d.Set("name_prefix", d.Get("name_prefix").(string))

	_, isFifo := queueAttributes.Attributes[sqs.QueueAttributeNameFifoQueue]
	_ = d.Set("fifo_queue", isFifo)

	sqsElements := SQSAttributesToResource(resourceScalewayMNQQueueSQS().Schema, queueAttributes.Attributes)
	sqsElements["access_key"] = d.Get("sqs.0.access_key").(string)
	sqsElements["secret_key"] = d.Get("sqs.0.secret_key").(string)
	_ = d.Set("sqs", []map[string]interface{}{sqsElements})

	return nil
}

func resourceScalewayMNQQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	namespace, err := getMNQNamespaceFromComposedID(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	switch namespace.Protocol {
	case mnq.NamespaceProtocolSqsSns:
		return resourceScalewayMNQQueueUpdateSQS(ctx, d, meta)
	// case mnq.NamespaceProtocolNats:
	// 	return resourceScalewayMNQQueueUpdateNATS(ctx, d, meta)
	default:
		return diag.Errorf("unknown protocol %s", namespace.Protocol)
	}
}

func resourceScalewayMNQQueueUpdateSQS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, queueName, err := decomposeMNQID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	queue, err := retryWhenAWSErrCodeEquals(ctx, []string{sqs.ErrCodeQueueDoesNotExist}, &RetryWhenConfig[*sqs.GetQueueUrlOutput]{
		Timeout:  d.Timeout(schema.TimeoutUpdate),
		Interval: DefaultMNQQueueRetryInterval,
		Function: func() (*sqs.GetQueueUrlOutput, error) {
			return api.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			})
		},
	})
	if err != nil {
		return diag.Errorf("failed to get the SQS Queue URL: %s", err)
	}

	var attributes map[string]*string
	if v, ok := d.GetOk("sqs"); ok {
		attributes = SQSResourceToAttributes(resourceScalewayMNQQueueSQS().Schema, v.([]interface{})[0].(map[string]interface{}))
	} else {
		attributes = make(map[string]*string)
	}
	attributes[sqs.QueueAttributeNameFifoQueue] = aws.String(strconv.FormatBool(d.Get("fifo_queue").(bool)))

	_, err = api.SetQueueAttributesWithContext(ctx, &sqs.SetQueueAttributesInput{
		QueueUrl:   queue.QueueUrl,
		Attributes: attributes,
	})
	if err != nil {
		return diag.Errorf("failed to update SQS Queue attributes: %s", err)
	}

	return resourceScalewayMNQQueueReadSQS(ctx, d, meta)
}

func resourceScalewayMNQQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	namespace, err := getMNQNamespaceFromComposedID(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	switch namespace.Protocol {
	case mnq.NamespaceProtocolSqsSns:
		return resourceScalewayMNQQueueDeleteSQS(ctx, d, meta)
	// case mnq.NamespaceProtocolNats:
	// 	return resourceScalewayMNQQueueDeleteNATS(ctx, d, meta)
	default:
		return diag.Errorf("unknown protocol %s", namespace.Protocol)
	}
}

func resourceScalewayMNQQueueDeleteSQS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, queueName, err := decomposeMNQID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	queue, err := api.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
			return nil
		}

		return diag.Errorf("failed to get the SQS Queue URL: %s", err)
	}

	_, err = api.DeleteQueueWithContext(ctx, &sqs.DeleteQueueInput{
		QueueUrl: queue.QueueUrl,
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
			return nil
		}

		return diag.Errorf("failed to delete SQS Queue (%s): %s", d.Id(), err)
	}

	_, _ = retryWhenAWSErrCodeNotEquals(ctx, []string{sqs.ErrCodeQueueDoesNotExist}, &RetryWhenConfig[*sqs.GetQueueUrlOutput]{
		Timeout:  d.Timeout(schema.TimeoutCreate),
		Interval: DefaultMNQQueueRetryInterval,
		Function: func() (*sqs.GetQueueUrlOutput, error) {
			return api.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			})
		},
	})

	return nil
}
