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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultObjectBucketTimeout = 10 * time.Minute
	retryOnAWSAPI              = 2 * time.Minute
)

type scalewayS3EndpointResolver struct {
	Region string
}

func (r *scalewayS3EndpointResolver) ResolveEndpoint(service, region string, options ...interface{}) (aws.Endpoint, error) {
	return aws.Endpoint{
		URL:               "https://s3." + region + ".scw.cloud",
		HostnameImmutable: true,
		SigningRegion:     region,
		Source:            aws.EndpointSourceCustom,
	}, nil
}

func newS3Client(ctx context.Context, httpClient *http.Client, region, accessKey, secretKey string) (*s3.Client, error) {
	config, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awsconfig.WithEndpointResolverWithOptions(&scalewayS3EndpointResolver{Region: region}),
		awsconfig.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}
	if strings.ToLower(os.Getenv("TF_LOG")) == "debug" {
		config.ClientLogMode = aws.LogRequestWithBody | aws.LogRetries
	}

	return s3.NewFromConfig(config), nil
}

func newS3ClientFromMeta(meta *Meta) (*s3.Client, error) {
	region, _ := meta.scwClient.GetDefaultRegion()
	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()
	return newS3Client(context.Background(), meta.httpClient, region.String(), accessKey, secretKey)
}

func s3ClientWithRegion(ctx context.Context, d *schema.ResourceData, m interface{}) (*s3.Client, scw.Region, error) {
	meta := m.(*Meta)
	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()

	s3Client, err := newS3Client(ctx, meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", err
	}

	return s3Client, region, err
}

func s3ClientWithRegionAndName(ctx context.Context, m interface{}, name string) (*s3.Client, scw.Region, string, error) {
	meta := m.(*Meta)
	region, name, err := parseRegionalID(name)
	if err != nil {
		return nil, "", name, err
	}
	accessKey, _ := meta.scwClient.GetAccessKey()
	secretKey, _ := meta.scwClient.GetSecretKey()
	s3Client, err := newS3Client(ctx, meta.httpClient, region.String(), accessKey, secretKey)
	if err != nil {
		return nil, "", "", err
	}
	return s3Client, region, name, err
}

func flattenObjectBucketTags(tagsSet []s3types.Tag) map[string]interface{} {
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

func expandObjectBucketTags(tags interface{}) []s3types.Tag {
	tagsSet := []s3types.Tag(nil)
	for key, value := range tags.(map[string]interface{}) {
		tagsSet = append(tagsSet, s3types.Tag{
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
//   - err is of type aws err.Error
//   - Error.ErrorCode() matches code
//   - Error.ErrorMessage() contains message
func isS3ErrCode(err error, code string, message string) bool {
	var awsErr smithy.APIError
	if errors.As(err, &awsErr) {
		return awsErr.ErrorCode() == code && strings.Contains(awsErr.ErrorMessage(), message)
	}
	return false
}

func isS3Err[E error](err error, awsErr E) bool {
	return errors.As(err, &awsErr)
}

func flattenObjectBucketVersioning(versioningResponse *s3.GetBucketVersioningOutput) []map[string]interface{} {
	vcl := []map[string]interface{}{{}}
	vcl[0]["enabled"] = versioningResponse.Status == s3types.BucketVersioningStatusEnabled
	return vcl
}

func expandObjectBucketVersioning(v []interface{}) *s3types.VersioningConfiguration {
	vc := &s3types.VersioningConfiguration{}
	vc.Status = s3types.BucketVersioningStatusSuspended
	if len(v) > 0 {
		if c := v[0].(map[string]interface{}); c["enabled"].(bool) {
			vc.Status = s3types.BucketVersioningStatusEnabled
		}
	}
	return vc
}

func flattenBucketCORS(cors *s3.GetBucketCorsOutput) []map[string]interface{} {
	corsRules := make([]map[string]interface{}, 0)
	if cors != nil && len(cors.CORSRules) > 0 {
		corsRules = make([]map[string]interface{}, 0, len(cors.CORSRules))
		for _, ruleObject := range cors.CORSRules {
			rule := make(map[string]interface{})
			rule["allowed_headers"] = flattenSliceString(ruleObject.AllowedHeaders)
			rule["allowed_methods"] = flattenSliceString(ruleObject.AllowedMethods)
			rule["allowed_origins"] = flattenSliceString(ruleObject.AllowedOrigins)
			// Both the "ExposeHeaders" and "MaxAgeSeconds" might not be set.
			if ruleObject.AllowedOrigins != nil {
				rule["expose_headers"] = flattenSliceString(ruleObject.ExposeHeaders)
			}
			rule["max_age_seconds"] = int(ruleObject.MaxAgeSeconds)
			corsRules = append(corsRules, rule)
		}
	}
	return corsRules
}

func expandBucketCORS(ctx context.Context, rawCors []interface{}, bucket string) []s3types.CORSRule {
	rules := make([]s3types.CORSRule, 0, len(rawCors))
	for _, cors := range rawCors {
		corsMap := cors.(map[string]interface{})
		r := s3types.CORSRule{}
		for k, v := range corsMap {
			tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, put CORS: %#v, %#v", bucket, k, v))
			if k == "max_age_seconds" {
				r.MaxAgeSeconds = int32(v.(int))
			} else {
				vMap := make([]string, len(v.([]interface{})))
				for i, vv := range v.([]interface{}) {
					if str, ok := vv.(string); ok {
						vMap[i] = str
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

func deleteS3ObjectVersion(ctx context.Context, conn *s3.Client, bucketName string, key string, versionID string, force bool) error {
	input := &s3.DeleteObjectInput{
		Bucket: scw.StringPtr(bucketName),
		Key:    scw.StringPtr(key),
	}
	if versionID != "" {
		input.VersionId = scw.StringPtr(versionID)
	}
	if force {
		input.BypassGovernanceRetention = force
	}

	_, err := conn.DeleteObject(ctx, input)
	return err
}

// removeS3ObjectVersionLegalHold remove legal hold from an ObjectVersion if it is on
// returns true if legal hold was removed
func removeS3ObjectVersionLegalHold(ctx context.Context, conn *s3.Client, bucketName string, objectVersion s3types.ObjectVersion) (bool, error) {
	objectHead, err := conn.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:    scw.StringPtr(bucketName),
		Key:       objectVersion.Key,
		VersionId: objectVersion.VersionId,
	})
	if err != nil {
		err = fmt.Errorf("failed to get S3 object meta data: %s", err)
		return false, err
	}
	if objectHead.ObjectLockLegalHoldStatus != s3types.ObjectLockLegalHoldStatusOn {
		return false, nil
	}
	_, err = conn.PutObjectLegalHold(ctx, &s3.PutObjectLegalHoldInput{
		Bucket:    scw.StringPtr(bucketName),
		Key:       objectVersion.Key,
		VersionId: objectVersion.VersionId,
		LegalHold: &s3types.ObjectLockLegalHold{
			Status: s3types.ObjectLockLegalHoldStatusOff,
		},
	})
	if err != nil {
		err = fmt.Errorf("failed to put S3 object legal hold: %s", err)
		return false, err
	}
	return true, nil
}

func deleteS3ObjectVersions(ctx context.Context, conn *s3.Client, bucketName string, force bool) error {
	listInput := &s3.ListObjectVersionsInput{
		Bucket: scw.StringPtr(bucketName),
	}
	objectVersionsPaginator := NewListObjectVersionsPaginator(listInput)
	for objectVersionsPaginator.HasMorePages() {
		page, err := objectVersionsPaginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get object version page: %w", err)
		}
		for _, objectVersion := range page.Versions {
			objectKey := objectVersion.Key
			objectVersionID := objectVersion.VersionId
			err = deleteS3ObjectVersion(ctx, conn, bucketName, *objectKey, *objectVersionID, force)

			if isS3ErrCode(err, ErrCodeAccessDenied, "") && force {
				legalHoldRemoved, errLegal := removeS3ObjectVersionLegalHold(ctx, conn, bucketName, objectVersion)
				if errLegal != nil {
					return fmt.Errorf("failed to remove legal hold: %s", errLegal)
				}
				if legalHoldRemoved {
					err = deleteS3ObjectVersion(ctx, conn, bucketName, *objectKey, *objectVersionID, force)
				}
			}
			if err != nil {
				return fmt.Errorf("failed to delete S3 object: %s", err)
			}
		}
	}
	objectVersionsPaginator = NewListObjectVersionsPaginator(listInput)
	for objectVersionsPaginator.HasMorePages() {
		page, err := objectVersionsPaginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get object version page: %w", err)
		}
		for _, deleteMarkerEntry := range page.DeleteMarkers {
			deleteMarkerKey := deleteMarkerEntry.Key
			deleteMarkerVersionsID := deleteMarkerEntry.VersionId
			err = deleteS3ObjectVersion(ctx, conn, bucketName, *deleteMarkerKey, *deleteMarkerVersionsID, force)

			if err != nil {
				return fmt.Errorf("failed to delete S3 object delete marker: %s", err)
			}
		}
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

func retryOnAWSError[E error, O any](ctx context.Context, awsErr E, f func() (O, error)) (O, error) {
	var resp O
	err := resource.RetryContext(ctx, retryOnAWSAPI, func() *resource.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			if errors.As(err, &awsErr) {
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

func retryOnAWSCode[O any](ctx context.Context, code string, f func() (O, error)) (O, error) {
	var resp O
	err := resource.RetryContext(ctx, retryOnAWSAPI, func() *resource.RetryError {
		var err error
		resp, err = f()

		var ae smithy.APIError
		if err != nil {
			if errors.As(err, &ae) && ae.ErrorCode() == code {
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

func SuppressEquivalentPolicyDiffs(k, old, newP string, d *schema.ResourceData) bool {
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

func resourceBucketWebsiteConfigurationWebsiteEndpoint(ctx context.Context, conn *s3.Client, bucket string, region scw.Region) (*S3Website, error) {
	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	}

	output, err := conn.GetBucketLocation(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error getting Object Bucket (%s) Location: %w", bucket, err)
	}

	if output.LocationConstraint != "" {
		region = scw.Region(output.LocationConstraint)
	}

	return WebsiteEndpoint(bucket, region), nil
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
