package scaleway

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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

// Sets a specific SNS attribute from the resource data
func snsResourceDataToAttribute(snsAttributes map[string]*string, snsAttribute string, resourceValue interface{}, resourcePath string, resourceSchemas map[string]*schema.Schema) error {
	resourceSchema := resolveSchemaPath(resourcePath, resourceSchemas)
	if resourceSchema == nil {
		return fmt.Errorf("unable to resolve schema for %s", resourcePath)
	}

	// Only set writable attributes
	if !resourceSchema.Optional && !resourceSchema.Required {
		return nil
	}

	var s string
	switch resourceSchema.Type {
	case schema.TypeBool:
		s = strconv.FormatBool(resourceValue.(bool))
	case schema.TypeInt:
		s = strconv.Itoa(resourceValue.(int))
	case schema.TypeString:
		s = resourceValue.(string)
	default:
		return fmt.Errorf("unsupported type %s for %s", resourceSchema.Type, resourcePath)
	}

	snsAttributes[snsAttribute] = &s
	return nil
}

// snsResourceDataToAttributes returns a map of attributes from SNS Topic resource
func snsResourceDataToAttributes(d *schema.ResourceData, resourceSchemas map[string]*schema.Schema) (map[string]*string, error) {
	attributes := make(map[string]*string)

	for attribute, resourcePath := range SQSAttributesToResourceMap {
		if v, ok := d.GetOk(resourcePath); ok {
			err := snsResourceDataToAttribute(attributes, attribute, v, resourcePath, resourceSchemas)
			if err != nil {
				return nil, err
			}
		}
	}

	return attributes, nil
}

// Sets a specific resource data from the SQS attribute
func snsAttributeToResourceData(values map[string]interface{}, value string, resourcePath string, resourceSchemas map[string]*schema.Schema) error {
	resourceSchema := resolveSchemaPath(resourcePath, resourceSchemas)
	if resourceSchema == nil {
		return fmt.Errorf("unable to resolve schema for %s", resourcePath)
	}

	switch resourceSchema.Type {
	case schema.TypeBool:
		b, _ := strconv.ParseBool(value)
		setResourceValue(values, resourcePath, b, resourceSchemas)
	case schema.TypeInt:
		i, _ := strconv.Atoi(value)
		setResourceValue(values, resourcePath, i, resourceSchemas)
	case schema.TypeString:
		setResourceValue(values, resourcePath, value, resourceSchemas)
	default:
		return fmt.Errorf("unsupported type %s for %s", resourceSchema.Type, resourcePath)
	}

	return nil
}

// snsAttributesToResourceData returns a map of valid values for a terraform schema from a topic attributes map
func snsAttributesToResourceData(attributes map[string]*string, resourceSchemas map[string]*schema.Schema) (map[string]interface{}, error) {
	values := make(map[string]interface{})

	for attribute, resourcePath := range SNSTopicAttributesToResourceMap {
		if value, ok := attributes[attribute]; ok && value != nil {
			err := snsAttributeToResourceData(values, *value, resourcePath, resourceSchemas)
			if err != nil {
				return nil, err
			}
		}
	}

	return values, nil
}

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
