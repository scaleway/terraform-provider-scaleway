package scaleway

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	// Maximum amount of time to wait for SQS queue attribute changes to propagate
	// This timeout should not be increased without strong consideration
	// as this will negatively impact user experience when configurations
	// have incorrect references or permissions.
	// Reference: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SetQueueAttributes.html
	queueAttributePropagationTimeout = 2 * time.Minute

	// If you delete a queue, you must wait at least 60 seconds before creating a queue with the same name.
	// ReferenceL https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_CreateQueue.html
	queueCreatedTimeout = 70 * time.Second
	queueReadTimeout    = 20 * time.Second

	queueAttributeStateNotEqual = "notequal"
	queueAttributeStateEqual    = "equal"
)

func newMNQAPI(d *schema.ResourceData, m interface{}) (*mnq.API, scw.Region, error) {
	meta := m.(*Meta)
	api := mnq.NewAPI(meta.scwClient)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func SQSClientWithRegion(d *schema.ResourceData, m interface{}) (*sqs.SQS, scw.Region, error) {
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	accessKey := d.Get("access_key").(string)
	projectID, isDefaultProjectID, err := extractProjectID(d, meta)
	if err == nil && !isDefaultProjectID {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}
	secretKey := d.Get("secret_key").(string)
	sqsClient, err := newSQSClient(meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return sqsClient, region, err
}

func newSQSClient(httpClient *http.Client, region, accessKey, secretKey string) (*sqs.SQS, error) {
	config := &aws.Config{}
	config.WithRegion(region)
	config.WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	config.WithEndpoint("http://sqs-sns.mnq." + region + ".scw.cloud")
	config.WithHTTPClient(httpClient)
	if strings.ToLower(os.Getenv("TF_LOG")) == "debug" {
		config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	s, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return sqs.New(s), nil
}

func mnqAPIWithRegionAndID(m interface{}, regionalID string) (*mnq.API, scw.Region, string, error) {
	meta := m.(*Meta)
	mnqAPI := mnq.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}
	return mnqAPI, region, ID, nil
}

// AttributeMap represents a map of Terraform resource attribute name to AWS API attribute name.
// Useful for SQS Queue or SNS Topic attribute handling.
type attributeInfo struct {
	alwaysSendConfiguredValueOnCreate bool
	apiAttributeName                  string
	tfType                            schema.ValueType
	tfComputed                        bool
	tfOptional                        bool
	isIAMPolicy                       bool
	missingSetToNil                   bool
	skipUpdate                        bool
}

type AttributeMap map[string]attributeInfo

// NewAttrMap returns a new AttributeMap from the specified Terraform resource attribute name to AWS API attribute name map and resource schema.
func NewAttrMap(attrMap map[string]string, schemaMap map[string]*schema.Schema) AttributeMap {
	attributeMap := make(AttributeMap)

	for tfAttributeName, apiAttributeName := range attrMap {
		if s, ok := schemaMap[tfAttributeName]; ok {
			attributeInfo := attributeInfo{
				apiAttributeName: apiAttributeName,
				tfType:           s.Type,
			}

			attributeInfo.tfComputed = s.Computed
			attributeInfo.tfOptional = s.Optional

			attributeMap[tfAttributeName] = attributeInfo
		} else {
			log.Printf("[ERROR] Unknown attribute: %s", tfAttributeName)
		}
	}

	return attributeMap
}

// WithIAMPolicyAttribute marks the specified Terraform attribute as holding an AWS IAM policy.
// AWS IAM policies get special handling.
// This method is intended to be chained with other similar helper methods in a builder pattern.
func (m AttributeMap) WithIAMPolicyAttribute(tfAttributeName string) AttributeMap {
	if attributeInfo, ok := m[tfAttributeName]; ok {
		attributeInfo.isIAMPolicy = true
		m[tfAttributeName] = attributeInfo
	}

	return m
}

// WithMissingSetToNil marks the specified Terraform attribute as being set to nil if it's missing after reading the API.
// An attribute name of "*" means all attributes get marked.
// This method is intended to be chained with other similar helper methods in a builder pattern.
func (m AttributeMap) WithMissingSetToNil(tfAttributeName string) AttributeMap {
	if tfAttributeName == "*" {
		for k, attributeInfo := range m {
			attributeInfo.missingSetToNil = true
			m[k] = attributeInfo
		}
	} else if attributeInfo, ok := m[tfAttributeName]; ok {
		attributeInfo.missingSetToNil = true
		m[tfAttributeName] = attributeInfo
	}

	return m
}

// WithSkipUpdate marks the specified Terraform attribute as skipping update handling.
// This method is intended to be chained with other similar helper methods in a builder pattern.
func (m AttributeMap) WithSkipUpdate(tfAttributeName string) AttributeMap {
	if attributeInfo, ok := m[tfAttributeName]; ok {
		attributeInfo.skipUpdate = true
		m[tfAttributeName] = attributeInfo
	}

	return m
}

// WithAlwaysSendConfiguredBooleanValueOnCreate marks the specified Terraform Boolean attribute as always having any configured value sent on resource create.
// By default, a Boolean value is only sent to the API on resource create if its configured value is true.
// This method is intended to be chained with other similar helper methods in a builder pattern.
func (m AttributeMap) WithAlwaysSendConfiguredBooleanValueOnCreate(tfAttributeName string) AttributeMap {
	if attributeInfo, ok := m[tfAttributeName]; ok && attributeInfo.tfType == schema.TypeBool {
		attributeInfo.alwaysSendConfiguredValueOnCreate = true
		m[tfAttributeName] = attributeInfo
	}

	return m
}

// ResourceDataToAPIAttributesCreate returns a map of AWS API attributes from Terraform ResourceData.
// The API attributes map is suitable for resource create.
func (m AttributeMap) ResourceDataToAPIAttributesCreate(d *schema.ResourceData) (map[string]string, error) {
	apiAttributes := map[string]string{}

	for tfAttributeName, attributeInfo := range m {
		// Purely Computed values aren't specified on creation.
		if attributeInfo.tfComputed && !attributeInfo.tfOptional {
			continue
		}

		var apiAttributeValue string
		configuredValue := d.GetRawConfig().GetAttr(tfAttributeName)
		tfOptionalComputed := attributeInfo.tfComputed && attributeInfo.tfOptional

		switch v, t := d.Get(tfAttributeName), attributeInfo.tfType; t {
		case schema.TypeBool:
			if v := v.(bool); v || (attributeInfo.alwaysSendConfiguredValueOnCreate && !configuredValue.IsNull()) {
				apiAttributeValue = strconv.FormatBool(v)
			}
		case schema.TypeInt:
			// On creation don't specify any zero Optional/Computed attribute integer values.
			if v := v.(int); !tfOptionalComputed || v != 0 {
				apiAttributeValue = strconv.Itoa(v)
			}
		case schema.TypeString:
			apiAttributeValue = v.(string)

			if attributeInfo.isIAMPolicy && apiAttributeValue != "" {
				policy, err := structure.NormalizeJsonString(apiAttributeValue)
				if err != nil {
					return nil, fmt.Errorf("policy (%s) is invalid JSON: %w", apiAttributeValue, err)
				}

				apiAttributeValue = policy
			}
		default:
			return nil, fmt.Errorf("attribute %s is of unsupported type: %d", tfAttributeName, t)
		}

		if apiAttributeValue != "" {
			apiAttributes[attributeInfo.apiAttributeName] = apiAttributeValue
		}
	}

	return apiAttributes, nil
}

func (m AttributeMap) ResourceDataToAPIAttributesUpdate(d *schema.ResourceData) (map[string]string, error) {
	apiAttributes := map[string]string{}

	for tfAttributeName, attributeInfo := range m {
		if attributeInfo.skipUpdate {
			continue
		}

		// Purely Computed values aren't specified on update.
		if attributeInfo.tfComputed && !attributeInfo.tfOptional {
			continue
		}

		if d.HasChange(tfAttributeName) {
			v := d.Get(tfAttributeName)

			var apiAttributeValue string

			switch t := attributeInfo.tfType; t {
			case schema.TypeBool:
				apiAttributeValue = strconv.FormatBool(v.(bool))
			case schema.TypeInt:
				apiAttributeValue = strconv.Itoa(v.(int))
			case schema.TypeString:
				apiAttributeValue = v.(string)

				if attributeInfo.isIAMPolicy {
					policy, err := structure.NormalizeJsonString(apiAttributeValue)
					if err != nil {
						return nil, fmt.Errorf("policy (%s) is invalid JSON: %w", apiAttributeValue, err)
					}

					apiAttributeValue = policy
				}
			default:
				return nil, fmt.Errorf("attribute %s is of unsupported type: %d", tfAttributeName, t)
			}

			apiAttributes[attributeInfo.apiAttributeName] = apiAttributeValue
		}
	}

	return apiAttributes, nil
}

func getQueueAttributeMap() AttributeMap {
	return NewAttrMap(map[string]string{
		"arn":                         sqs.QueueAttributeNameQueueArn,
		"content_based_deduplication": sqs.QueueAttributeNameContentBasedDeduplication,
		"delay_seconds":               sqs.QueueAttributeNameDelaySeconds,
		"fifo_queue":                  sqs.QueueAttributeNameFifoQueue,
		"kms_master_key_id":           sqs.QueueAttributeNameKmsMasterKeyId,
		"max_message_size":            sqs.QueueAttributeNameMaximumMessageSize,
		"message_retention_seconds":   sqs.QueueAttributeNameMessageRetentionPeriod,
		"receive_wait_time_seconds":   sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds,
		"visibility_timeout_seconds":  sqs.QueueAttributeNameVisibilityTimeout,
	}, queueSchema).WithIAMPolicyAttribute("policy").WithMissingSetToNil("*").WithAlwaysSendConfiguredBooleanValueOnCreate("sqs_managed_sse_enabled")
}
