package object

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceObjectBucketCreate,
		ReadContext:   resourceObjectBucketRead,
		UpdateContext: resourceObjectBucketUpdate,
		DeleteContext: resourceObjectBucketDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultObjectBucketTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc:    bucketSchema,
		Identity:      identity.DefaultRegional(),
		CustomizeDiff: validateBucket,
	}
}

func bucketSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The name of the bucket",
		},
		"object_lock_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
			Description: "Enable object lock",
		},
		"acl": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "private",
			Description: "ACL of the bucket: either 'private', 'public-read', 'public-read-write' or 'authenticated-read'.",
			ValidateFunc: validation.StringInSlice([]string{
				string(s3Types.BucketCannedACLPrivate),
				string(s3Types.BucketCannedACLPublicRead),
				string(s3Types.BucketCannedACLPublicReadWrite),
				string(s3Types.BucketCannedACLAuthenticatedRead),
			}, false),
			Deprecated: "ACL attribute is deprecated. Please use the resource scaleway_object_bucket_acl instead.",
		},
		"tags": {
			Type: schema.TypeMap,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "The tags associated with this bucket",
		},
		"endpoint": {
			Type:        schema.TypeString,
			Description: "Endpoint of the bucket",
			Computed:    true,
		},
		"api_endpoint": {
			Type:        schema.TypeString,
			Description: "API URL of the bucket",
			Computed:    true,
		},
		"cors_rule": {
			Type:        schema.TypeList,
			Description: "List of CORS rules",
			Optional:    true,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"allowed_headers": {
						Type:        schema.TypeList,
						Description: "Allowed headers in the CORS rule",
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"allowed_methods": {
						Type:        schema.TypeList,
						Description: "Allowed HTTP methods allowed in the CORS rule",
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"allowed_origins": {
						Type:        schema.TypeList,
						Description: "Allowed origins allowed in the CORS rule",
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"expose_headers": {
						Type:        schema.TypeList,
						Description: "Exposed headers in the CORS rule",
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"max_age_seconds": {
						Type:        schema.TypeInt,
						Description: "Max age of the CORS rule",
						Optional:    true,
					},
				},
			},
		},
		"force_destroy": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Delete objects in bucket",
		},
		"lifecycle_rule": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Lifecycle configuration is a set of rules that define actions that Scaleway Object Storage applies to a group of objects",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:         schema.TypeString,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.StringLenBetween(0, 255),
						Description:  "Unique identifier for the rule",
					},
					"prefix": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The prefix identifying one or more objects to which the rule applies",
					},
					"tags": {
						Type: schema.TypeMap,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Optional:    true,
						Description: "The tags associated with the bucket lifecycle",
					},
					"object_size_greater_than": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Minimum object size (in bytes) to which the rule applies",
					},
					"object_size_less_than": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Maximum object size (in bytes) to which the rule applies",
					},
					"enabled": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Specifies if the configuration rule is Enabled or Disabled",
					},
					"abort_incomplete_multipart_upload_days": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "Specifies the number of days after initiating a multipart upload when the multipart upload must be completed",
					},
					"expiration": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Specifies a period in the object's expire",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"date": {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: validBucketLifecycleTimestamp,
									Description:  "Specifies the date the object is to be moved or deleted. The date value must be in RFC3339 full-date format e.g. `2023-08-22`",
								},
								"days": {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntAtLeast(1),
									Description:  "Specifies the number of days after object creation when the specific rule action takes effect",
								},
								"expired_object_delete_marker": {
									Type:        schema.TypeBool,
									Optional:    true,
									Description: "Specifies whether Scaleway Object will remove a delete marker with no noncurrent versions. If set to `true`, the delete marker will be expired; if set to `false` the policy takes no action",
								},
							},
						},
					},
					"transition": {
						Type:        schema.TypeSet,
						Optional:    true,
						Set:         transitionHash,
						Description: "Define when objects transition to another storage class",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"date": {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: validBucketLifecycleTimestamp,
									Description:  "Specifies the date objects are transitioned to the specified storage class. The date value must be in RFC3339 full-date format e.g. `2023-08-22`",
								},
								"days": {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntAtLeast(0),
									Description:  "Specifies the number of days after object creation when the specific rule action takes effect",
								},
								"storage_class": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice(TransitionSCWStorageClassValues(), false),
									Description:  "Specifies the Scaleway Object Storage class to which you want the object to transition",
								},
							},
						},
					},
					"noncurrent_version_expiration": {
						Type:        schema.TypeList,
						MaxItems:    1,
						Optional:    true,
						Description: "Configuration block that specifies when noncurrent object versions expire",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"newer_noncurrent_versions": {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntBetween(1, 100),
									Description:  "Number of noncurrent versions Scaleway Object Storage will retain. Must be a non-zero positive integer",
								},
								"noncurrent_days": {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntAtLeast(1),
									Description:  "Number of days an object is noncurrent before Scaleway Object Storage can perform the associated action. Must be a positive integer",
								},
							},
						},
					},
					"noncurrent_version_transition": {
						Type:        schema.TypeSet,
						Optional:    true,
						Description: "Set of configuration blocks that specify the transition rule for the lifecycle rule that describes when noncurrent objects transition to a specific storage class",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"newer_noncurrent_versions": {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntBetween(1, 100),
									Description:  "Number of noncurrent versions Scaleway Object Storage will retain. Must be a non-zero positive integer",
								},
								"noncurrent_days": {
									Type:         schema.TypeInt,
									Required:     true,
									ValidateFunc: validation.IntAtLeast(1),
									Description:  "Number of days an object is noncurrent before Scaleway Object Storage can perform the associated action",
								},
								"storage_class": {
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice(TransitionSCWStorageClassValues(), false),
									Description:  "Specifies the Scaleway Object Storage class to which you want the object to transition",
								},
							},
						},
					},
				},
			},
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"versioning": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Description: "Allow multiple versions of an object in the same bucket",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Description: "Enable versioning. Once you version-enable a bucket, it can never return to an unversioned state",
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
}

/*
*** CREATE
 */

func resourceObjectBucketCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("name").(string)

	input := &s3.CreateBucketInput{
		Bucket: new(bucketName),
	}

	if v, ok := d.GetOk("object_lock_enabled"); ok {
		input.ObjectLockEnabledForBucket = new(v.(bool))
	}

	_, err = s3Client.CreateBucket(ctx, input)

	if v, ok := d.GetOk("acl"); ok {
		input.ACL = s3Types.BucketCannedACL(v.(string))
	}

	if TimedOut(err) {
		_, err = s3Client.CreateBucket(ctx, input)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	projectId := d.Get("project_id").(string)

	if projectId != "" {
		err = identity.SetRegionalIdentity(d, region, bucketName+"@"+projectId)
	} else {
		err = identity.SetRegionalIdentity(d, region, bucketName)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	tagsSet := ExpandObjectBucketTags(d.Get("tags"))

	if len(tagsSet) > 0 {
		_, err = s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
			Bucket: new(bucketName),
			Tagging: &s3Types.Tagging{
				TagSet: tagsSet,
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceObjectBucketUpdate(ctx, d, m)
}

/*
*** UPDATE
 */

func resourceObjectBucketUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("acl") {
		acl := d.Get("acl").(string)

		_, err := s3Client.PutBucketAcl(ctx, &s3.PutBucketAclInput{
			Bucket: new(bucketName),
			ACL:    s3Types.BucketCannedACL(acl),
		})
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Couldn't update bucket ACL: %s", err))

			return diag.FromErr(fmt.Errorf("couldn't update bucket ACL: %w", err))
		}
	}

	// Object Lock enables versioning so we don't want to update versioning it is enabled
	objectLockEnabled := d.Get("object_lock_enabled").(bool)
	if !objectLockEnabled && d.HasChange("versioning") {
		if err := resourceObjectBucketVersioningUpdate(ctx, s3Client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		tagsSet := ExpandObjectBucketTags(d.Get("tags"))

		if len(tagsSet) > 0 {
			_, err = s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
				Bucket: new(bucketName),
				Tagging: &s3Types.Tagging{
					TagSet: tagsSet,
				},
			})
		} else {
			_, err = s3Client.DeleteBucketTagging(ctx, &s3.DeleteBucketTaggingInput{
				Bucket: new(bucketName),
			})
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("cors_rule") {
		if err := resourceS3BucketCorsUpdate(ctx, s3Client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("lifecycle_rule") {
		if err := resourceBucketLifecycleUpdate(ctx, s3Client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceObjectBucketRead(ctx, d, m)
}

//gocyclo:ignore
func resourceBucketLifecycleUpdate(ctx context.Context, conn *s3.Client, d *schema.ResourceData) error {
	bucket := d.Get("name").(string)

	lifecycleRules := d.Get("lifecycle_rule").([]any)

	if len(lifecycleRules) == 0 || lifecycleRules[0] == nil {
		i := &s3.DeleteBucketLifecycleInput{
			Bucket: aws.String(bucket),
		}

		_, err := conn.DeleteBucketLifecycle(ctx, i)
		if err != nil {
			return fmt.Errorf("error removing S3 lifecycle for bucket %s: %w", bucket, err)
		}

		return nil
	}

	rules := make([]s3Types.LifecycleRule, 0, len(lifecycleRules))

	for i, lifecycleRule := range lifecycleRules {
		r := lifecycleRule.(map[string]any)

		rule := s3Types.LifecycleRule{}

		// ID
		if val, ok := r["id"].(string); ok && val != "" {
			rule.ID = aws.String(val)
		} else {
			rule.ID = aws.String(id.PrefixedUniqueId("tf-scw-bucket-lifecycle-"))
		}

		// Filter
		rule.Filter = extractFilter(r, d)

		// Enabled
		if val, ok := r["enabled"].(bool); ok && val {
			rule.Status = s3Types.ExpirationStatusEnabled
		} else {
			rule.Status = s3Types.ExpirationStatusDisabled
		}

		// AbortIncompleteMultipartUpload
		if val, ok := r["abort_incomplete_multipart_upload_days"].(int); ok && val > 0 {
			days := int32(val)
			rule.AbortIncompleteMultipartUpload = &s3Types.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int32(days),
			}
		}

		// Expiration
		expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.expiration", i)).([]any)
		if len(expiration) > 0 && expiration[0] != nil {
			e := expiration[0].(map[string]any)
			i := &s3Types.LifecycleExpiration{}

			if val, ok := e["days"].(int); ok && val > 0 {
				days := int32(val)
				i.Days = aws.Int32(days)
			}

			if val, ok := e["date"].(string); ok && val != "" {
				date, err := time.Parse("2006-01-02", val)
				if err != nil {
					return fmt.Errorf("error while parsing expiration date '%s': %w", val, err)
				}

				i.Date = aws.Time(date)
			}

			if val, ok := e["expired_object_delete_marker"].(bool); ok {
				i.ExpiredObjectDeleteMarker = aws.Bool(val)
			}

			rule.Expiration = i
		}

		// Transitions
		transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.transition", i)).(*schema.Set).List()
		if len(transitions) > 0 {
			rule.Transitions = []s3Types.Transition{}

			for _, transition := range transitions {
				transition := transition.(map[string]any)
				i := s3Types.Transition{}

				if val, ok := transition["days"].(int); ok && val >= 0 {
					days := int32(val)
					i.Days = aws.Int32(days)
				}

				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = s3Types.TransitionStorageClass(val)
				}

				if val, ok := transition["date"].(string); ok && val != "" {
					date, err := time.Parse(time.RFC3339, val)
					if err != nil {
						return fmt.Errorf("error while parsing transition date '%s': %w", date, err)
					}

					i.Date = aws.Time(date)
				}

				rule.Transitions = append(rule.Transitions, i)
			}
		}

		// NoncurrentVersionExpiration
		noncurrentVersionExpiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.noncurrent_version_expiration", i)).([]any)
		if len(noncurrentVersionExpiration) > 0 && noncurrentVersionExpiration[0] != nil {
			expiration := noncurrentVersionExpiration[0].(map[string]any)
			i := &s3Types.NoncurrentVersionExpiration{}

			if val, ok := expiration["noncurrent_days"].(int); ok && val > 0 {
				noncurrentDays := int32(val)
				i.NoncurrentDays = aws.Int32(noncurrentDays)
			}

			if val, ok := expiration["newer_noncurrent_versions"].(int); ok && val > 0 {
				newerNoncurrentVersions := int32(val)
				i.NewerNoncurrentVersions = aws.Int32(newerNoncurrentVersions)
			}

			rule.NoncurrentVersionExpiration = i
		}

		// NoncurrentVersionTransitions
		noncurrentVersionTransitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.noncurrent_version_transition", i)).(*schema.Set).List()
		if len(noncurrentVersionTransitions) > 0 {
			rule.NoncurrentVersionTransitions = []s3Types.NoncurrentVersionTransition{}

			for _, transition := range noncurrentVersionTransitions {
				transition := transition.(map[string]any)
				i := s3Types.NoncurrentVersionTransition{}

				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = s3Types.TransitionStorageClass(val)
				}

				if val, ok := transition["noncurrent_days"].(int); ok && val > 0 {
					noncurrentDays := int32(val)
					i.NoncurrentDays = aws.Int32(noncurrentDays)
				}

				if val, ok := transition["newer_noncurrent_versions"].(int); ok && val > 0 {
					newerNoncurrentVersions := int32(val)
					i.NewerNoncurrentVersions = aws.Int32(newerNoncurrentVersions)
				}

				rule.NoncurrentVersionTransitions = append(rule.NoncurrentVersionTransitions, i)
			}
		}

		rules = append(rules, rule)
	}

	i := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3Types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}

	if _, err := conn.PutBucketLifecycleConfiguration(ctx, i); err != nil {
		return fmt.Errorf("error applying lifecycle configuration to bucket %s: %w", bucket, err)
	}

	return nil
}

func extractFilter(r map[string]any, resourceData *schema.ResourceData) *s3Types.LifecycleRuleFilter {
	prefix := r["prefix"].(string)
	tags := ExpandObjectBucketTags(r["tags"])
	objectSizeGreaterThan := r["object_size_greater_than"].(int)
	objectSizeLessThan := r["object_size_less_than"].(int)

	filterElements := []any{prefix, tags, objectSizeGreaterThan, objectSizeLessThan}

	fieldsCounter := 0
	for _, e := range filterElements {
		fieldsCounter += countFilters(e)
	}

	filter := &s3Types.LifecycleRuleFilter{}

	if fieldsCounter == 1 {
		// If a single filter field is set, put it in "filter"
		if len(tags) > 0 {
			filter.Tag = &tags[0]
		}

		if prefix != "" {
			filter.Prefix = new(r["prefix"].(string))
		}

		if objectSizeGreaterThan > 0 {
			filter.ObjectSizeGreaterThan = new(int64(objectSizeGreaterThan))
		}

		if objectSizeLessThan > 0 {
			filter.ObjectSizeLessThan = new(int64(objectSizeLessThan))
		}
	} else if fieldsCounter >= 2 {
		// If several fields are set, put them into "filter.and"
		lifecycleRuleAndOp := &s3Types.LifecycleRuleAndOperator{}

		if len(tags) > 0 {
			lifecycleRuleAndOp.Tags = tags
		}

		if prefix != "" {
			lifecycleRuleAndOp.Prefix = new(r["prefix"].(string))
		}

		if objectSizeGreaterThan > 0 {
			lifecycleRuleAndOp.ObjectSizeGreaterThan = new(int64(objectSizeGreaterThan))
		}

		if objectSizeLessThan > 0 {
			lifecycleRuleAndOp.ObjectSizeLessThan = new(int64(objectSizeLessThan))
		}

		filter.And = lifecycleRuleAndOp
	}

	return filter
}

// Return as many elements as represented by "i"
// This function was designed around the need to count each tag as separate filters
func countFilters(i any) int {
	switch v := i.(type) {
	case string:
		if v != "" {
			return 1
		}
	case int:
		if v > 0 {
			return 1
		}
	case []s3Types.Tag:
		return len(v)
	default:
		if v != nil {
			return 1
		}
	}

	return 0
}

func resourceObjectBucketVersioningUpdate(ctx context.Context, s3conn *s3.Client, d *schema.ResourceData) error {
	v := d.Get("versioning").([]any)
	bucketName := d.Get("name").(string)
	vc := expandObjectBucketVersioning(v)

	i := &s3.PutBucketVersioningInput{
		Bucket:                  new(bucketName),
		VersioningConfiguration: vc,
	}
	tflog.Debug(ctx, fmt.Sprintf("S3 put bucket versioning: %#v", i))

	_, err := s3conn.PutBucketVersioning(ctx, i)
	if err != nil {
		return fmt.Errorf("error putting S3 versioning: %w", err)
	}

	return nil
}

func resourceS3BucketCorsUpdate(ctx context.Context, s3conn *s3.Client, d *schema.ResourceData) error {
	bucketName := d.Get("name").(string)
	rawCors := d.Get("cors_rule").([]any)

	if len(rawCors) == 0 {
		// Delete CORS
		tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, delete CORS", bucketName))

		_, err := s3conn.DeleteBucketCors(ctx, &s3.DeleteBucketCorsInput{
			Bucket: new(bucketName),
		})
		if err != nil {
			return fmt.Errorf("error deleting S3 CORS: %w", err)
		}
	} else {
		// Put CORS
		rules := expandBucketCORS(ctx, rawCors, bucketName)
		corsInput := &s3.PutBucketCorsInput{
			Bucket: new(bucketName),
			CORSConfiguration: &s3Types.CORSConfiguration{
				CORSRules: rules,
			},
		}
		tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, put CORS: %#v", bucketName, corsInput))

		_, err := s3conn.PutBucketCors(ctx, corsInput)
		if err != nil {
			return fmt.Errorf("error putting S3 CORS: %w", err)
		}
	}

	return nil
}

/*
*** READ
 */

//gocyclo:ignore
func resourceObjectBucketRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, region, bucketName, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	_ = d.Set("name", bucketName)
	_ = d.Set("region", region)

	projectId, diags, ok := setProjectId(ctx, d, bucketName, s3Client, &diags)
	if !ok {
		return diags
	}

	if projectId != "" {
		err = identity.SetRegionalIdentity(d, region, bucketName+"@"+projectId)
	} else {
		err = identity.SetRegionalIdentity(d, region, bucketName)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Get object_lock_enabled
	objectLockConfiguration, err := s3Client.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		bucketFound, objectLockFound := addReadBucketErrorDiagnostic(&diags, err, "object lock configuration", ErrCodeObjectLockConfigurationNotFoundError)
		if !bucketFound {
			d.SetId("")

			return diags
		}

		if !objectLockFound {
			_ = d.Set("object_lock_enabled", false)
		}
	} else if objectLockConfiguration.ObjectLockConfiguration != nil {
		_ = d.Set("object_lock_enabled", true)
	}

	// We do not read `acl` attribute because it could be impossible to find
	// the right canned ACL from a complex ACL object.
	//
	// Known issue:
	// Import a bucket (eg. terraform import scaleway_object_bucket.x fr-par/x)
	// will always trigger a diff (eg. terraform plan) on acl attribute because
	// we do not read it, and it has a "private" default value.
	// AWS has the same issue: https://github.com/terraform-providers/terraform-provider-aws/issues/6193

	_, err = s3Client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: new(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "objects", ""); !bucketFound {
			d.SetId("")

			return diags
		}
	}

	var tagsSet []s3Types.Tag

	tagsResponse, err := s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: new(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "tags", ErrCodeNoSuchTagSet); !bucketFound {
			d.SetId("")

			return diags
		}
	} else {
		tagsSet = tagsResponse.TagSet
	}

	_ = d.Set("tags", flattenObjectBucketTags(tagsSet))
	_ = d.Set("endpoint", objectBucketEndpointURL(bucketName, region))
	_ = d.Set("api_endpoint", objectBucketAPIEndpointURL(region))

	// Read the CORS
	corsResponse, err := s3Client.GetBucketCors(ctx, &s3.GetBucketCorsInput{
		Bucket: new(bucketName),
	})

	if err != nil && !IsS3Err(err, ErrCodeNoSuchCORSConfiguration, "The CORS configuration does not exist") {
		return diag.FromErr(err)
	}

	_ = d.Set("cors_rule", flattenBucketCORS(corsResponse))

	// Read the versioning configuration
	versioningResponse, err := s3Client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: new(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "versioning", ""); !bucketFound {
			d.SetId("")

			return diags
		}
	}

	_ = d.Set("versioning", FlattenObjectBucketVersioning(versioningResponse))

	// Read the lifecycle configuration
	lifecycle, err := s3Client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: new(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "lifecycle configuration", ErrCodeNoSuchLifecycleConfiguration); !bucketFound {
			d.SetId("")

			return diags
		}
	}

	lifecycleRules := resourceBucketLifecycleRulesRead(lifecycle, d)

	if err := d.Set("lifecycle_rule", lifecycleRules); err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("error setting lifecycle_rule: %s", err),
		})
	}

	return diags
}

func resourceBucketLifecycleRulesRead(
	lifecycle *s3.GetBucketLifecycleConfigurationOutput, d *schema.ResourceData,
) []map[string]any {
	lifecycleRules := make([]map[string]any, 0)

	if lifecycle != nil && len(lifecycle.Rules) > 0 {
		lifecycleRules = make([]map[string]any, 0, len(lifecycle.Rules))

		for _, lifecycleRule := range lifecycle.Rules {
			log.Printf("[DEBUG] SCW bucket: %s, read lifecycle rule: %v", d.Id(), lifecycleRule)

			rule := make(map[string]any)

			// ID
			if lifecycleRule.ID != nil && aws.ToString(lifecycleRule.ID) != "" {
				rule["id"] = aws.ToString(lifecycleRule.ID)
			}

			// Filter
			resourceBucketLifecycleRulesFilterRead(lifecycleRule.Filter, rule)

			// Enabled
			if lifecycleRule.Status == s3Types.ExpirationStatusEnabled {
				rule["enabled"] = true
			} else {
				rule["enabled"] = false
			}

			// AbortIncompleteMultipartUploadDays
			if lifecycleRule.AbortIncompleteMultipartUpload != nil {
				if lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
					rule["abort_incomplete_multipart_upload_days"] = int(aws.ToInt32(
						lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation,
					))
				}
			}

			// Expiration
			resourceBucketLifecycleRulesExpirationRead(lifecycleRule.Expiration, rule)

			// Transitions
			resourceBucketLifecycleRulesTransitionsRead(lifecycleRule.Transitions, rule)

			// NonCurrentVersionExpiration
			resourceBucketLifecycleRulesNonCurrentVersionExpiration(lifecycleRule.NoncurrentVersionExpiration, rule)

			// NonCurrentVersionTransition
			resourceBucketLifecycleRulesNonCurrentVersionTransitions(lifecycleRule.NoncurrentVersionTransitions, rule)

			lifecycleRules = append(lifecycleRules, rule)
		}
	}

	return lifecycleRules
}

func resourceBucketLifecycleRulesFilterRead(filter *s3Types.LifecycleRuleFilter, rule map[string]any) {
	if filter == nil {
		return
	}

	if filter.And != nil {
		// Prefix
		if filter.And.Prefix != nil && aws.ToString(filter.And.Prefix) != "" {
			rule["prefix"] = aws.ToString(filter.And.Prefix)
		}
		// Tag
		if len(filter.And.Tags) > 0 {
			rule["tags"] = flattenObjectBucketTags(filter.And.Tags)
		}
		// ObjectSizeGreaterThan
		if filter.And.ObjectSizeGreaterThan != nil && *filter.And.ObjectSizeGreaterThan > 0 {
			rule["object_size_greater_than"] = filter.And.ObjectSizeGreaterThan
		}
		// ObjectSizeLessThan
		if filter.And.ObjectSizeLessThan != nil && *filter.And.ObjectSizeLessThan > 0 {
			rule["object_size_less_than"] = filter.And.ObjectSizeLessThan
		}
	} else {
		// Prefix
		if filter.Prefix != nil && aws.ToString(filter.Prefix) != "" {
			rule["prefix"] = aws.ToString(filter.Prefix)
		}
		// Tag
		if filter.Tag != nil {
			rule["tags"] = flattenObjectBucketTags([]s3Types.Tag{*filter.Tag})
		}
		// ObjectSizeGreaterThan
		if filter.ObjectSizeGreaterThan != nil && *filter.ObjectSizeGreaterThan > 0 {
			rule["object_size_greater_than"] = filter.ObjectSizeGreaterThan
		}
		// ObjectSizeLessThan
		if filter.ObjectSizeLessThan != nil && *filter.ObjectSizeLessThan > 0 {
			rule["object_size_less_than"] = filter.ObjectSizeLessThan
		}
	}
}

func resourceBucketLifecycleRulesExpirationRead(expiration *s3Types.LifecycleExpiration, rule map[string]any) {
	if expiration == nil {
		return
	}

	e := make(map[string]any)
	if expiration.Days != nil {
		e["days"] = int(aws.ToInt32(expiration.Days))
	}

	if expiration.Date != nil {
		e["date"] = aws.ToString(new(expiration.Date.Format("2006-01-02")))
	}

	if expiration.ExpiredObjectDeleteMarker != nil {
		e["expired_object_delete_marker"] = aws.ToBool(expiration.ExpiredObjectDeleteMarker)
	}

	rule["expiration"] = []any{e}
}

func resourceBucketLifecycleRulesTransitionsRead(transitions []s3Types.Transition, rule map[string]any) {
	if len(transitions) < 1 {
		return
	}

	parsedTransitions := make([]any, 0, len(transitions))

	for _, v := range transitions {
		t := make(map[string]any)
		if v.Days != nil {
			t["days"] = int(aws.ToInt32(v.Days))
		}

		if v.Date != nil {
			t["date"] = aws.ToString(new(v.Date.Format("2006-01-02")))
		}

		if v.StorageClass != "" {
			t["storage_class"] = string(v.StorageClass)
		}

		parsedTransitions = append(parsedTransitions, t)
	}

	rule["transition"] = schema.NewSet(transitionHash, parsedTransitions)
}

func resourceBucketLifecycleRulesNonCurrentVersionExpiration(noncurrentExpiration *s3Types.NoncurrentVersionExpiration, rule map[string]any) {
	if noncurrentExpiration == nil {
		return
	}

	e := make(map[string]any)
	if noncurrentExpiration.NewerNoncurrentVersions != nil {
		e["newer_noncurrent_versions"] = int(aws.ToInt32(noncurrentExpiration.NewerNoncurrentVersions))
	}

	if noncurrentExpiration.NoncurrentDays != nil {
		e["noncurrent_days"] = int(aws.ToInt32(noncurrentExpiration.NoncurrentDays))
	}

	rule["noncurrent_version_expiration"] = []any{e}
}

func resourceBucketLifecycleRulesNonCurrentVersionTransitions(noncurrentVersionTransitions []s3Types.NoncurrentVersionTransition, rule map[string]any) {
	if len(noncurrentVersionTransitions) < 1 {
		return
	}

	noncurrentTransitions := make([]any, 0, len(noncurrentVersionTransitions))

	for _, v := range noncurrentVersionTransitions {
		t := make(map[string]any)
		if v.NoncurrentDays != nil {
			t["noncurrent_days"] = int(aws.ToInt32(v.NoncurrentDays))
		}

		if v.NewerNoncurrentVersions != nil {
			t["newer_noncurrent_versions"] = int(aws.ToInt32(v.NewerNoncurrentVersions))
		}

		if v.StorageClass != "" {
			t["storage_class"] = string(v.StorageClass)
		}

		noncurrentTransitions = append(noncurrentTransitions, t)
	}

	rule["noncurrent_version_transition"] = schema.NewSet(noncurrentVersionTransitionHash, noncurrentTransitions)
}

/*
*** DELETE
 */

func resourceObjectBucketDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())

	var nObjectDeleted int64

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: new(bucketName),
	})
	if err != nil {
		if _, ok := errors.AsType[*s3Types.NoSuchBucket](err); ok {
			return nil // Bucket does not exist, so consider it deleted
		}

		if IsS3Err(err, ErrCodeBucketNotEmpty, "") {
			if d.Get("force_destroy").(bool) {
				nObjectDeleted, err = emptyBucket(ctx, s3Client, bucketName, true)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error S3 bucket force_destroy: %w", err))
				}

				log.Printf("[DEBUG] Deleted %d S3 objects", nObjectDeleted)

				return resourceObjectBucketDelete(ctx, d, m)
			}
		}
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

/*
*** VALIDATE
 */

func validBucketLifecycleTimestamp(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	_, err := time.Parse(time.RFC3339, value+"T00:00:00Z")
	if err != nil {
		errors = append(errors, fmt.Errorf(
			"%q cannot be parsed as RFC3339 Timestamp Format", value))
	}

	return
}

func validateBucket(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
	// Object lock and versioning
	if diff.Get("object_lock_enabled").(bool) {
		if diff.HasChange("versioning") && !diff.Get("versioning.0.enabled").(bool) {
			return errors.New("versioning must be enabled when object lock is enabled")
		}
	}

	// Lifecycle rules
	ruleCount := diff.Get("lifecycle_rule.#").(int)

	for i := range ruleCount {
		// Expiration
		if _, ok := diff.GetOk(fmt.Sprintf("lifecycle_rule.%d.expiration", i)); ok {
			if err := validateLifecycleExpiration(diff, i); err != nil {
				return err
			}
		}

		// Transition
		if v, ok := diff.GetOk(fmt.Sprintf("lifecycle_rule.%d.transition", i)); ok {
			// Special treatment for "TypeSet" (can't be simply indexed)
			transitionSet := v.(*schema.Set)
			for _, transitionRaw := range transitionSet.List() {
				transition := transitionRaw.(map[string]any)
				if err := validateLifecycleTransition(transition); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateLifecycleExpiration(diff *schema.ResourceDiff, i int) error {
	prefix := fmt.Sprintf("lifecycle_rule.%d.expiration.0.", i)

	_, daysOk := diff.GetOk(prefix + "days")
	_, dateOk := diff.GetOk(prefix + "date")
	_, markerOk := diff.GetOk(prefix + "expired_object_delete_marker")

	// Implement "ExactlyOneOf"
	count := 0
	if daysOk {
		count++
	}

	if dateOk {
		count++
	}

	if markerOk {
		count++
	}

	if count == 0 {
		return fmt.Errorf("lifecycle_rule.%d.expiration: one (only one) of 'days', 'date', 'expired_object_delete_marker' should be defined", i)
	}

	if count > 1 {
		return fmt.Errorf("lifecycle_rule.%d.expiration: 'days', 'date', 'expired_object_delete_marker' are mutually exclusive", i)
	}

	return nil
}

func validateLifecycleTransition(transition map[string]any) error {
	// At this point, the "days" and "date" fields are initialized.
	// Either with the filled values, or with default zero values, which makes
	// the "ok" value obsolete.
	daysVal := transition["days"]
	dateVal := transition["date"]

	days := daysVal.(int)
	date := dateVal.(string)

	// Implement "ExactlyOneOf"
	count := 0
	if days > 0 {
		count++
	}

	if date != "" {
		count++
	}

	if count == 0 {
		// This case also happens when "days = 0", which is supported according to AWS tests.
		// No errors.
		return nil
	}

	if count > 1 {
		return errors.New("lifecycle_rule.transition: 'days', 'date' are mutually exclusive")
	}

	return nil
}
