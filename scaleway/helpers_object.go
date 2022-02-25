package scaleway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultObjectBucketTimeout = 10 * time.Minute
	retryOnAWSAPI              = 2 * time.Minute
)

func newS3Client(httpClient *http.Client, region, accessKey, secretKey string) (*s3.S3, error) {
	config := &aws.Config{}
	config.WithRegion(region)
	config.WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	config.WithEndpoint("https://s3." + region + ".scw.cloud")
	config.WithHTTPClient(httpClient)
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
	return newS3Client(meta.httpClient, region.String(), accessKey, secretKey)
}

func s3ClientWithRegion(d *schema.ResourceData, m interface{}) (*s3.S3, scw.Region, error) {
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(meta.httpClient, region.String(), accessKey, secretKey)
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
	s3Client, err := newS3Client(meta.httpClient, region.String(), accessKey, secretKey)
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
	tagsSet := []*s3.Tag(nil)
	for key, value := range tags.(map[string]interface{}) {
		tagsSet = append(tagsSet, &s3.Tag{
			Key:   scw.StringPtr(key),
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
	vcl := []map[string]interface{}{{}}
	vcl[0]["enabled"] = versioningResponse.Status != nil && *versioningResponse.Status == s3.BucketVersioningStatusEnabled
	return vcl
}

func expandObjectBucketVersioning(v []interface{}) *s3.VersioningConfiguration {
	vc := &s3.VersioningConfiguration{}
	vc.Status = scw.StringPtr(s3.BucketVersioningStatusSuspended)
	if len(v) > 0 {
		if c := v[0].(map[string]interface{}); c["enabled"].(bool) {
			vc.Status = scw.StringPtr(s3.BucketVersioningStatusEnabled)
		}
	}
	return vc
}

func flattenBucketCORS(corsResponse interface{}) []map[string]interface{} {
	corsRules := make([]map[string]interface{}, 0)
	if cors, ok := corsResponse.(*s3.GetBucketCorsOutput); ok && len(cors.CORSRules) > 0 {
		corsRules = make([]map[string]interface{}, 0, len(cors.CORSRules))
		for _, ruleObject := range cors.CORSRules {
			rule := make(map[string]interface{})
			rule["allowed_headers"] = flattenSliceStringPtr(ruleObject.AllowedHeaders)
			rule["allowed_methods"] = flattenSliceStringPtr(ruleObject.AllowedMethods)
			rule["allowed_origins"] = flattenSliceStringPtr(ruleObject.AllowedOrigins)
			// Both the "ExposeHeaders" and "MaxAgeSeconds" might not be set.
			if ruleObject.AllowedOrigins != nil {
				rule["expose_headers"] = flattenSliceStringPtr(ruleObject.ExposeHeaders)
			}
			if ruleObject.MaxAgeSeconds != nil {
				rule["max_age_seconds"] = int(*ruleObject.MaxAgeSeconds)
			}
			corsRules = append(corsRules, rule)
		}
	}
	return corsRules
}

func expandBucketCORS(rawCors []interface{}, bucket string) []*s3.CORSRule {
	rules := make([]*s3.CORSRule, 0, len(rawCors))
	for _, cors := range rawCors {
		corsMap := cors.(map[string]interface{})
		r := &s3.CORSRule{}
		for k, v := range corsMap {
			l.Debugf("S3 bucket: %s, put CORS: %#v, %#v", bucket, k, v)
			if k == "max_age_seconds" {
				r.MaxAgeSeconds = scw.Int64Ptr(int64(v.(int)))
			} else {
				vMap := make([]*string, len(v.([]interface{})))
				for i, vv := range v.([]interface{}) {
					if str, ok := vv.(string); ok {
						vMap[i] = scw.StringPtr(str)
					}
				}
				switch k {
				case "allowed_headers":
					r.AllowedHeaders = vMap
				case "allowed_methods":
					r.AllowedMethods = vMap
				case "allowed_origins":
					r.AllowedOrigins = vMap
				case "expose_headers":
					r.ExposeHeaders = vMap
				}
			}
		}
		rules = append(rules, r)
	}
	return rules
}

func deleteS3ObjectVersion(conn *s3.S3, bucketName string, key string, versionID string, force bool) error {
	input := &s3.DeleteObjectInput{
		Bucket: scw.StringPtr(bucketName),
		Key:    scw.StringPtr(key),
	}
	if versionID != "" {
		input.VersionId = scw.StringPtr(versionID)
	}
	if force {
		input.BypassGovernanceRetention = scw.BoolPtr(force)
	}

	_, err := conn.DeleteObject(input)
	return err
}

// removeS3ObjectVersionLegalHold remove legal hold from an ObjectVersion if it is on
// returns true if legal hold was removed
func removeS3ObjectVersionLegalHold(conn *s3.S3, bucketName string, objectVersion *s3.ObjectVersion) (bool, error) {
	objectHead, err := conn.HeadObject(&s3.HeadObjectInput{
		Bucket:    scw.StringPtr(bucketName),
		Key:       objectVersion.Key,
		VersionId: objectVersion.VersionId,
	})
	if err != nil {
		err = fmt.Errorf("failed to get S3 object meta data: %s", err)
		return false, err
	}
	if aws.StringValue(objectHead.ObjectLockLegalHoldStatus) != s3.ObjectLockLegalHoldStatusOn {
		return false, nil

	}
	_, err = conn.PutObjectLegalHold(&s3.PutObjectLegalHoldInput{
		Bucket:    scw.StringPtr(bucketName),
		Key:       objectVersion.Key,
		VersionId: objectVersion.VersionId,
		LegalHold: &s3.ObjectLockLegalHold{
			Status: scw.StringPtr(s3.ObjectLockLegalHoldStatusOff),
		},
	})
	if err != nil {
		err = fmt.Errorf("failed to put S3 object legal hold: %s", err)
		return false, err
	}
	return true, nil

}

func deleteS3ObjectVersions(ctx context.Context, conn *s3.S3, bucketName string, force bool) error {
	var err error
	listInput := &s3.ListObjectVersionsInput{
		Bucket: scw.StringPtr(bucketName),
	}
	listErr := conn.ListObjectVersionsPagesWithContext(ctx, listInput, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		for _, objectVersion := range page.Versions {
			objectKey := aws.StringValue(objectVersion.Key)
			objectVersionID := aws.StringValue(objectVersion.VersionId)
			err = deleteS3ObjectVersion(conn, bucketName, objectKey, objectVersionID, force)

			if isS3Err(err, ErrCodeAccessDenied, "") && force {
				legalHoldRemoved, errLegal := removeS3ObjectVersionLegalHold(conn, bucketName, objectVersion)
				if errLegal != nil {
					err = fmt.Errorf("failed to remove legal hold: %s", errLegal)
					return false
				}
				if legalHoldRemoved {
					err = deleteS3ObjectVersion(conn, bucketName, objectKey, objectVersionID, force)
				}
			}
			if err != nil {
				err = fmt.Errorf("failed to delete S3 object: %s", err)
				return false
			}
		}
		return true
	})
	if listErr != nil {
		return fmt.Errorf("error listing S3 objects: %s", err)
	}
	if err != nil {
		return err
	}
	listErr = conn.ListObjectVersionsPagesWithContext(ctx, listInput, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		for _, deleteMarkerEntry := range page.DeleteMarkers {
			deleteMarkerKey := aws.StringValue(deleteMarkerEntry.Key)
			deleteMarkerVersionsID := aws.StringValue(deleteMarkerEntry.VersionId)
			err = deleteS3ObjectVersion(conn, bucketName, deleteMarkerKey, deleteMarkerVersionsID, force)

			if err != nil {
				err = fmt.Errorf("failed to delete S3 object delete marker: %s", err)
				return false
			}
		}
		return true
	})
	if listErr != nil {
		return fmt.Errorf("error listing S3 objects for delete markers: %s", err)
	}
	if err != nil {
		return err
	}
	return nil
}

func transitionHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["days"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["storage_class"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return StringHashcode(buf.String())
}

// StringHashcode hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// and non-negative integer. Here we cast to an integer
// and invert it if the result is negative.
func StringHashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func retryOnAWSCode(ctx context.Context, code string, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := resource.RetryContext(ctx, retryOnAWSAPI, func() *resource.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			if tfawserr.ErrCodeEquals(err, code) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if TimedOut(err) {
		resp, err = f()
	}

	return resp, err
}

const (
	// TransitionStorageClassStandard is a TransitionStorageClass enum value
	TransitionStorageClassStandard = "STANDARD"

	// TransitionStorageClassGlacier is a TransitionStorageClass enum value
	TransitionStorageClassGlacier = "GLACIER"

	// TransitionStorageClassOnezoneIa is a TransitionStorageClass enum value
	TransitionStorageClassOnezoneIa = "ONEZONE_IA"
)

// TransitionSCWStorageClassValues returns all elements of the TransitionStorageClass enum supported by scaleway
func TransitionSCWStorageClassValues() []string {
	return []string{
		TransitionStorageClassStandard,
		TransitionStorageClassGlacier,
		TransitionStorageClassOnezoneIa,
	}
}
