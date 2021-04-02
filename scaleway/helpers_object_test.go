package scaleway

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func TestExpandObjectBucketTags(t *testing.T) {
	tests := []struct {
		name string
		tags interface{}
		want []*s3.Tag
	}{
		{
			name: "no tags",
			tags: map[string]interface{}{},
			want: []*s3.Tag(nil),
		},
		{
			name: "single tag",
			tags: map[string]interface{}{"key1": "val1"},
			want: []*s3.Tag{
				{Key: scw.StringPtr("key1"), Value: scw.StringPtr("val1")},
			},
		},
		{
			name: "many tags",
			tags: map[string]interface{}{"key1": "val1", "key2": "val2", "key3": "val3"},
			want: []*s3.Tag{
				{Key: scw.StringPtr("key1"), Value: scw.StringPtr("val1")},
				{Key: scw.StringPtr("key2"), Value: scw.StringPtr("val2")},
				{Key: scw.StringPtr("key3"), Value: scw.StringPtr("val3")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.want, expandObjectBucketTags(tt.tags))
		})
	}
}
