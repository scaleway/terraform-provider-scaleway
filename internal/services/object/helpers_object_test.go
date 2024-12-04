package object_test

import (
	"testing"

	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object"
	"github.com/stretchr/testify/assert"
)

const (
	objectTestsMainRegion      = "nl-ams"
	objectTestsSecondaryRegion = "pl-waw"
)

func TestExpandObjectBucketTags(t *testing.T) {
	tests := []struct {
		name string
		tags interface{}
		want []s3Types.Tag
	}{
		{
			name: "no tags",
			tags: map[string]interface{}{},
			want: []s3Types.Tag(nil),
		},
		{
			name: "single tag",
			tags: map[string]interface{}{
				"key1": "val1",
			},
			want: []s3Types.Tag{
				{Key: scw.StringPtr("key1"), Value: types.ExpandStringPtr("val1")},
			},
		},
		{
			name: "many tags",
			tags: map[string]interface{}{
				"key1": "val1",
				"key2": "val2",
				"key3": "val3",
			},
			want: []s3Types.Tag{
				{Key: scw.StringPtr("key1"), Value: types.ExpandStringPtr("val1")},
				{Key: scw.StringPtr("key2"), Value: types.ExpandStringPtr("val2")},
				{Key: scw.StringPtr("key3"), Value: types.ExpandStringPtr("val3")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.want, object.ExpandObjectBucketTags(tt.tags))
		})
	}
}
