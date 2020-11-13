package scaleway

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultObjectTimeout = 10 * time.Minute
)

func newS3Client(region, accessKey, secretKey string) (*s3.S3, error) {
	config := &aws.Config{}
	config.WithRegion(region)
	config.WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	config.WithEndpoint("https://s3." + region + ".scw.cloud")
	if strings.ToLower(os.Getenv("TF_LOG")) == "debug" {
		config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	s, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return s3.New(s), nil
}

func newS3ClientFromMeta(meta *Meta) (*s3.S3, error) {
	region, _ := meta.scwClient.GetDefaultRegion()
	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()
	return newS3Client(region.String(), accessKey, secretKey)
}

func s3ClientWithRegion(d *schema.ResourceData, m interface{}) (*s3.S3, scw.Region, error) {
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return s3Client, region, err
}

func s3ClientWithRegionAndName(m interface{}, name string) (*s3.S3, scw.Region, string, error) {
	meta := m.(*Meta)
	region, name, err := parseRegionalID(name)
	if err != nil {
		return nil, "", name, err
	}
	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()
	s3Client, err := newS3Client(region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", "", err
	}
	return s3Client, region, name, err
}

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

func objectBucketEndpointURL(bucketName string, region scw.Region) string {
	return fmt.Sprintf("https://%s.s3.%s.scw.cloud", bucketName, region)
}

// Returns true if the error matches all these conditions:
//  * err is of type awserr.Error
//  * Error.Code() matches code
//  * Error.Message() contains message
func isS3Err(err error, code string, message string) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == code && strings.Contains(awsErr.Message(), message)
	}
	return false
}

func flattenObjectBucketVersioning(versioningResponse *s3.GetBucketVersioningOutput) []map[string]interface{} {
	vcl := make([]map[string]interface{}, 0, 1)
	vc := make(map[string]interface{})
	if versioningResponse.Status != nil && aws.StringValue(versioningResponse.Status) == s3.BucketVersioningStatusEnabled {
		vc["enabled"] = true
	} else {
		vc["enabled"] = false
	}
	vcl = append(vcl, vc)
	return vcl
}

func expandObjectBucketVersioning(v []interface{}) *s3.VersioningConfiguration {
	vc := &s3.VersioningConfiguration{}

	if len(v) > 0 {
		c := v[0].(map[string]interface{})

		if c["enabled"].(bool) {
			vc.Status = aws.String(s3.BucketVersioningStatusEnabled)
		} else {
			vc.Status = aws.String(s3.BucketVersioningStatusSuspended)
		}
	} else {
		vc.Status = aws.String(s3.BucketVersioningStatusSuspended)
	}
	return vc
}
