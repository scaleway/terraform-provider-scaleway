package scaleway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/workerpool"
)

const (
	defaultObjectBucketTimeout = 10 * time.Minute

	maxObjectVersionDeletionWorkers = 8

	objectTestsMainRegion      = "nl-ams"
	objectTestsSecondaryRegion = "pl-waw"

	errCodeForbidden = "Forbidden"
)

func newS3Client(httpClient *http.Client, region, accessKey, secretKey string) (*s3.S3, error) {
	config := &aws.Config{}
	config.WithRegion(region)
	config.WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, ""))
	if ep := os.Getenv("SCW_S3_ENDPOINT"); ep != "" {
		config.WithEndpoint(ep)
	} else {
		config.WithEndpoint("https://s3." + region + ".scw.cloud")
	}
	config.WithHTTPClient(httpClient)
	if logging.IsDebugOrHigher() {
		config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	s, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return s3.New(s), nil
}

func newS3ClientFromMeta(meta *Meta, region string) (*s3.S3, error) {
	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()

	projectID, _ := meta.scwClient.GetDefaultProjectID()
	if projectID != "" {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}

	if region == "" {
		defaultRegion, _ := meta.scwClient.GetDefaultRegion()
		region = defaultRegion.String()
	}

	return newS3Client(meta.httpClient, region, accessKey, secretKey)
}

func s3ClientWithRegion(d *schema.ResourceData, m interface{}) (*s3.S3, scw.Region, error) {
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	accessKey, _ := meta.scwClient.GetAccessKey()
	if projectID, _, err := extractProjectID(d, meta); err == nil {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return s3Client, region, err
}

func s3ClientWithRegionAndName(d *schema.ResourceData, m interface{}, id string) (*s3.S3, scw.Region, string, error) {
	meta := m.(*Meta)
	region, name, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	parts := strings.Split(name, "@")
	if len(parts) > 2 {
		return nil, "", "", fmt.Errorf("invalid ID %q: expected ID in format <region>/<name>[@<project_id>]", id)
	}
	name = parts[0]

	d.SetId(fmt.Sprintf("%s/%s", region, name))

	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()

	if len(parts) == 2 {
		accessKey = accessKeyWithProjectID(accessKey, parts[1])
	} else {
		projectID, _, err := extractProjectID(d, meta)
		if err == nil {
			accessKey = accessKeyWithProjectID(accessKey, projectID)
		}
	}

	s3Client, err := newS3Client(meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", "", err
	}

	return s3Client, region, name, nil
}

func s3ClientWithRegionAndNestedName(d *schema.ResourceData, m interface{}, name string) (*s3.S3, scw.Region, string, string, error) {
	meta := m.(*Meta)
	region, outerID, innerID, err := regional.ParseNestedID(name)
	if err != nil {
		return nil, "", "", "", err
	}

	accessKey, _ := meta.scwClient.GetAccessKey()
	if projectID, _, err := extractProjectID(d, meta); err == nil {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", "", "", err
	}

	return s3Client, region, outerID, innerID, err
}

func s3ClientWithRegionWithNameACL(d *schema.ResourceData, m interface{}, name string) (*s3.S3, scw.Region, string, string, error) {
	meta := m.(*Meta)
	region, name, outerID, err := locality.ParseLocalizedNestedOwnerID(name)
	if err != nil {
		return nil, "", name, "", err
	}

	accessKey, _ := meta.scwClient.GetAccessKey()
	if projectID, _, err := extractProjectID(d, meta); err == nil {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(meta.httpClient, region, accessKey, secretKey)
	if err != nil {
		return nil, "", "", "", err
	}
	return s3Client, scw.Region(region), name, outerID, err
}

func s3ClientForceRegion(d *schema.ResourceData, m interface{}, region string) (*s3.S3, error) {
	meta := m.(*Meta)

	accessKey, _ := meta.scwClient.GetAccessKey()
	if projectID, _, err := extractProjectID(d, meta); err == nil {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(meta.httpClient, region, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	return s3Client, err
}

func accessKeyWithProjectID(accessKey string, projectID string) string {
	return accessKey + "@" + projectID
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

func objectBucketAPIEndpointURL(region scw.Region) string {
	return fmt.Sprintf("https://s3.%s.scw.cloud", region)
}

// Returns true if the error matches all these conditions:
//   - err is of type aws err.Error
//   - Error.Code() matches code
//   - Error.Message() contains message
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

func expandBucketCORS(ctx context.Context, rawCors []interface{}, bucket string) []*s3.CORSRule {
	rules := make([]*s3.CORSRule, 0, len(rawCors))
	for _, cors := range rawCors {
		corsMap := cors.(map[string]interface{})
		r := &s3.CORSRule{}
		for k, v := range corsMap {
			tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, put CORS: %#v, %#v", bucket, k, v))
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
	var globalErr error
	listInput := &s3.ListObjectVersionsInput{
		Bucket: scw.StringPtr(bucketName),
	}

	deletionWorkers := runtime.NumCPU()
	if deletionWorkers > maxObjectVersionDeletionWorkers {
		deletionWorkers = maxObjectVersionDeletionWorkers
	}

	listErr := conn.ListObjectVersionsPagesWithContext(ctx, listInput, func(page *s3.ListObjectVersionsOutput, _ bool) bool {
		pool := workerpool.NewWorkerPool(deletionWorkers)

		for _, objectVersion := range page.Versions {
			objectVersion := objectVersion

			pool.AddTask(func() error {
				objectKey := aws.StringValue(objectVersion.Key)
				objectVersionID := aws.StringValue(objectVersion.VersionId)
				err := deleteS3ObjectVersion(conn, bucketName, objectKey, objectVersionID, force)

				if isS3Err(err, ErrCodeAccessDenied, "") && force {
					legalHoldRemoved, errLegal := removeS3ObjectVersionLegalHold(conn, bucketName, objectVersion)
					if errLegal != nil {
						return fmt.Errorf("failed to remove legal hold: %s", errLegal)
					}

					if legalHoldRemoved {
						err = deleteS3ObjectVersion(conn, bucketName, objectKey, objectVersionID, force)
					}
				}

				if err != nil {
					return fmt.Errorf("failed to delete S3 object: %s", err)
				}

				return nil
			})
		}

		errs := pool.CloseAndWait()
		if len(errs) > 0 {
			globalErr = multierror.Append(nil, errs...)
			return false
		}

		return true
	})
	if listErr != nil {
		return fmt.Errorf("error listing S3 objects: %s", globalErr)
	}
	if globalErr != nil {
		return globalErr
	}

	listErr = conn.ListObjectVersionsPagesWithContext(ctx, listInput, func(page *s3.ListObjectVersionsOutput, _ bool) bool {
		pool := workerpool.NewWorkerPool(deletionWorkers)

		for _, deleteMarkerEntry := range page.DeleteMarkers {
			deleteMarkerEntry := deleteMarkerEntry

			pool.AddTask(func() error {
				deleteMarkerKey := aws.StringValue(deleteMarkerEntry.Key)
				deleteMarkerVersionsID := aws.StringValue(deleteMarkerEntry.VersionId)
				err := deleteS3ObjectVersion(conn, bucketName, deleteMarkerKey, deleteMarkerVersionsID, force)
				if err != nil {
					return fmt.Errorf("failed to delete S3 object delete marker: %s", err)
				}

				return nil
			})
		}

		errs := pool.CloseAndWait()
		if len(errs) > 0 {
			globalErr = multierror.Append(nil, errs...)
			return false
		}

		return true
	})
	if listErr != nil {
		return fmt.Errorf("error listing S3 objects for delete markers: %s", globalErr)
	}
	if globalErr != nil {
		return globalErr
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
		buf.WriteString(v.(string) + "-")
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

func SuppressEquivalentPolicyDiffs(k, old, newP string, _ *schema.ResourceData) bool {
	tflog.Debug(context.Background(),
		fmt.Sprintf("[DEBUG] suppress policy on key: %s, old: %s new: %s", k, old, newP))
	if strings.TrimSpace(old) == "" && strings.TrimSpace(newP) == "" {
		return true
	}

	if strings.TrimSpace(old) == "{}" && strings.TrimSpace(newP) == "" {
		return true
	}

	if strings.TrimSpace(old) == "" && strings.TrimSpace(newP) == "{}" {
		return true
	}

	if strings.TrimSpace(old) == "{}" && strings.TrimSpace(newP) == "{}" {
		return true
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(old, newP)
	if err != nil {
		return false
	}

	return equivalent
}

func SecondJSONUnlessEquivalent(old, newP string) (string, error) {
	// valid empty JSON is "{}" not "" so handle special case to avoid
	// Error unmarshalling policy: unexpected end of JSON input
	if strings.TrimSpace(newP) == "" {
		return "", nil
	}

	if strings.TrimSpace(newP) == "{}" {
		return "{}", nil
	}

	if strings.TrimSpace(old) == "" || strings.TrimSpace(old) == "{}" {
		return newP, nil
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(old, newP)
	if err != nil {
		return "", err
	}

	if equivalent {
		return old, nil
	}

	return newP, nil
}

type S3Website struct {
	Endpoint, Domain string
}

func WebsiteEndpoint(bucket string, region scw.Region) *S3Website {
	domain := WebsiteDomainURL(region.String())
	return &S3Website{Endpoint: fmt.Sprintf("%s.%s", bucket, domain), Domain: domain}
}

func WebsiteDomainURL(region string) string {
	// Different regions have different syntax for website endpoints
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_website_region_endpoints
	return fmt.Sprintf("s3-website.%s.scw.cloud", region)
}

func buildBucketOwnerID(id *string) *string {
	s := fmt.Sprintf("%[1]s:%[1]s", *id)
	return &s
}

func normalizeOwnerID(id *string) *string {
	tab := strings.Split(*id, ":")
	if len(tab) != 2 {
		return id
	}

	return &tab[0]
}

func addReadBucketErrorDiagnostic(diags *diag.Diagnostics, err error, resource string, awsResourceNotFoundCode string) (bucketFound bool, resourceFound bool) {
	switch {
	case isS3Err(err, s3.ErrCodeNoSuchBucket, ""):
		*diags = append(*diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Bucket not found",
			Detail:   "Got 404 error while reading bucket, removing from state",
		})
		return false, false

	case isS3Err(err, awsResourceNotFoundCode, ""):
		return true, false

	case isS3Err(err, ErrCodeAccessDenied, ""):
		d := diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Cannot read bucket %s: Forbidden", resource),
			Detail:   fmt.Sprintf("Got 403 error while reading bucket %s, please check your IAM permissions and your bucket policy", resource),
		}

		attributes := map[string]string{
			"acl":                       "acl",
			"object lock configuration": "object_lock_enabled",
			"objects":                   "",
			"tags":                      "tags",
			"CORS configuration":        "cors_rule",
			"versioning":                "versioning",
			"lifecycle configuration":   "lifecycle_rule",
		}
		if attributeName, ok := attributes[resource]; ok {
			d.AttributePath = cty.GetAttrPath(attributeName)
		}

		*diags = append(*diags, d)
		return true, true

	default:
		*diags = append(*diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Errorf("couldn't read bucket %s: %w", resource, err).Error(),
		})
		return true, true
	}
}
