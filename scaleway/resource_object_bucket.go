package scaleway

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayObjectBucket() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayObjectBucketCreate,
		ReadContext:   resourceScalewayObjectBucketRead,
		UpdateContext: resourceScalewayObjectBucketUpdate,
		DeleteContext: resourceScalewayObjectBucketDelete,
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
			"acl": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "private",
				Description: "ACL of the bucket: either 'public-read' or 'private'.",
				ValidateFunc: validation.StringInSlice([]string{
					s3.ObjectCannedACLPrivate,
					s3.ObjectCannedACLPublicRead,
					s3.ObjectCannedACLPublicReadWrite,
					s3.ObjectCannedACLAuthenticatedRead,
				}, false),
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
			"cors_rule": {
				Type:     schema.TypeList,
				Optional: true,
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
			"region": regionSchema(),
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
							Default:     false,
						},
					},
				},
			},
		},
	}
}

func resourceScalewayObjectBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	bucketName := d.Get("name").(string)
	acl := d.Get("acl").(string)

	s3Client, region, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &s3.CreateBucketInput{
		Bucket: scw.StringPtr(bucketName),
		ACL:    scw.StringPtr(acl),
	}
	_, err = s3Client.CreateBucketWithContext(ctx, req)
	if TimedOut(err) {
		_, err = s3Client.CreateBucketWithContext(ctx, req)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	tagsSet := expandObjectBucketTags(d.Get("tags"))

	if len(tagsSet) > 0 {
		_, err = s3Client.PutBucketTaggingWithContext(ctx, &s3.PutBucketTaggingInput{
			Bucket: scw.StringPtr(bucketName),
			Tagging: &s3.Tagging{
				TagSet: tagsSet,
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(newRegionalIDString(region, bucketName))

	return resourceScalewayObjectBucketUpdate(ctx, d, meta)
}

func resourceScalewayObjectBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("acl") {
		acl := d.Get("acl").(string)

		_, err := s3Client.PutBucketAclWithContext(ctx, &s3.PutBucketAclInput{
			Bucket: scw.StringPtr(bucketName),
			ACL:    scw.StringPtr(acl),
		})
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Couldn't update bucket ACL: %s", err))
			return diag.FromErr(fmt.Errorf("couldn't update bucket ACL: %s", err))
		}
	}

	if d.HasChange("versioning") {
		if err := resourceScalewayObjectBucketVersioningUpdate(ctx, s3Client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		tagsSet := expandObjectBucketTags(d.Get("tags"))

		if len(tagsSet) > 0 {
			_, err = s3Client.PutBucketTaggingWithContext(ctx, &s3.PutBucketTaggingInput{
				Bucket: scw.StringPtr(bucketName),
				Tagging: &s3.Tagging{
					TagSet: tagsSet,
				},
			})
		} else {
			_, err = s3Client.DeleteBucketTaggingWithContext(ctx, &s3.DeleteBucketTaggingInput{
				Bucket: scw.StringPtr(bucketName),
			})
		}
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("cors_rule") {
		if err := resourceScalewayS3BucketCorsUpdate(ctx, s3Client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("lifecycle_rule") {
		if err := resourceBucketLifecycleUpdate(ctx, s3Client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayObjectBucketRead(ctx, d, meta)
}

//gocyclo:ignore
func resourceBucketLifecycleUpdate(ctx context.Context, conn *s3.S3, d *schema.ResourceData) error {
	bucket := d.Get("name").(string)

	lifecycleRules := d.Get("lifecycle_rule").([]interface{})

	if len(lifecycleRules) == 0 || lifecycleRules[0] == nil {
		i := &s3.DeleteBucketLifecycleInput{
			Bucket: aws.String(bucket),
		}

		_, err := conn.DeleteBucketLifecycle(i)
		if err != nil {
			return fmt.Errorf("error removing S3 lifecycle: %s", err)
		}
		return nil
	}

	rules := make([]*s3.LifecycleRule, 0, len(lifecycleRules))

	for i, lifecycleRule := range lifecycleRules {
		r := lifecycleRule.(map[string]interface{})

		rule := &s3.LifecycleRule{}

		// Filter
		tags := expandObjectBucketTags(r["tags"])
		filter := &s3.LifecycleRuleFilter{}
		if len(tags) > 0 {
			lifecycleRuleAndOp := &s3.LifecycleRuleAndOperator{}
			if len(r["prefix"].(string)) > 0 {
				lifecycleRuleAndOp.SetPrefix(r["prefix"].(string))
			}
			lifecycleRuleAndOp.SetTags(tags)
			filter.SetAnd(lifecycleRuleAndOp)
		} else if len(r["prefix"].(string)) > 0 {
			filter.SetPrefix(r["prefix"].(string))
		}
		rule.SetFilter(filter)

		// ID
		if val, ok := r["id"].(string); ok && val != "" {
			rule.ID = aws.String(val)
		} else {
			rule.ID = aws.String(resource.PrefixedUniqueId("tf-scw-bucket-lifecycle-"))
		}

		// Enabled
		if val, ok := r["enabled"].(bool); ok && val {
			rule.Status = aws.String(s3.ExpirationStatusEnabled)
		} else {
			rule.Status = aws.String(s3.ExpirationStatusDisabled)
		}

		// AbortIncompleteMultipartUpload
		if val, ok := r["abort_incomplete_multipart_upload_days"].(int); ok && val > 0 {
			rule.AbortIncompleteMultipartUpload = &s3.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int64(int64(val)),
			}
		}

		// Expiration
		expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.expiration", i)).([]interface{})
		if len(expiration) > 0 && expiration[0] != nil {
			e := expiration[0].(map[string]interface{})
			i := &s3.LifecycleExpiration{}
			if val, ok := e["days"].(int); ok && val > 0 {
				i.Days = aws.Int64(int64(val))
			}
			rule.Expiration = i
		}

		// Transitions
		transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.transition", i)).(*schema.Set).List()
		if len(transitions) > 0 {
			rule.Transitions = make([]*s3.Transition, 0, len(transitions))
			for _, transition := range transitions {
				transition := transition.(map[string]interface{})
				i := &s3.Transition{}
				if val, ok := transition["days"].(int); ok && val >= 0 {
					i.Days = aws.Int64(int64(val))
				}
				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = aws.String(val)
				}

				rule.Transitions = append(rule.Transitions, i)
			}
		}

		// As a lifecycle rule requires 1 or more transition/expiration actions,
		// we explicitly pass a default ExpiredObjectDeleteMarker value to be able to create
		// the rule while keeping the policy unaffected if the conditions are not met.
		if rule.Expiration == nil && rule.NoncurrentVersionExpiration == nil &&
			rule.Transitions == nil && rule.NoncurrentVersionTransitions == nil &&
			rule.AbortIncompleteMultipartUpload == nil {
			rule.Expiration = &s3.LifecycleExpiration{ExpiredObjectDeleteMarker: aws.Bool(false)}
		}

		rules = append(rules, rule)
	}

	i := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}

	_, err := retryOnAWSCode(ctx, s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return conn.PutBucketLifecycleConfigurationWithContext(ctx, i) //nolint:wrapcheck
	})
	if err != nil {
		return fmt.Errorf("error putting Object Storage lifecycle: %s", err)
	}

	return nil
}

//gocyclo:ignore
func resourceScalewayObjectBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, bucketName, err := s3ClientWithRegionAndName(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", bucketName)
	_ = d.Set("region", region)

	// We do not read `acl` attribute because it could be impossible to find
	// the right canned ACL from a complex ACL object.
	//
	// Known issue:
	// Import a bucket (eg. terraform import scaleway_object_bucket.x fr-par/x)
	// will always trigger a diff (eg. terraform plan) on acl attribute because
	// we do not read it and it has a "private" default value.
	// AWS has the same issue: https://github.com/terraform-providers/terraform-provider-aws/issues/6193

	_, err = s3Client.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		if s3err, ok := err.(awserr.Error); ok && s3err.Code() == s3.ErrCodeNoSuchBucket {
			tflog.Error(ctx, fmt.Sprintf("Bucket %q was not found - removing from state!", bucketName))
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("couldn't read bucket: %s", err))
	}

	var tagsSet []*s3.Tag

	tagsResponse, err := s3Client.GetBucketTaggingWithContext(ctx, &s3.GetBucketTaggingInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		if s3err, ok := err.(awserr.Error); !ok || s3err.Code() != ErrCodeNoSuchTagSet {
			return diag.FromErr(fmt.Errorf("couldn't read tags from bucket: %s", err))
		}
	} else {
		tagsSet = tagsResponse.TagSet
	}

	_ = d.Set("tags", flattenObjectBucketTags(tagsSet))

	_ = d.Set("endpoint", objectBucketEndpointURL(bucketName, region))

	// Read the CORS
	corsResponse, err := s3Client.GetBucketCorsWithContext(ctx, &s3.GetBucketCorsInput{
		Bucket: scw.StringPtr(bucketName),
	})

	if err != nil && !isS3Err(err, ErrCodeNoSuchCORSConfiguration, "") {
		return diag.FromErr(fmt.Errorf("error getting S3 Bucket CORS configuration: %s", err))
	}

	_ = d.Set("cors_rule", flattenBucketCORS(corsResponse))

	_ = d.Set("endpoint", fmt.Sprintf("https://%s.s3.%s.scw.cloud", bucketName, region))

	// Read the versioning configuration
	versioningResponse, err := s3Client.GetBucketVersioningWithContext(ctx, &s3.GetBucketVersioningInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("versioning", flattenObjectBucketVersioning(versioningResponse))

	// Read the lifecycle configuration
	lifecycleResponse, err := retryOnAWSCode(ctx, s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return s3Client.GetBucketLifecycleConfigurationWithContext(ctx, &s3.GetBucketLifecycleConfigurationInput{ //nolint:wrapcheck
			Bucket: scw.StringPtr(bucketName),
		})
	})
	if err != nil && !tfawserr.ErrMessageContains(err, ErrCodeNoSuchLifecycleConfiguration, "") {
		return diag.FromErr(err)
	}

	lifecycleRules := make([]map[string]interface{}, 0)
	if lifecycle, ok := lifecycleResponse.(*s3.GetBucketLifecycleConfigurationOutput); ok && len(lifecycle.Rules) > 0 {
		lifecycleRules = make([]map[string]interface{}, 0, len(lifecycle.Rules))

		for _, lifecycleRule := range lifecycle.Rules {
			log.Printf("[DEBUG] SCW bucket: %s, read lifecycle rule: %v", d.Id(), lifecycleRule)
			rule := make(map[string]interface{})

			// ID
			if lifecycleRule.ID != nil && aws.StringValue(lifecycleRule.ID) != "" {
				rule["id"] = aws.StringValue(lifecycleRule.ID)
			}
			filter := lifecycleRule.Filter
			if filter != nil {
				if filter.And != nil {
					// Prefix
					if filter.And.Prefix != nil && aws.StringValue(filter.And.Prefix) != "" {
						rule["prefix"] = aws.StringValue(filter.And.Prefix)
					}
					// Tag
					if len(filter.And.Tags) > 0 {
						rule["tags"] = flattenObjectBucketTags(filter.And.Tags)
					}
				} else {
					// Prefix
					if filter.Prefix != nil && aws.StringValue(filter.Prefix) != "" {
						rule["prefix"] = aws.StringValue(filter.Prefix)
					}
					// Tag
					if filter.Tag != nil {
						rule["tags"] = flattenObjectBucketTags([]*s3.Tag{filter.Tag})
					}
				}
			} else {
				if lifecycleRule.Prefix != nil {
					rule["prefix"] = aws.StringValue(lifecycleRule.Prefix)
				}
			}

			// Enabled
			if lifecycleRule.Status != nil {
				if aws.StringValue(lifecycleRule.Status) == s3.ExpirationStatusEnabled {
					rule["enabled"] = true
				} else {
					rule["enabled"] = false
				}
			}

			// AbortIncompleteMultipartUploadDays
			if lifecycleRule.AbortIncompleteMultipartUpload != nil {
				if lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
					rule["abort_incomplete_multipart_upload_days"] = int(aws.Int64Value(lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation))
				}
			}

			// expiration
			if lifecycleRule.Expiration != nil {
				e := make(map[string]interface{})
				if lifecycleRule.Expiration.Days != nil {
					e["days"] = int(aws.Int64Value(lifecycleRule.Expiration.Days))
				}
				rule["expiration"] = []interface{}{e}
			}
			//// transition
			if len(lifecycleRule.Transitions) > 0 {
				transitions := make([]interface{}, 0, len(lifecycleRule.Transitions))
				for _, v := range lifecycleRule.Transitions {
					t := make(map[string]interface{})
					if v.Days != nil {
						t["days"] = int(aws.Int64Value(v.Days))
					}
					if v.StorageClass != nil {
						t["storage_class"] = aws.StringValue(v.StorageClass)
					}
					transitions = append(transitions, t)
				}
				rule["transition"] = schema.NewSet(transitionHash, transitions)
			}

			lifecycleRules = append(lifecycleRules, rule)
		}
	}
	if err := d.Set("lifecycle_rule", lifecycleRules); err != nil {
		return diag.FromErr(fmt.Errorf("error setting lifecycle_rule: %s", err))
	}

	return nil
}

func resourceScalewayObjectBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = s3Client.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
		Bucket: scw.StringPtr(bucketName),
	})

	if isS3Err(err, s3.ErrCodeNoSuchBucket, "") {
		return nil
	}

	if isS3Err(err, ErrCodeBucketNotEmpty, "") {
		if d.Get("force_destroy").(bool) {
			err = deleteS3ObjectVersions(ctx, s3Client, bucketName, true)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error S3 bucket force_destroy: %s", err))
			}
			// Try to delete bucket again after deleting objects
			return resourceScalewayObjectBucketDelete(ctx, d, meta)
		}
	}
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceScalewayObjectBucketVersioningUpdate(ctx context.Context, s3conn *s3.S3, d *schema.ResourceData) error {
	v := d.Get("versioning").([]interface{})
	bucketName := d.Get("name").(string)
	vc := expandObjectBucketVersioning(v)

	i := &s3.PutBucketVersioningInput{
		Bucket:                  scw.StringPtr(bucketName),
		VersioningConfiguration: vc,
	}
	tflog.Debug(ctx, fmt.Sprintf("S3 put bucket versioning: %#v", i))

	_, err := s3conn.PutBucketVersioningWithContext(ctx, i)
	if err != nil {
		return fmt.Errorf("error putting S3 versioning: %s", err)
	}

	return nil
}

func resourceScalewayS3BucketCorsUpdate(ctx context.Context, s3conn *s3.S3, d *schema.ResourceData) error {
	bucketName := d.Get("name").(string)
	rawCors := d.Get("cors_rule").([]interface{})

	if len(rawCors) == 0 {
		// Delete CORS
		tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, delete CORS", bucketName))

		_, err := s3conn.DeleteBucketCorsWithContext(ctx, &s3.DeleteBucketCorsInput{
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
			CORSConfiguration: &s3.CORSConfiguration{
				CORSRules: rules,
			},
		}
		tflog.Debug(ctx, fmt.Sprintf("S3 bucket: %s, put CORS: %#v", bucketName, corsInput))

		_, err := s3conn.PutBucketCorsWithContext(ctx, corsInput)
		if err != nil {
			return fmt.Errorf("error putting S3 CORS: %s", err)
		}
	}

	return nil
}
