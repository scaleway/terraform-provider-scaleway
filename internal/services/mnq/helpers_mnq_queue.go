package mnq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	awstype "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	smithylogging "github.com/aws/smithy-go/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	natsjwt "github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultMNQQueueTimeout       = 5 * time.Minute
	defaultMNQQueueRetryInterval = 5 * time.Second

	DefaultQueueMaximumMessageSize            = 262_144 // 256 KiB.
	DefaultQueueMessageRetentionPeriod        = 345_600 // 4 days.
	DefaultQueueReceiveMessageWaitTimeSeconds = 0
	DefaultQueueVisibilityTimeout             = 30
)

type httpDebugLogger struct{}

func (h *httpDebugLogger) Logf(classification smithylogging.Classification, format string, v ...any) {
	if classification == smithylogging.Debug {
		log.Printf("[HTTP DEBUG] %s", fmt.Sprintf(format, v...))
	}
}

func SQSClientWithRegion(ctx context.Context, d *schema.ResourceData, m any) (*sqs.Client, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	endpoint := d.Get("sqs_endpoint").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)

	sqsClient, err := NewSQSClient(ctx, meta.ExtractHTTPClient(m), region.String(), endpoint, accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return sqsClient, region, err
}

func NewSQSClient(ctx context.Context, httpClient *http.Client, region string, endpoint string, accessKey string, secretKey string) (*sqs.Client, error) {
	customEndpoint := strings.ReplaceAll(endpoint, "{region}", region)
	customConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithBaseEndpoint(customEndpoint),
		config.WithHTTPClient(httpClient),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)

	if logging.IsDebugOrHigher() {
		customConfig.Logger = &httpDebugLogger{}
	}

	if err != nil {
		return nil, err
	}

	return sqs.NewFromConfig(customConfig), nil
}

func NATSClientWithRegion( //nolint:ireturn,nolintlint
	d *schema.ResourceData,
	m any,
) (nats.JetStreamContext, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	endpoint := d.Get("endpoint").(string)
	creds := d.Get("credentials").(string)

	js, err := newNATSJetStreamClient(region.String(), endpoint, creds)
	if err != nil {
		return nil, "", err
	}

	return js, region, err
}

func newNATSJetStreamClient( //nolint:ireturn,nolintlint
	region string,
	endpoint string,
	credentials string,
) (nats.JetStreamContext, error) {
	jwt, seed, err := splitNATSJWTAndSeed(credentials)
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(strings.ReplaceAll(endpoint, "{region}", region), nats.UserJWTAndSeed(jwt, seed))
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	return js, nil
}

func splitNATSJWTAndSeed(credentials string) (string, string, error) {
	jwt, err := natsjwt.ParseDecoratedJWT([]byte(credentials))
	if err != nil {
		return "", "", err
	}

	nkey, err := natsjwt.ParseDecoratedUserNKey([]byte(credentials))
	if err != nil {
		return "", "", err
	}

	seed, err := nkey.Seed()
	if err != nil {
		return "", "", err
	}

	return jwt, string(seed), nil
}

const SQSFIFOQueueNameSuffix = ".fifo"

var SQSAttributesToResourceMap = map[string]string{
	string(awstype.QueueAttributeNameMaximumMessageSize):            "message_max_size",
	string(awstype.QueueAttributeNameMessageRetentionPeriod):        "message_max_age",
	string(awstype.QueueAttributeNameFifoQueue):                     "fifo_queue",
	string(awstype.QueueAttributeNameContentBasedDeduplication):     "content_based_deduplication",
	string(awstype.QueueAttributeNameReceiveMessageWaitTimeSeconds): "receive_wait_time_seconds",
	string(awstype.QueueAttributeNameVisibilityTimeout):             "visibility_timeout_seconds",
	string(awstype.QueueAttributeNameRedrivePolicy):                 "dead_letter_queue",
	string(awstype.QueueAttributeNameQueueArn):                      "arn",
}

// Returns all managed SQS attribute names
func getSQSAttributeNames() []awstype.QueueAttributeName {
	attributeNames := make([]awstype.QueueAttributeName, 0, len(SQSAttributesToResourceMap))

	for attribute := range SQSAttributesToResourceMap {
		attributeNames = append(attributeNames, awstype.QueueAttributeName(attribute))
	}

	return attributeNames
}

func resourceMNQQueueName(name any, prefix any, isSQS bool, isSQSFifo bool) string {
	if value, ok := name.(string); ok && value != "" {
		return value
	}

	var output string
	if value, ok := prefix.(string); ok && value != "" {
		output = id.PrefixedUniqueId(value)
	} else {
		output = types.NewRandomName("queue")
	}

	if isSQS && isSQSFifo {
		return output + SQSFIFOQueueNameSuffix
	}

	return output
}

func resourceMNQQueueCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ any) error {
	isSQSFifo := d.Get("fifo_queue").(bool)

	var name string
	if d.Id() == "" {
		name = resourceMNQQueueName(d.Get("name"), d.Get("name_prefix"), true, isSQSFifo)
	} else {
		name = d.Get("name").(string)
	}

	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)

	if isSQSFifo {
		nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,75}\` + SQSFIFOQueueNameSuffix + `$`)
	}

	contentBasedDeduplication := d.Get("content_based_deduplication").(bool)
	if !isSQSFifo && contentBasedDeduplication {
		return errors.New("content-based deduplication can only be set for FIFO queue")
	}

	// Validate dead-letter queue configuration
	if deadLetterConfig, ok := d.GetOk("dead_letter_queue"); ok {
		deadLetterList := deadLetterConfig.([]any)
		if len(deadLetterList) > 0 {
			config := deadLetterList[0].(map[string]any)
			queueID := config["id"].(string)
			maxReceiveCount := config["max_receive_count"].(int)

			if queueID == "" || strings.Contains(queueID, "scaleway_mnq_sqs_queue") {
				return nil
			}

			if queueID == "" {
				return errors.New("dead-letter queue ID cannot be empty")
			}

			if maxReceiveCount < 1 || maxReceiveCount > 1000 {
				return errors.New("max_receive_count must be between 1 and 1,000")
			}
		}
	}

	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid queue name: %s (format is %s)", name, nameRegex.String())
	}

	return nil
}
