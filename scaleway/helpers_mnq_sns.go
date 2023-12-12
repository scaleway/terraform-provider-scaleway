package scaleway

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func SNSClientWithRegion(d *schema.ResourceData, m interface{}) (*sns.SNS, scw.Region, error) {
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	endpoint := d.Get("endpoint").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)

	snsClient, err := newSNSClient(meta.httpClient, region.String(), endpoint, accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return snsClient, region, err
}

func SNSClientWithRegionAndID(d *schema.ResourceData, m interface{}, regionalID string) (*sns.SNS, scw.Region, string, error) {
	meta := m.(*Meta)
	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	endpoint := d.Get("sns_endpoint").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)

	snsClient, err := newSNSClient(meta.httpClient, region.String(), endpoint, accessKey, secretKey)
	if err != nil {
		return nil, "", "", err
	}

	return snsClient, region, ID, err
}

func newSNSClient(httpClient *http.Client, region string, endpoint string, accessKey string, secretKey string) (*sns.SNS, error) {
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

	return sns.New(s), nil
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
		output = newRandomName("topic")
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
		return fmt.Errorf("content-based deduplication can only be set for FIFO topics")
	}

	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid topic name: %s (format is %s)", name, nameRegex.String())
	}

	return nil
}
