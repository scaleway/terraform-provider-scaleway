package scaleway

import (
	"github.com/aws/aws-sdk-go/service/s3"
)

func flattenObjectBucketTags(tagsSet []*s3.Tag) map[string]interface{} {
	tags := map[string]interface{}{}

	for _, tagSet := range tagsSet {
		var key string
		var value string
		if tagSet.Key != nil {
			key = *tagSet.Key
		}
		if tagSet.Value != nil {
			value = *tagSet.Value
		}
		tags[key] = value
	}

	return tags
}

func expandObjectBucketTags(tags interface{}) []*s3.Tag {
	tagsSet := make([]*s3.Tag, 0)

	for key, value := range tags.(map[string]interface{}) {
		tagsSet = append(tagsSet, &s3.Tag{
			Key:   &key,
			Value: expandStringPtr(value),
		})
	}

	return tagsSet
}
