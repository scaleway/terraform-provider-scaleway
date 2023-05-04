package scaleway

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nats-io/nats.go"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultMNQQueueRetryInterval = 5 * time.Second

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
				Optional: true,
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
			"sqs": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The SQS attributes of the queue",
				Elem:        resourceScalewayMNQQueueSQS(),
			},
			"nats": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The NATS attributes of the queue",
				Elem:        resourceScalewayMNQQueueNATS(),
			},
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
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the queue",
			},
		},
	}
}

func resourceScalewayMNQQueueNATS() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"credentials": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Line jump separated key and seed",
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
	case mnq.NamespaceProtocolNats:
		return resourceScalewayMNQQueueCreateNATS(ctx, d, meta)
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
	name := resourceMNQQueueName(d.Get("name"), d.Get("name_prefix"), true, isFifo)

	attributes, err := sqsResourceDataToAttributes(d, resourceScalewayMNQQueue().Schema)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] SQS Queue attributes CREATE: %+v", attributes)

	input := &sqs.CreateQueueInput{
		QueueName:  aws.String(name),
		Attributes: attributes,
	}

	_, err = retryWhenAWSErrCodeEquals(ctx, []string{sqs.ErrCodeQueueDeletedRecently}, &RetryWhenConfig[*sqs.CreateQueueOutput]{
		Timeout:  d.Timeout(schema.TimeoutCreate),
		Interval: defaultMNQQueueRetryInterval,
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
	d.SetId(composeMNQQueueID(namespaceRegion, namespaceID, *input.QueueName))

	return resourceScalewayMNQQueueReadSQS(ctx, d, meta)
}

func resourceScalewayMNQQueueCreateNATS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	js, _, err := NATSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	name := resourceMNQQueueName(d.Get("name"), d.Get("name_prefix"), false, false)
	maxAge := d.Get("message_max_age").(int)
	maxSize := d.Get("message_max_size").(int)

	var retention nats.RetentionPolicy
	if d.Get("fifo_queue").(bool) {
		retention = nats.InterestPolicy
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:      name,
		MaxAge:    time.Duration(maxAge) * time.Second,
		MaxBytes:  int64(maxSize),
		Retention: retention,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceRegion, namespaceID, err := parseRegionalID(d.Get("namespace_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(composeMNQQueueID(namespaceRegion, namespaceID, name))

	return resourceScalewayMNQQueueReadNATS(ctx, d, meta)
}

func resourceScalewayMNQQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	namespace, err := getMNQNamespaceFromComposedQueueID(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	switch namespace.Protocol {
	case mnq.NamespaceProtocolSqsSns:
		return resourceScalewayMNQQueueReadSQS(ctx, d, meta)
	case mnq.NamespaceProtocolNats:
		return resourceScalewayMNQQueueReadNATS(ctx, d, meta)
	default:
		return diag.Errorf("unknown protocol %s", namespace.Protocol)
	}
}

func resourceScalewayMNQQueueReadSQS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceRegion, namespaceID, queueName, err := decomposeMNQQueueID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	queue, err := retryWhenAWSErrCodeEquals(ctx, []string{sqs.ErrCodeQueueDoesNotExist}, &RetryWhenConfig[*sqs.GetQueueUrlOutput]{
		Timeout:  d.Timeout(schema.TimeoutRead),
		Interval: defaultMNQQueueRetryInterval,
		Function: func() (*sqs.GetQueueUrlOutput, error) {
			return api.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			})
		},
	})
	if err != nil {
		return diag.Errorf("[READ] failed to get the SQS Queue URL: %s", err)
	}

	queueAttributes, err := api.GetQueueAttributesWithContext(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       queue.QueueUrl,
		AttributeNames: getSQSAttributeNames(),
	})
	if err != nil {
		return diag.Errorf("failed to get the SQS Queue attributes: %s", err)
	}

	_ = d.Set("namespace_id", newRegionalIDString(namespaceRegion, namespaceID))
	_ = d.Set("name", queueName)

	values, err := sqsAttributesToResourceData(queueAttributes.Attributes, resourceScalewayMNQQueue().Schema)
	if err != nil {
		return diag.Errorf("failed to convert SQS Queue attributes to resource data: %s", err)
	}

	sqs := values["sqs"].([]interface{})[0].(map[string]interface{})
	sqs["url"] = flattenStringPtr(queue.QueueUrl)
	sqs["access_key"] = d.Get("sqs.0.access_key").(string)
	sqs["secret_key"] = d.Get("sqs.0.secret_key").(string)

	if _, ok := sqs["visibility_timeout_seconds"]; !ok {
		return diag.Errorf("failed to get the SQS Queue visibility timeout: %+v", *queueAttributes.Attributes["VisibilityTimeout"])
	}

	log.Printf("[DEBUG] SQS Queue attributes READ: %+v", values)

	for k, v := range values {
		_ = d.Set(k, v) // lintignore: R001
	}

	return nil
}

func resourceScalewayMNQQueueReadNATS(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	js, _, err := NATSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceRegion, namespaceID, queueName, err := decomposeMNQQueueID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	stream, err := js.StreamInfo(queueName)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("namespace_id", newRegionalIDString(namespaceRegion, namespaceID))
	_ = d.Set("name", queueName)
	_ = d.Set("message_max_age", int(stream.Config.MaxAge.Seconds()))
	_ = d.Set("message_max_size", int(stream.Config.MaxBytes))

	return nil
}

func resourceScalewayMNQQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	namespace, err := getMNQNamespaceFromComposedQueueID(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	switch namespace.Protocol {
	case mnq.NamespaceProtocolSqsSns:
		return resourceScalewayMNQQueueUpdateSQS(ctx, d, meta)
	case mnq.NamespaceProtocolNats:
		return resourceScalewayMNQQueueUpdateNATS(ctx, d, meta)
	default:
		return diag.Errorf("unknown protocol %s", namespace.Protocol)
	}
}

func resourceScalewayMNQQueueUpdateSQS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, queueName, err := decomposeMNQQueueID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	queue, err := retryWhenAWSErrCodeEquals(ctx, []string{sqs.ErrCodeQueueDoesNotExist}, &RetryWhenConfig[*sqs.GetQueueUrlOutput]{
		Timeout:  d.Timeout(schema.TimeoutUpdate),
		Interval: defaultMNQQueueRetryInterval,
		Function: func() (*sqs.GetQueueUrlOutput, error) {
			return api.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			})
		},
	})
	if err != nil {
		return diag.Errorf("[UPDATE] failed to get the SQS Queue URL: %s", err)
	}

	attributes, err := sqsResourceDataToAttributes(d, resourceScalewayMNQQueue().Schema)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.SetQueueAttributesWithContext(ctx, &sqs.SetQueueAttributesInput{
		QueueUrl:   queue.QueueUrl,
		Attributes: attributes,
	})
	if err != nil {
		return diag.Errorf("failed to update SQS Queue attributes: %s", err)
	}

	return resourceScalewayMNQQueueReadSQS(ctx, d, meta)
}

func resourceScalewayMNQQueueUpdateNATS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	js, _, err := NATSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, queueName, err := decomposeMNQQueueID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	stream, err := js.StreamInfo(queueName)
	if err != nil {
		return diag.FromErr(err)
	}

	maxAge := d.Get("message_max_age").(int)
	maxSize := d.Get("message_max_size").(int)

	var retention nats.RetentionPolicy
	if d.Get("fifo_queue").(bool) {
		retention = nats.InterestPolicy
	}

	_, err = js.UpdateStream(&nats.StreamConfig{
		Name:      queueName,
		Subjects:  stream.Config.Subjects,
		MaxAge:    time.Duration(maxAge) * time.Second,
		MaxBytes:  int64(maxSize),
		Retention: retention,
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceScalewayMNQQueueReadNATS(ctx, d, meta)
}

func resourceScalewayMNQQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	namespace, err := getMNQNamespaceFromComposedQueueID(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	switch namespace.Protocol {
	case mnq.NamespaceProtocolSqsSns:
		return resourceScalewayMNQQueueDeleteSQS(ctx, d, meta)
	case mnq.NamespaceProtocolNats:
		return resourceScalewayMNQQueueDeleteNATS(ctx, d, meta)
	default:
		return diag.Errorf("unknown protocol %s", namespace.Protocol)
	}
}

func resourceScalewayMNQQueueDeleteSQS(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, _, err := SQSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, queueName, err := decomposeMNQQueueID(d.Id())
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

		return diag.Errorf("[DELETE] failed to get the SQS Queue URL: %s", err)
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
		Interval: defaultMNQQueueRetryInterval,
		Function: func() (*sqs.GetQueueUrlOutput, error) {
			return api.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
				QueueName: aws.String(queueName),
			})
		},
	})

	return nil
}

func resourceScalewayMNQQueueDeleteNATS(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	js, _, err := NATSClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, _, queueName, err := decomposeMNQQueueID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = js.DeleteStream(queueName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
