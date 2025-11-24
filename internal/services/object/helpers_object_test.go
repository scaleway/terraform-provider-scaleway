package object_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/object"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/stretchr/testify/assert"
)

const (
	objectTestsMainRegion      = "nl-ams"
	objectTestsSecondaryRegion = "pl-waw"
)

func TestExpandObjectBucketTags(t *testing.T) {
	tests := []struct {
		name string
		tags any
		want []s3Types.Tag
	}{
		{
			name: "no tags",
			tags: map[string]any{},
			want: []s3Types.Tag(nil),
		},
		{
			name: "single tag",
			tags: map[string]any{
				"key1": "val1",
			},
			want: []s3Types.Tag{
				{Key: scw.StringPtr("key1"), Value: types.ExpandStringPtr("val1")},
			},
		},
		{
			name: "many tags",
			tags: map[string]any{
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

func TestFlattenObjectBucketVersioning(t *testing.T) {
	tests := []struct {
		name     string
		input    *s3.GetBucketVersioningOutput
		expected []map[string]any
	}{
		{
			name:  "nil input",
			input: nil,
			expected: []map[string]any{
				{"enabled": false},
			},
		},
		{
			name: "versioning enabled",
			input: &s3.GetBucketVersioningOutput{
				Status: s3Types.BucketVersioningStatusEnabled,
			},
			expected: []map[string]any{
				{"enabled": true},
			},
		},
		{
			name: "versioning suspended",
			input: &s3.GetBucketVersioningOutput{
				Status: s3Types.BucketVersioningStatusSuspended,
			},
			expected: []map[string]any{
				{"enabled": false},
			},
		},
		{
			name:  "versioning empty struct",
			input: &s3.GetBucketVersioningOutput{},
			expected: []map[string]any{
				{"enabled": false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := object.FlattenObjectBucketVersioning(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, 1, "Result should contain exactly one map")
			assert.Contains(t, result[0], "enabled", "Result map should contain 'enabled' key")
			_, ok := result[0]["enabled"].(bool)
			assert.True(t, ok, "'enabled' value should be a boolean")
		})
	}
}
