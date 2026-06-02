package object_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				{Key: new("key1"), Value: types.ExpandStringPtr("val1")},
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
				{Key: new("key1"), Value: types.ExpandStringPtr("val1")},
				{Key: new("key2"), Value: types.ExpandStringPtr("val2")},
				{Key: new("key3"), Value: types.ExpandStringPtr("val3")},
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

func TestComputeObjectBucketURLs(t *testing.T) {
	resourceSchema := map[string]*schema.Schema{
		"s3_use_path_style": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "",
		},
		"endpoints": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}

	tests := []struct {
		name                string
		d                   *schema.ResourceData
		m                   any
		bucketName          string
		region              scw.Region
		expectedEndpoint    string
		expectedAPIEndpoint string
	}{
		{
			name: "s3 valid endpoint without path style",
			d: schema.TestResourceDataRaw(t, resourceSchema, map[string]any{
				"s3_use_path_style": false,
				"endpoints": map[string]any{
					"s3": "https://mys3.endpoint.com",
				},
			}),
			bucketName:          "my-bucket",
			region:              scw.RegionPlWaw,
			expectedEndpoint:    "https://my-bucket.mys3.endpoint.com",
			expectedAPIEndpoint: "https://mys3.endpoint.com",
		},
		{
			name: "s3 valid endpoint with path style",
			d: schema.TestResourceDataRaw(t, resourceSchema, map[string]any{
				"s3_use_path_style": true,
				"endpoints": map[string]any{
					"s3": "https://mys3.endpoint.com",
				},
			}),
			bucketName:          "my-bucket-hehe",
			region:              scw.RegionPlWaw,
			expectedEndpoint:    "https://mys3.endpoint.com/my-bucket-hehe",
			expectedAPIEndpoint: "https://mys3.endpoint.com",
		},
		{
			name: "s3 empty endpoint with path style",
			d: schema.TestResourceDataRaw(t, resourceSchema, map[string]any{
				"s3_use_path_style": true,
				"endpoints": map[string]any{
					"s3": "",
				},
			}),
			bucketName:          "my-bucket-hehe",
			region:              scw.RegionPlWaw,
			expectedEndpoint:    "https://s3.pl-waw.scw.cloud/my-bucket-hehe",
			expectedAPIEndpoint: "https://s3.pl-waw.scw.cloud",
		},
		{
			name: "s3 empty endpoint with path style, version 2",
			d: schema.TestResourceDataRaw(t, resourceSchema, map[string]any{
				"s3_use_path_style": true,
			}),
			bucketName:          "my-bucket-hehe",
			region:              scw.RegionPlWaw,
			expectedEndpoint:    "https://s3.pl-waw.scw.cloud/my-bucket-hehe",
			expectedAPIEndpoint: "https://s3.pl-waw.scw.cloud",
		},
		{
			name:                "s3 empty endpoint without path style",
			d:                   schema.TestResourceDataRaw(t, resourceSchema, map[string]any{}),
			bucketName:          "my-bucket-hehe",
			region:              scw.RegionPlWaw,
			expectedEndpoint:    "https://my-bucket-hehe.s3.pl-waw.scw.cloud",
			expectedAPIEndpoint: "https://s3.pl-waw.scw.cloud",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, apiEndpoint := object.ComputeObjectBucketURLs(tt.d, tt.m, tt.bucketName, tt.region)
			assert.Equal(t, tt.expectedEndpoint, endpoint)
			assert.Equal(t, tt.expectedAPIEndpoint, apiEndpoint)
		})
	}
}
