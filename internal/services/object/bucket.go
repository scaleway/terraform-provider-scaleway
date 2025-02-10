package object

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/scw"
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
		Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_methods": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_origins": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
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
									"days": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description:  "Specifies the number of days after object creation when the specific rule action takes effect",
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
		},
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
			if diff.Get("object_lock_enabled").(bool) {
				if diff.HasChange("versioning") && !diff.Get("versioning.0.enabled").(bool) {
					return errors.New("versioning must be enabled when object lock is enabled")
				}
			}

			return nil
		},
	}
}

func resourceObjectBucketCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	bucketName := d.Get("name").(string)
	s3Client, region, err := s3ClientWithRegion(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &s3.CreateBucketInput{
		Bucket: scw.StringPtr(bucketName),
	}

	if v, ok := d.GetOk("object_lock_enabled"); ok {
		req.ObjectLockEnabledForBucket = scw.BoolPtr(v.(bool))
	}

	_, err = s3Client.CreateBucket(ctx, req)

	if v, ok := d.GetOk("acl"); ok {
		req.ACL = s3Types.BucketCannedACL(v.(string))
	}

	if TimedOut(err) {
		_, err = s3Client.CreateBucket(ctx, req)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, bucketName))

	tagsSet := ExpandObjectBucketTags(d.Get("tags"))

	if len(tagsSet) > 0 {
		_, err = s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
			Bucket: scw.StringPtr(bucketName),
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

func resourceObjectBucketUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("acl") {
		acl := d.Get("acl").(string)

		_, err := s3Client.PutBucketAcl(ctx, &s3.PutBucketAclInput{
			Bucket: scw.StringPtr(bucketName),
			ACL:    s3Types.BucketCannedACL(acl),
		})
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Couldn't update bucket ACL: %s", err))

			return diag.FromErr(fmt.Errorf("couldn't update bucket ACL: %s", err))
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
				Bucket: scw.StringPtr(bucketName),
				Tagging: &s3Types.Tagging{
					TagSet: tagsSet,
				},
			})
		} else {
			_, err = s3Client.DeleteBucketTagging(ctx, &s3.DeleteBucketTaggingInput{
				Bucket: scw.StringPtr(bucketName),
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

	lifecycleRules := d.Get("lifecycle_rule").([]interface{})

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
		r := lifecycleRule.(map[string]interface{})

		rule := s3Types.LifecycleRule{}

		// Filter
		prefix := r["prefix"].(string)
		tags := ExpandObjectBucketTags(r["tags"])
		ruleHasPrefix := prefix != ""
		filter := &s3Types.LifecycleRuleFilter{}

		if len(tags) > 1 || (ruleHasPrefix && len(tags) == 1) {
			lifecycleRuleAndOp := &s3Types.LifecycleRuleAndOperator{
				Tags: tags,
			}
			if ruleHasPrefix {
				prefix := r["prefix"].(string)
				lifecycleRuleAndOp.Prefix = &prefix
			}
			filter.And = lifecycleRuleAndOp
		}

		if !ruleHasPrefix && len(tags) == 1 {
			filter.Tag = &tags[0]
		} else if ruleHasPrefix && len(tags) == 0 {
			prefix := r["prefix"].(string)
			filter.Prefix = &prefix
		}

		rule.Filter = filter

		// ID
		if val, ok := r["id"].(string); ok && val != "" {
			rule.ID = aws.String(val)
		} else {
			rule.ID = aws.String(id.PrefixedUniqueId("tf-scw-bucket-lifecycle-"))
		}

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
		expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.expiration", i)).([]interface{})
		if len(expiration) > 0 && expiration[0] != nil {
			e := expiration[0].(map[string]interface{})
			i := &s3Types.LifecycleExpiration{}
			if val, ok := e["days"].(int); ok && val > 0 {
				days := int32(val)
				i.Days = aws.Int32(days)
			}
			rule.Expiration = i
		}

		// Transitions
		transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.transition", i)).(*schema.Set).List()
		if len(transitions) > 0 {
			rule.Transitions = []s3Types.Transition{}
			for _, transition := range transitions {
				transition := transition.(map[string]interface{})
				i := s3Types.Transition{}
				if val, ok := transition["days"].(int); ok && val >= 0 {
					days := int32(val)
					i.Days = aws.Int32(days)
				}
				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = s3Types.TransitionStorageClass(val)
				}

				rule.Transitions = append(rule.Transitions, i)
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

//gocyclo:ignore
func resourceObjectBucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, region, bucketName, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	_ = d.Set("name", bucketName)
	_ = d.Set("region", region)

	acl, err := s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "acl", ""); !bucketFound {
			return diags
		}
	} else if acl != nil && acl.Owner != nil {
		_ = d.Set("project_id", NormalizeOwnerID(acl.Owner.ID))
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
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "objects", ""); !bucketFound {
			d.SetId("")

			return diags
		}
	}

	var tagsSet []s3Types.Tag

	tagsResponse, err := s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: scw.StringPtr(bucketName),
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
		Bucket: scw.StringPtr(bucketName),
	})

	if err != nil && !IsS3Err(err, ErrCodeNoSuchCORSConfiguration, "The CORS configuration does not exist") {
		return diag.FromErr(err)
	}

	_ = d.Set("cors_rule", flattenBucketCORS(corsResponse))

	// Read the versioning configuration
	versioningResponse, err := s3Client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "versioning", ""); !bucketFound {
			d.SetId("")

			return diags
		}
	}
	_ = d.Set("versioning", flattenObjectBucketVersioning(versioningResponse))

	// Read the lifecycle configuration
	lifecycle, err := s3Client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "lifecycle configuration", ErrCodeNoSuchLifecycleConfiguration); !bucketFound {
			d.SetId("")

			return diags
		}
	}

	lifecycleRules := make([]map[string]interface{}, 0)
	if lifecycle != nil && len(lifecycle.Rules) > 0 {
		lifecycleRules = make([]map[string]interface{}, 0, len(lifecycle.Rules))

		for _, lifecycleRule := range lifecycle.Rules {
			log.Printf("[DEBUG] SCW bucket: %s, read lifecycle rule: %v", d.Id(), lifecycleRule)
			rule := make(map[string]interface{})

			// ID
			if lifecycleRule.ID != nil && aws.ToString(lifecycleRule.ID) != "" {
				rule["id"] = aws.ToString(lifecycleRule.ID)
			}
			filter := lifecycleRule.Filter
			if filter != nil {
				if filter.And != nil {
					// Prefix
					if filter.And.Prefix != nil && aws.ToString(filter.And.Prefix) != "" {
						rule["prefix"] = aws.ToString(filter.And.Prefix)
					}
					// Tag
					if len(filter.And.Tags) > 0 {
						rule["tags"] = flattenObjectBucketTags(filter.And.Tags)
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
				}
			} else {
				if lifecycleRule.Filter != nil && lifecycleRule.Filter.Prefix != nil {
					rule["prefix"] = aws.ToString(lifecycleRule.Filter.Prefix)
				}
			}

			// Enabled
			if lifecycleRule.Status == s3Types.ExpirationStatusEnabled {
				rule["enabled"] = true
			} else {
				rule["enabled"] = false
			}

			// AbortIncompleteMultipartUploadDays
			if lifecycleRule.AbortIncompleteMultipartUpload != nil {
				if lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
					rule["abort_incomplete_multipart_upload_days"] = int(aws.ToInt32(lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation))
				}
			}

			// expiration
			if lifecycleRule.Expiration != nil {
				e := make(map[string]interface{})
				if lifecycleRule.Expiration.Days != nil {
					e["days"] = int(aws.ToInt32(lifecycleRule.Expiration.Days))
				}
				rule["expiration"] = []interface{}{e}
			}
			//// transition
			if len(lifecycleRule.Transitions) > 0 {
				transitions := make([]interface{}, 0, len(lifecycleRule.Transitions))
				for _, v := range lifecycleRule.Transitions {
					t := make(map[string]interface{})
					if v.Days != nil {
						t["days"] = int(aws.ToInt32(v.Days))
					}
					if v.StorageClass != "" {
						t["storage_class"] = string(v.StorageClass)
					}
					transitions = append(transitions, t)
				}
				rule["transition"] = schema.NewSet(transitionHash, transitions)
			}

			lifecycleRules = append(lifecycleRules, rule)
		}
	}
	if err := d.Set("lifecycle_rule", lifecycleRules); err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("error setting lifecycle_rule: %s", err),
		})
	}

	return diags
}

func resourceObjectBucketDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, d, m, d.Id())
	var nObjectDeleted int64
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		var noSuchBucket *s3Types.NoSuchBucket
		if errors.As(err, &noSuchBucket) {
			return nil // Bucket does not exist, so consider it deleted
		}

		if IsS3Err(err, ErrCodeBucketNotEmpty, "") {
			if d.Get("force_destroy").(bool) {
				nObjectDeleted, err = emptyBucket(ctx, s3Client, bucketName, true)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error S3 bucket force_destroy: %s", err))
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

func resourceObjectBucketVersioningUpdate(ctx context.Context, s3conn *s3.Client, d *schema.ResourceData) error {
	v := d.Get("versioning").([]interface{})
	bucketName := d.Get("name").(string)
	vc := expandObjectBucketVersioning(v)

	i := &s3.PutBucketVersioningInput{
		Bucket:                  scw.StringPtr(bucketName),
		VersioningConfiguration: vc,
	}
	tflog.Debug(ctx, fmt.Sprintf("S3 put bucket versioning: %#v", i))

	_, err := s3conn.PutBucketVersioning(ctx, i)
	if err != nil {
		return fmt.Errorf("error putting S3 versioning: %s", err)
	}

	return nil
}

func resourceS3BucketCorsUpdate(ctx context.Context, s3conn *s3.Client, d *schema.ResourceData) error {
	bucketName := d.Get("name").(string)
	rawCors := d.Get("cors_rule").([]interface{})

	if len(rawCors) == 0 {
		// Delete CORS
		tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, delete CORS", bucketName))

		_, err := s3conn.DeleteBucketCors(ctx, &s3.DeleteBucketCorsInput{
			Bucket: scw.StringPtr(bucketName),
		})
		if err != nil {
			return fmt.Errorf("error deleting S3 CORS: %s", err)
		}
	} else {
		// Put CORS
		rules := expandBucketCORS(ctx, rawCors, bucketName)
		corsInput := &s3.PutBucketCorsInput{
			Bucket: scw.StringPtr(bucketName),
			CORSConfiguration: &s3Types.CORSConfiguration{
				CORSRules: rules,
			},
		}
		tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, put CORS: %#v", bucketName, corsInput))

		_, err := s3conn.PutBucketCors(ctx, corsInput)
		if err != nil {
			return fmt.Errorf("error putting S3 CORS: %s", err)
		}
	}

	return nil
}
