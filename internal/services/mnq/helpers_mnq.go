package mnq

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/smithy-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mnq "github.com/scaleway/scaleway-sdk-go/api/mnq/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	AWSErrQueueDeletedRecently = "AWS.SimpleQueueService.QueueDeletedRecently"
	AWSErrNonExistentQueue     = "AWS.SimpleQueueService.NonExistentQueue"

	// SQS ARN prefix
	SQSPrefix = "arn:scw:sqs:"

	// Dead letter queue resource path
	DeadLetterQueuePath = "dead_letter_queue"
)

func newMNQNatsAPI(d *schema.ResourceData, m any) (*mnq.NatsAPI, scw.Region, error) {
	api := mnq.NewNatsAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func NewNatsAPIWithRegionAndID(m any, regionalID string) (*mnq.NatsAPI, scw.Region, string, error) {
	api := mnq.NewNatsAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func newSQSAPI(d *schema.ResourceData, m any) (*mnq.SqsAPI, scw.Region, error) {
	api := mnq.NewSqsAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func NewSQSAPIWithRegionAndID(m any, regionalID string) (*mnq.SqsAPI, scw.Region, string, error) {
	api := mnq.NewSqsAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func newMNQSNSAPI(d *schema.ResourceData, m any) (*mnq.SnsAPI, scw.Region, error) {
	api := mnq.NewSnsAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

func NewSNSAPIWithRegionAndID(m any, regionalID string) (*mnq.SnsAPI, scw.Region, string, error) {
	api := mnq.NewSnsAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func composeMNQID(region scw.Region, projectID string, queueName string) string {
	return fmt.Sprintf("%s/%s/%s", region, projectID, queueName)
}

func DecomposeMNQID(id string) (region scw.Region, projectID string, name string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid ID format: %q", id)
	}

	region, err = scw.ParseRegion(parts[0])
	if err != nil {
		return "", "", "", err
	}

	return region, parts[1], parts[2], nil
}

type ARN struct {
	Subject         string
	Region          scw.Region
	ProjectID       string
	ResourceName    string
	ExtraResourceID string
}

func (a ARN) String() string {
	if a.ExtraResourceID == "" {
		return fmt.Sprintf("arn:scw:%s:%s:project-%s:%s", a.Subject, a.Region, a.ProjectID, a.ResourceName)
	}

	return fmt.Sprintf("arn:scw:%s:%s:project-%s:%s:%s", a.Subject, a.Region, a.ProjectID, a.ResourceName, a.ExtraResourceID)
}

// decomposeARN decomposes an arn with a potential extra-resource-id
// example: arn:scw:sns:fr-par:project-d4730602-0495-4bb6-bb94-de3a9b000660:test-mnq-sns-topic-basic:b9f52ee5-fa03-42ad-9065-587e3e22efd9
// the last id may be omitted
func decomposeARN(arn string) (*ARN, error) {
	elems := strings.Split(arn, ":")
	if len(elems) < 6 || len(elems) > 7 {
		return nil, fmt.Errorf("wrong number of parts in arn, expected 6 or 7, got %d", len(elems))
	}

	if elems[0] != "arn" {
		return nil, fmt.Errorf("expected part 0 to be \"arn\", got %q", elems[0])
	}

	if elems[1] != "scw" {
		return nil, fmt.Errorf("expected part 1 to be \"scw\", got %q", elems[1])
	}

	region, err := scw.ParseRegion(elems[3])
	if err != nil {
		return nil, fmt.Errorf("expected part 2 to be a valid region: %w", err)
	}

	projectID, found := strings.CutPrefix(elems[4], "project-")
	if !found {
		return nil, errors.New("expected part 3 to have format \"project-{uuid}\"")
	}

	a := &ARN{
		Subject:      elems[0],
		Region:       region,
		ProjectID:    projectID,
		ResourceName: elems[5],
	}
	if len(elems) == 7 {
		a.ExtraResourceID = elems[6]
	}

	return a, nil
}

func composeARN(subject string, region scw.Region, projectID string, resourceName string) string {
	return ARN{
		Subject:      subject,
		Region:       region,
		ProjectID:    projectID,
		ResourceName: resourceName,
	}.String()
}

func ComposeSNSARN(region scw.Region, projectID string, resourceName string) string {
	return composeARN("sns", region, projectID, resourceName)
}

func setResourceValue(values map[string]any, resourcePath string, value any, resourceSchemas map[string]*schema.Schema) {
	parts := strings.Split(resourcePath, ".")
	if len(parts) > 1 {
		if _, ok := values[parts[0]]; !ok {
			values[parts[0]] = []any{make(map[string]any)}
		}

		setResourceValue(values[parts[0]].([]any)[0].(map[string]any), strings.Join(parts[2:], "."), value, resourceSchemas[parts[0]].Elem.(*schema.Resource).SchemaFunc())

		return
	}

	values[resourcePath] = value
}

// Get the schema for the resource path (e.g. a.0.b gives b's schema)
func resolveSchemaPath(resourcePath string, resourceSchemas map[string]*schema.Schema) *schema.Schema {
	if resourceSchema, ok := resourceSchemas[resourcePath]; ok {
		return resourceSchema
	}

	parts := strings.Split(resourcePath, ".")
	if len(parts) > 1 {
		return resolveSchemaPath(strings.Join(parts[2:], "."), resourceSchemas[parts[0]].Elem.(*schema.Resource).SchemaFunc())
	}

	return nil
}

// Sets a specific SNS attribute from the resource data
func awsResourceDataToAttribute(awsAttributes map[string]string, awsAttribute string, resourceValue any, resourcePath string, resourceSchemas map[string]*schema.Schema) error {
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
	case schema.TypeList:
		if resourcePath == DeadLetterQueuePath {
			deadLetterConfig := resourceValue.([]any)
			if len(deadLetterConfig) > 0 {
				config := deadLetterConfig[0].(map[string]any)
				queueID := config["id"].(string)
				maxReceiveCount := config["max_receive_count"].(int)

				var scwARN string

				switch {
				case strings.HasPrefix(queueID, SQSPrefix):
					scwARN = queueID
				case strings.Contains(queueID, "/"):
					parts := strings.Split(queueID, "/")
					if len(parts) == 3 {
						region := parts[0]
						projectID := parts[1]
						queueName := parts[2]

						scwARN = fmt.Sprintf("arn:scw:sqs:%s:project-%s:%s", region, projectID, queueName)
					} else {
						return fmt.Errorf("invalid queue ID format for dead-letter queue: %s (expected region/project-id/queue-name or arn:scw:sqs:region:project-id:queue-name)", queueID)
					}
				default:
					scwARN = queueID
				}

				redrivePolicy := map[string]any{
					"deadLetterTargetArn": scwARN,
					"maxReceiveCount":     maxReceiveCount,
				}

				jsonData, err := json.Marshal(redrivePolicy)
				if err != nil {
					return fmt.Errorf("failed to marshal redrive policy: %w", err)
				}

				s = string(jsonData)
			}
		} else {
			return fmt.Errorf("unsupported list type for %s", resourcePath)
		}
	default:
		return fmt.Errorf("unsupported type %s for %s", resourceSchema.Type, resourcePath)
	}

	awsAttributes[awsAttribute] = s

	return nil
}

// awsResourceDataToAttributes returns a map of attributes from a terraform schema and a conversion map
func awsResourceDataToAttributes(d *schema.ResourceData, resourceSchemas map[string]*schema.Schema, attributesToResourceMap map[string]string) (map[string]string, error) {
	attributes := make(map[string]string)

	for attribute, resourcePath := range attributesToResourceMap {
		if v, ok := d.GetOk(resourcePath); ok {
			err := awsResourceDataToAttribute(attributes, attribute, v, resourcePath, resourceSchemas)
			if err != nil {
				return nil, err
			}
		}
	}

	return attributes, nil
}

// awsAttributeToResourceData sets a specific resource data from the given attribute
func awsAttributeToResourceData(values map[string]any, value string, resourcePath string, resourceSchemas map[string]*schema.Schema) error {
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
	case schema.TypeList:
		if resourcePath == DeadLetterQueuePath && value != "" {
			var redrivePolicy map[string]any
			if err := json.Unmarshal([]byte(value), &redrivePolicy); err != nil {
				return fmt.Errorf("failed to unmarshal redrive policy: %w", err)
			}

			deadLetterTargetArn := redrivePolicy["deadLetterTargetArn"].(string)

			var terraformID string

			if strings.HasPrefix(deadLetterTargetArn, SQSPrefix) {
				parts := strings.Split(deadLetterTargetArn, ":")
				if len(parts) >= 6 {
					region := parts[3]
					projectID := strings.TrimPrefix(parts[4], "project-")
					queueName := parts[5]
					terraformID = fmt.Sprintf("%s/%s/%s", region, projectID, queueName)
				} else {
					terraformID = deadLetterTargetArn
				}
			} else {
				terraformID = deadLetterTargetArn
			}

			deadLetterConfig := map[string]any{
				"id":                terraformID,
				"max_receive_count": int(redrivePolicy["maxReceiveCount"].(float64)),
			}

			setResourceValue(values, resourcePath, []any{deadLetterConfig}, resourceSchemas)
		} else {
			return fmt.Errorf("unsupported list type for %s", resourcePath)
		}
	default:
		return fmt.Errorf("unsupported type %s for %s", resourceSchema.Type, resourcePath)
	}

	return nil
}

func awsAttributesToResourceData(attributes map[string]string, resourceSchemas map[string]*schema.Schema, attributesToResourceMap map[string]string) (map[string]any, error) {
	values := make(map[string]any)

	for attribute, resourcePath := range attributesToResourceMap {
		if value, ok := attributes[attribute]; ok {
			err := awsAttributeToResourceData(values, value, resourcePath, resourceSchemas)
			if err != nil {
				return nil, err
			}
		}
	}

	return values, nil
}

func IsAWSErrorCode(err error, code string) bool {
	var apiErr *smithy.GenericAPIError
	if errors.As(err, &apiErr) && apiErr.Code == code {
		return true
	}

	return false
}
