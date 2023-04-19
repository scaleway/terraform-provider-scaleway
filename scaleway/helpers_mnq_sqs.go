package scaleway

import (
	"strconv"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const SQSFIFOQueueNameSuffix = ".fifo"

var SQSAttributesToResourceMap = map[string]string{
	sqs.QueueAttributeNameContentBasedDeduplication:     "content_based_deduplication",
	sqs.QueueAttributeNameMaximumMessageSize:            "max_message_size",
	sqs.QueueAttributeNameMessageRetentionPeriod:        "message_retention_seconds",
	sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds: "receive_wait_time_seconds",
	sqs.QueueAttributeNameVisibilityTimeout:             "visibility_timeout_seconds",
}

func SQSResourceToAttributes(elements map[string]*schema.Schema, data map[string]interface{}) map[string]*string {
	attributesToResource := make(map[string]*string)
	for attribute, element := range SQSAttributesToResourceMap {
		elementSchema, ok := elements[element]
		if !ok {
			continue
		}

		value, ok := data[element]
		if !ok {
			continue
		}

		var s string
		switch elementSchema.Type {
		case schema.TypeBool:
			s = strconv.FormatBool(value.(bool))
		case schema.TypeInt:
			s = strconv.Itoa(value.(int))
		case schema.TypeString:
			s = value.(string)
		default:
			continue
		}

		attributesToResource[attribute] = &s
	}

	return attributesToResource
}

func SQSAttributesToResource(elements map[string]*schema.Schema, data map[string]*string) map[string]interface{} {
	attributesToResource := make(map[string]interface{})
	for attribute, element := range SQSAttributesToResourceMap {
		elementSchema, ok := elements[element]
		if !ok {
			continue
		}

		value, ok := data[attribute]
		if !ok {
			continue
		}

		switch elementSchema.Type {
		case schema.TypeBool:
			attributesToResource[element] = *value == "true"
		case schema.TypeInt:
			i, _ := strconv.Atoi(*value)
			attributesToResource[element] = i
		case schema.TypeString:
			attributesToResource[element] = *value
		default:
			continue
		}
	}

	return attributesToResource
}
