package mnq

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func SNSClientWithRegion(ctx context.Context, m interface{}, d *schema.ResourceData) (*sns.Client, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	endpoint := d.Get("sns_endpoint").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)

	snsClient, err := NewSNSClient(ctx, meta.ExtractHTTPClient(m), region.String(), endpoint, accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return snsClient, region, err
}

func SNSClientWithRegionFromID(ctx context.Context, d *schema.ResourceData, m interface{}, regionalID string) (*sns.Client, scw.Region, error) {
	tab := strings.SplitN(regionalID, "/", 2)
	if len(tab) != 2 {
		return nil, "", errors.New("invalid ID format, expected parts separated by slashes")
	}

	region, err := scw.ParseRegion(tab[0])
	if err != nil {
		return nil, "", fmt.Errorf("invalid region in id: %w", err)
	}

	endpoint := d.Get("sns_endpoint").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)

	snsClient, err := NewSNSClient(ctx, meta.ExtractHTTPClient(m), region.String(), endpoint, accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return snsClient, region, err
}

func NewSNSClient(ctx context.Context, httpClient *http.Client, region string, endpoint string, accessKey string, secretKey string) (*sns.Client, error) {
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

	return sns.NewFromConfig(customConfig), nil
}

func composeMNQSubscriptionID(region scw.Region, projectID string, topicName string, subscriptionID string) string {
	return fmt.Sprintf("%s/%s/%s/%s", region, projectID, topicName, subscriptionID)
}

func DecomposeMNQSubscriptionID(id string) (arn *ARN, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid ID format: %q", id)
	}

	region, err := scw.ParseRegion(parts[0])
	if err != nil {
		return nil, err
	}

	return &ARN{
		Subject:         "sns",
		Region:          region,
		ProjectID:       parts[1],
		ResourceName:    parts[2],
		ExtraResourceID: parts[3],
	}, nil
}

var (
	SNSTopicAttributesToResourceMap = map[string]string{
		"ContentBasedDeduplication": "content_based_deduplication",
		"FifoTopic":                 "fifo_topic",
		"Owner":                     "owner",
		"TopicArn":                  "arn",
	}
	SNSTopicSubscriptionAttributesToResourceMap = map[string]string{
		"RedrivePolicy": "redrive_policy",
	}
)

func resourceMNQSNSTopicName(name interface{}, prefix interface{}, isSQS bool, isSQSFifo bool) string {
	if value, ok := name.(string); ok && value != "" {
		return value
	}

	var output string
	if value, ok := prefix.(string); ok && value != "" {
		output = id.PrefixedUniqueId(value)
	} else {
		output = types.NewRandomName("topic")
	}

	if isSQS && isSQSFifo {
		return output + SQSFIFOQueueNameSuffix
	}

	return output
}

func resourceMNQSSNSTopicCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	isFifoTopic := d.Get("fifo_topic").(bool)

	var name string
	if d.Id() == "" {
		name = resourceMNQSNSTopicName(d.Get("name"), d.Get("name_prefix"), true, isFifoTopic)
	} else {
		name = d.Get("name").(string)
	}

	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)

	if isFifoTopic {
		nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,75}\` + SQSFIFOQueueNameSuffix + `$`)
	}

	contentBasedDeduplication := d.Get("content_based_deduplication").(bool)
	if !isFifoTopic && contentBasedDeduplication {
		return errors.New("content-based deduplication can only be set for FIFO topics")
	}

	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid topic name: %s (format is %s)", name, nameRegex.String())
	}

	return nil
}
