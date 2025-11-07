package object

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/workerpool"
)

const (
	defaultObjectBucketTimeout = 10 * time.Minute

	maxObjectVersionDeletionWorkers = 8

	ErrCodeForbidden = "Forbidden"
)

type scalewayResolver struct {
	region string
}

func (r *scalewayResolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	return s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

func newS3Client(ctx context.Context, region, accessKey, secretKey string, httpClient *http.Client) (*s3.Client, error) {
	endpoint := "https://s3." + region + ".scw.cloud"

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.EndpointResolverV2 = &scalewayResolver{region: region}
	})

	return client, nil
}

func NewS3ClientFromMeta(ctx context.Context, meta *meta.Meta, region string) (*s3.Client, error) {
	accessKey, _ := meta.ScwClient().GetAccessKey()
	secretKey, _ := meta.ScwClient().GetSecretKey()

	projectID, _ := meta.ScwClient().GetDefaultProjectID()
	if projectID != "" {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}

	if region == "" {
		defaultRegion, _ := meta.ScwClient().GetDefaultRegion()
		region = defaultRegion.String()
	}

	return newS3Client(ctx, region, accessKey, secretKey, meta.HTTPClient())
}

func s3ClientWithRegion(ctx context.Context, d *schema.ResourceData, m any) (*s3.Client, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	accessKey, _ := meta.ExtractScwClient(m).GetAccessKey()
	if projectID, _, err := meta.ExtractProjectID(d, m); err == nil {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}

	secretKey, _ := meta.ExtractScwClient(m).GetSecretKey()

	s3Client, err := newS3Client(ctx, region.String(), accessKey, secretKey, meta.ExtractHTTPClient(m))
	if err != nil {
		return nil, "", err
	}

	return s3Client, region, err
}

func s3ClientWithRegionAndName(ctx context.Context, d *schema.ResourceData, m any, id string) (*s3.Client, scw.Region, string, error) {
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

	accessKey, _ := meta.ExtractScwClient(m).GetAccessKey()
	secretKey, _ := meta.ExtractScwClient(m).GetSecretKey()

	if len(parts) == 2 {
		accessKey = accessKeyWithProjectID(accessKey, parts[1])
	} else {
		projectID, _, err := meta.ExtractProjectID(d, m)
		if err == nil {
			accessKey = accessKeyWithProjectID(accessKey, projectID)
		}
	}

	s3Client, err := newS3Client(ctx, region.String(), accessKey, secretKey, meta.ExtractHTTPClient(m))
	if err != nil {
		return nil, "", "", err
	}

	return s3Client, region, name, nil
}

func s3ClientWithRegionAndNestedName(ctx context.Context, d *schema.ResourceData, m any, name string) (*s3.Client, scw.Region, string, string, error) {
	region, outerID, innerID, err := regional.ParseNestedID(name)
	if err != nil {
		return nil, "", "", "", err
	}

	accessKey, _ := meta.ExtractScwClient(m).GetAccessKey()
	if projectID, _, err := meta.ExtractProjectID(d, m); err == nil {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}

	secretKey, _ := meta.ExtractScwClient(m).GetSecretKey()

	s3Client, err := newS3Client(ctx, region.String(), accessKey, secretKey, meta.ExtractHTTPClient(m))
	if err != nil {
		return nil, "", "", "", err
	}

	return s3Client, region, outerID, innerID, err
}

func s3ClientWithRegionWithNameACL(ctx context.Context, d *schema.ResourceData, m any, name string) (*s3.Client, scw.Region, string, string, error) {
	region, name, outerID, err := locality.ParseLocalizedNestedOwnerID(name)
	if err != nil {
		return nil, "", name, "", err
	}

	accessKey, _ := meta.ExtractScwClient(m).GetAccessKey()
	if projectID, _, err := meta.ExtractProjectID(d, m); err == nil {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}

	secretKey, _ := meta.ExtractScwClient(m).GetSecretKey()

	s3Client, err := newS3Client(ctx, region, accessKey, secretKey, meta.ExtractHTTPClient(m))
	if err != nil {
		return nil, "", "", "", err
	}

	return s3Client, scw.Region(region), name, outerID, err
}

func s3ClientForceRegion(ctx context.Context, d *schema.ResourceData, m any, region string) (*s3.Client, error) {
	accessKey, _ := meta.ExtractScwClient(m).GetAccessKey()
	if projectID, _, err := meta.ExtractProjectID(d, m); err == nil {
		accessKey = accessKeyWithProjectID(accessKey, projectID)
	}

	secretKey, _ := meta.ExtractScwClient(m).GetSecretKey()

	s3Client, err := newS3Client(ctx, region, accessKey, secretKey, meta.ExtractHTTPClient(m))
	if err != nil {
		return nil, err
	}

	return s3Client, err
}

func accessKeyWithProjectID(accessKey string, projectID string) string {
	return accessKey + "@" + projectID
}

func flattenObjectBucketTags(tagsSet []s3Types.Tag) map[string]any {
	tags := map[string]any{}

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

func ExpandObjectBucketTags(tags any) []s3Types.Tag {
	tagsSet := make([]s3Types.Tag, 0, len(tags.(map[string]any)))
	for key, value := range tags.(map[string]any) {
		tagsSet = append(tagsSet, s3Types.Tag{
			Key:   scw.StringPtr(key),
			Value: types.ExpandStringPtr(value),
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

// IsS3Err returns true if the error matches all these conditions:
//   - err is of type aws err.Error
//   - Error.Code() matches code
//   - Error.Message() contains message
func IsS3Err(err error, code string, message string) bool {
	var awsErr smithy.APIError
	if errors.As(err, &awsErr) {
		return awsErr.ErrorCode() == code && strings.Contains(awsErr.ErrorMessage(), message)
	}

	return false
}

func flattenObjectBucketVersioning(versioningResponse *s3.GetBucketVersioningOutput) []map[string]any {
	vcl := []map[string]any{{}}
	if versioningResponse != nil {
		vcl[0]["enabled"] = versioningResponse.Status == s3Types.BucketVersioningStatusEnabled

		return vcl
	}

	vcl[0]["enabled"] = false

	return vcl
}

func expandObjectBucketVersioning(v []any) *s3Types.VersioningConfiguration {
	vc := &s3Types.VersioningConfiguration{}
	vc.Status = s3Types.BucketVersioningStatusSuspended

	if len(v) > 0 {
		if c := v[0].(map[string]any); c["enabled"].(bool) {
			vc.Status = s3Types.BucketVersioningStatusEnabled
		}
	}

	return vc
}

func flattenBucketCORS(corsResponse any) []any {
	if corsResponse == nil {
		return nil
	}

	if cors, ok := corsResponse.(*s3.GetBucketCorsOutput); ok && cors != nil && len(cors.CORSRules) > 0 {
		var corsRules []any

		for _, ruleObject := range cors.CORSRules {
			rule := map[string]any{}
			if len(ruleObject.AllowedHeaders) > 0 {
				rule["allowed_headers"] = ruleObject.AllowedHeaders
			}

			if len(ruleObject.AllowedMethods) > 0 {
				rule["allowed_methods"] = ruleObject.AllowedMethods
			}

			if len(ruleObject.AllowedOrigins) > 0 {
				rule["allowed_origins"] = ruleObject.AllowedOrigins
			}

			if len(ruleObject.ExposeHeaders) > 0 {
				rule["expose_headers"] = ruleObject.ExposeHeaders
			}

			if ruleObject.MaxAgeSeconds != nil {
				rule["max_age_seconds"] = ruleObject.MaxAgeSeconds
			}

			corsRules = append(corsRules, rule)
		}

		return corsRules
	}

	return nil
}

func expandBucketCORS(ctx context.Context, rawCors []any, bucket string) []s3Types.CORSRule {
	if len(rawCors) == 0 {
		tflog.Warn(ctx, "No CORS configuration provided for bucket: "+bucket)

		return nil
	}

	// Preallocate memory for rules
	rules := make([]s3Types.CORSRule, 0, len(rawCors))

	for _, raw := range rawCors {
		corsMap, ok := raw.(map[string]any)
		if !ok {
			tflog.Warn(ctx, fmt.Sprintf("Invalid CORS entry for bucket %s: %#v", bucket, raw))

			continue
		}

		rule := s3Types.CORSRule{}

		for key, value := range corsMap {
			tflog.Debug(ctx, fmt.Sprintf("Processing CORS key: %s, value: %#v for bucket %s", key, value, bucket))

			switch key {
			case "allowed_headers":
				rule.AllowedHeaders = toStringSlice(ctx, value)
			case "allowed_methods":
				rule.AllowedMethods = toStringSlice(ctx, value)
			case "allowed_origins":
				rule.AllowedOrigins = toStringSlice(ctx, value)
			case "expose_headers":
				rule.ExposeHeaders = toStringSlice(ctx, value)
			case "max_age_seconds":
				if maxAge, ok := value.(int); ok {
					rule.MaxAgeSeconds = scw.Int32Ptr(int32(maxAge))
				} else {
					tflog.Warn(ctx, fmt.Sprintf("Invalid type for max_age_seconds in bucket %s: %T", bucket, value))
				}
			default:
				tflog.Warn(ctx, fmt.Sprintf("Unknown key in CORS configuration for bucket %s: %s", bucket, key))
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

func toStringSlice(ctx context.Context, input any) []string {
	var result []string

	switch v := input.(type) {
	case []any:
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
	case []string:
		return v
	default:
		tflog.Warn(ctx, fmt.Sprintf("Unexpected type for toStringSlice: %T", input))
	}

	return result
}

func deleteS3ObjectVersion(ctx context.Context, conn *s3.Client, bucketName string, key string, versionID string, _ bool) error {
	input := &s3.DeleteObjectInput{
		Bucket: scw.StringPtr(bucketName),
		Key:    scw.StringPtr(key),
	}
	if versionID != "" {
		input.VersionId = scw.StringPtr(versionID)
	}

	_, err := conn.DeleteObject(ctx, input)

	return err
}

// removeS3ObjectVersionLegalHold remove legal hold from an ObjectVersion if it is on
// returns true if legal hold was removed
func removeS3ObjectVersionLegalHold(ctx context.Context, conn *s3.Client, bucketName string, objectVersion *s3Types.ObjectVersion) (bool, error) {
	objectHead, err := conn.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:    scw.StringPtr(bucketName),
		Key:       objectVersion.Key,
		VersionId: objectVersion.VersionId,
	})
	if err != nil {
		err = fmt.Errorf("failed to get S3 object meta data: %w", err)

		return false, err
	}

	if objectHead.ObjectLockLegalHoldStatus != s3Types.ObjectLockLegalHoldStatusOn {
		return false, nil
	}

	_, err = conn.PutObjectLegalHold(ctx, &s3.PutObjectLegalHoldInput{
		Bucket:    scw.StringPtr(bucketName),
		Key:       objectVersion.Key,
		VersionId: objectVersion.VersionId,
		LegalHold: &s3Types.ObjectLockLegalHold{
			Status: s3Types.ObjectLockLegalHoldStatusOff,
		},
	})
	if err != nil {
		err = fmt.Errorf("failed to put S3 object legal hold: %w", err)

		return false, err
	}

	return true, nil
}

func emptyBucket(ctx context.Context, conn *s3.Client, bucketName string, force bool) (int64, error) {
	var nObject int64

	var nObjectMarker int64

	var globalErr error

	// Delete All Object Version
	nObject, err := processAllPagesObject(ctx, bucketName, conn, force, deleteVersionBucket)
	if err != nil {
		globalErr = multierror.Append(globalErr, err)
	}

	// Delete All Object Marker
	nObjectMarker, err = processAllPagesObject(ctx, bucketName, conn, force, deleteMarkerBucket)
	if err != nil {
		globalErr = multierror.Append(globalErr, err)
	}

	nObject += nObjectMarker

	return nObject, globalErr
}

func processAllPagesObject(ctx context.Context, bucketName string, conn *s3.Client, force bool, fn func(ctx context.Context, conn *s3.Client, bucket string, force bool, page *s3.ListObjectVersionsOutput, pool *workerpool.WorkerPool) (int64, error)) (int64, error) {
	var globalErr error

	deletionWorkers := findDeletionWorkerCapacity()
	nObject := int64(0)
	input := &s3.ListObjectVersionsInput{
		Bucket: scw.StringPtr(bucketName),
	}
	pages := s3.NewListObjectVersionsPaginator(conn, input)
	pool := workerpool.NewWorkerPool(deletionWorkers)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nObject, fmt.Errorf("error listing S3 objects: %w", err)
		}

		n, taskErr := fn(ctx, conn, bucketName, force, page, pool)
		if taskErr != nil {
			return nObject, taskErr
		}

		nObject += n
	}

	if errs := pool.CloseAndWait(); errs != nil {
		globalErr = multierror.Append(nil, errs...)

		return nObject, globalErr
	}

	return nObject, nil
}

func deleteMarkerBucket(ctx context.Context, conn *s3.Client, bucketName string, force bool, page *s3.ListObjectVersionsOutput, pool *workerpool.WorkerPool) (int64, error) {
	var nObject int64

	for _, deleteMarkerEntry := range page.DeleteMarkers {
		pool.AddTask(func() error {
			deleteMarkerKey := aws.ToString(deleteMarkerEntry.Key)
			deleteMarkerVersionsID := aws.ToString(deleteMarkerEntry.VersionId)

			err := deleteS3ObjectVersion(ctx, conn, bucketName, deleteMarkerKey, deleteMarkerVersionsID, force)
			if err != nil {
				return fmt.Errorf("failed to delete S3 object delete marker: %w", err)
			}

			nObject++

			return nil
		})
	}

	return nObject, nil
}

func deleteVersionBucket(ctx context.Context, conn *s3.Client, bucketName string, force bool, page *s3.ListObjectVersionsOutput, pool *workerpool.WorkerPool) (int64, error) {
	var nObject int64

	for _, objectVersion := range page.Versions {
		pool.AddTask(func() error {
			objectKey := aws.ToString(objectVersion.Key)
			objectVersionID := aws.ToString(objectVersion.VersionId)
			err := deleteS3ObjectVersion(ctx, conn, bucketName, objectKey, objectVersionID, force)

			if IsS3Err(err, ErrCodeAccessDenied, "") && force {
				legalHoldRemoved, errLegal := removeS3ObjectVersionLegalHold(ctx, conn, bucketName, &objectVersion)
				if errLegal != nil {
					return fmt.Errorf("failed to remove legal hold: %w", errLegal)
				}

				if legalHoldRemoved {
					err = deleteS3ObjectVersion(ctx, conn, bucketName, objectKey, objectVersionID, force)
				}
			}

			nObject++

			if err != nil {
				return fmt.Errorf("failed to delete S3 object: %w", err)
			}

			return nil
		})
	}

	return nObject, nil
}

func findDeletionWorkerCapacity() int {
	deletionWorkers := runtime.NumCPU()
	if deletionWorkers > maxObjectVersionDeletionWorkers {
		deletionWorkers = maxObjectVersionDeletionWorkers
	}

	return deletionWorkers
}

func transitionHash(v any) int {
	var buf bytes.Buffer

	m, ok := v.(map[string]any)

	if !ok {
		return 0
	}

	if v, ok := m["days"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}

	if v, ok := m["storage_class"]; ok {
		buf.WriteString(v.(string) + "-")
	}

	return types.StringHashcode(buf.String())
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

func NormalizeOwnerID(id *string) *string {
	tab := strings.Split(*id, ":")
	if len(tab) != 2 {
		return id
	}

	return &tab[0]
}

func addReadBucketErrorDiagnostic(diags *diag.Diagnostics, err error, resource string, awsResourceNotFoundCode string) (bucketFound bool, resourceFound bool) {
	switch {
	case errors.As(err, new(*s3Types.NoSuchBucket)):
		*diags = append(*diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Bucket not found",
			Detail:   "Got 404 error while reading bucket, removing from state",
		})

		return false, false

	case IsS3Err(err, awsResourceNotFoundCode, ""):
		return true, false

	case IsS3Err(err, ErrCodeAccessDenied, ""):
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
