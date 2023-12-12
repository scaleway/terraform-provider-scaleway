package scaleway

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	natsjwt "github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultMNQQueueTimeout       = 5 * time.Minute
	defaultMNQQueueRetryInterval = 5 * time.Second

	DefaultQueueMaximumMessageSize            = 262_144 // 256 KiB.
	DefaultQueueMessageRetentionPeriod        = 345_600 // 4 days.
	DefaultQueueReceiveMessageWaitTimeSeconds = 0
	DefaultQueueVisibilityTimeout             = 30
)

func SQSClientWithRegion(d *schema.ResourceData, m interface{}) (*sqs.SQS, scw.Region, error) {
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	endpoint := d.Get("sqs_endpoint").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)

	sqsClient, err := newSQSClient(meta.httpClient, region.String(), endpoint, accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return sqsClient, region, err
}

func newSQSClient(httpClient *http.Client, region string, endpoint string, accessKey string, secretKey string) (*sqs.SQS, error) {
	config := &aws.Config{}
	config.WithRegion(region)
	config.WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	config.WithEndpoint(strings.ReplaceAll(endpoint, "{region}", region))
	config.WithHTTPClient(httpClient)
	if logging.IsDebugOrHigher() {
		config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	s, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return sqs.New(s), nil
}

func NATSClientWithRegion(d *schema.ResourceData, m interface{}) (nats.JetStreamContext, scw.Region, error) { //nolint:ireturn
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	endpoint := d.Get("endpoint").(string)
	credentials := d.Get("credentials").(string)
	js, err := newNATSJetStreamClient(region.String(), endpoint, credentials)
	if err != nil {
		return nil, "", err
	}

	return js, region, err
}

func newNATSJetStreamClient(region string, endpoint string, credentials string) (nats.JetStreamContext, error) { //nolint:ireturn
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
	sqs.QueueAttributeNameMaximumMessageSize:            "message_max_size",
	sqs.QueueAttributeNameMessageRetentionPeriod:        "message_max_age",
	sqs.QueueAttributeNameFifoQueue:                     "fifo_queue",
	sqs.QueueAttributeNameContentBasedDeduplication:     "content_based_deduplication",
	sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds: "receive_wait_time_seconds",
	sqs.QueueAttributeNameVisibilityTimeout:             "visibility_timeout_seconds",
}

// Returns all managed SQS attribute names
func getSQSAttributeNames() []*string {
	var attributeNames []*string

	for attribute := range SQSAttributesToResourceMap {
		attributeNames = append(attributeNames, aws.String(attribute))
	}

	return attributeNames
}

func resourceMNQQueueName(name interface{}, prefix interface{}, isSQS bool, isSQSFifo bool) string {
	if value, ok := name.(string); ok && value != "" {
		return value
	}

	var output string
	if value, ok := prefix.(string); ok && value != "" {
		output = id.PrefixedUniqueId(value)
	} else {
		output = newRandomName("queue")
	}
	if isSQS && isSQSFifo {
		return output + SQSFIFOQueueNameSuffix
	}

	return output
}

func resourceMNQQueueCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
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
		return fmt.Errorf("content-based deduplication can only be set for FIFO queue")
	}

	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid queue name: %s (format is %s)", name, nameRegex.String())
	}

	return nil
}
