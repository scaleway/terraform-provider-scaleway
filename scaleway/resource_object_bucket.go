package scaleway

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	awssmithy "github.com/aws/smithy-go"
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
					string(s3types.ObjectCannedACLPrivate),
					string(s3types.ObjectCannedACLPublicRead),
					string(s3types.ObjectCannedACLPublicReadWrite),
					string(s3types.ObjectCannedACLAuthenticatedRead),
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

	s3Client, region, err := s3ClientWithRegion(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &s3.CreateBucketInput{
		Bucket: scw.StringPtr(bucketName),
		ACL:    s3types.BucketCannedACL(acl),
	}
	_, err = s3Client.CreateBucket(ctx, req)
	if TimedOut(err) {
		_, err = s3Client.CreateBucket(ctx, req)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	tagsSet := expandObjectBucketTags(d.Get("tags"))

	if len(tagsSet) > 0 {
		_, err = s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
			Bucket: scw.StringPtr(bucketName),
			Tagging: &s3types.Tagging{
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
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("acl") {
		acl := d.Get("acl").(string)

		_, err := s3Client.PutBucketAcl(ctx, &s3.PutBucketAclInput{
			Bucket: scw.StringPtr(bucketName),
			ACL:    s3types.BucketCannedACL(acl),
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
			_, err = s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
				Bucket: scw.StringPtr(bucketName),
				Tagging: &s3types.Tagging{
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
func resourceBucketLifecycleUpdate(ctx context.Context, conn *s3.Client, d *schema.ResourceData) error {
	bucket := d.Get("name").(string)

	lifecycleRules := d.Get("lifecycle_rule").([]interface{})

	if len(lifecycleRules) == 0 || lifecycleRules[0] == nil {
		i := &s3.DeleteBucketLifecycleInput{
			Bucket: aws.String(bucket),
		}

		_, err := conn.DeleteBucketLifecycle(ctx, i)
		if err != nil {
			return fmt.Errorf("error removing S3 lifecycle: %s", err)
		}
		return nil
	}

	rules := make([]s3types.LifecycleRule, 0, len(lifecycleRules))

	for i, lifecycleRule := range lifecycleRules {
		r := lifecycleRule.(map[string]interface{})

		rule := s3types.LifecycleRule{}

		// Filter
		tags := expandObjectBucketTags(r["tags"])
		if len(tags) > 0 {
			lifecycleRuleAndOp := s3types.LifecycleRuleAndOperator{}
			if len(r["prefix"].(string)) > 0 {
				lifecycleRuleAndOp.Prefix = scw.StringPtr(r["prefix"].(string))
			}
			lifecycleRuleAndOp.Tags = tags
			rule.Filter = &s3types.LifecycleRuleFilterMemberAnd{Value: lifecycleRuleAndOp}
		} else if len(r["prefix"].(string)) > 0 {
			rule.Filter = &s3types.LifecycleRuleFilterMemberPrefix{Value: r["prefix"].(string)}
		}

		// ID
		if val, ok := r["id"].(string); ok && val != "" {
			rule.ID = aws.String(val)
		} else {
			rule.ID = aws.String(resource.PrefixedUniqueId("tf-scw-bucket-lifecycle-"))
		}

		// Enabled
		if val, ok := r["enabled"].(bool); ok && val {
			rule.Status = s3types.ExpirationStatusEnabled
		} else {
			rule.Status = s3types.ExpirationStatusDisabled
		}

		// AbortIncompleteMultipartUpload
		if val, ok := r["abort_incomplete_multipart_upload_days"].(int); ok && val > 0 {
			rule.AbortIncompleteMultipartUpload = &s3types.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: int32(val),
			}
		}

		// Expiration
		expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.expiration", i)).([]interface{})
		if len(expiration) > 0 && expiration[0] != nil {
			e := expiration[0].(map[string]interface{})
			i := &s3types.LifecycleExpiration{}
			if val, ok := e["days"].(int); ok && val > 0 {
				i.Days = int32(val)
			}
			rule.Expiration = i
		}

		// Transitions
		transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.transition", i)).(*schema.Set).List()
		if len(transitions) > 0 {
			rule.Transitions = make([]s3types.Transition, 0, len(transitions))
			for _, transition := range transitions {
				transition := transition.(map[string]interface{})
				i := s3types.Transition{}
				if val, ok := transition["days"].(int); ok && val >= 0 {
					i.Days = int32(val)
				}
				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = s3types.TransitionStorageClass(val)
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
			rule.Expiration = &s3types.LifecycleExpiration{ExpiredObjectDeleteMarker: false}
		}

		rules = append(rules, rule)
	}

	i := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}

	_, err := retryOnAWSError(ctx, &s3types.NoSuchBucket{}, func() (*s3.PutBucketLifecycleConfigurationOutput, error) {
		return conn.PutBucketLifecycleConfiguration(ctx, i)
	})
	if err != nil {
		return fmt.Errorf("error putting Object Storage lifecycle: %s", err)
	}

	return nil
}

//gocyclo:ignore
func resourceScalewayObjectBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, bucketName, err := s3ClientWithRegionAndName(ctx, meta, d.Id())
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

	_, err = s3Client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		var noSuchBucket *s3types.NoSuchBucket
		if errors.As(err, &noSuchBucket) {
			tflog.Error(ctx, fmt.Sprintf("Bucket %q was not found - removing from state!", bucketName))
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("couldn't read bucket: %s", err))
	}

	var tagsSet []s3types.Tag

	tagsResponse, err := s3Client.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		if s3err, ok := err.(awssmithy.APIError); !ok || s3err.ErrorCode() != ErrCodeNoSuchTagSet {
			return diag.FromErr(fmt.Errorf("couldn't read tags from bucket: %s", err))
		}
	} else {
		tagsSet = tagsResponse.TagSet
	}

	_ = d.Set("tags", flattenObjectBucketTags(tagsSet))

	_ = d.Set("endpoint", objectBucketEndpointURL(bucketName, region))

	// Read the CORS
	corsResponse, err := s3Client.GetBucketCors(ctx, &s3.GetBucketCorsInput{
		Bucket: scw.StringPtr(bucketName),
	})

	if err != nil && !isS3ErrCode(err, ErrCodeNoSuchCORSConfiguration, "") {
		return diag.FromErr(fmt.Errorf("error getting S3 Bucket CORS configuration: %s", err))
	}

	_ = d.Set("cors_rule", flattenBucketCORS(corsResponse))

	_ = d.Set("endpoint", fmt.Sprintf("https://%s.s3.%s.scw.cloud", bucketName, region))

	// Read the versioning configuration
	versioningResponse, err := s3Client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: scw.StringPtr(bucketName),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("versioning", flattenObjectBucketVersioning(versioningResponse))

	// Read the lifecycle configuration
	lifecycle, err := retryOnAWSError(ctx, &s3types.NoSuchBucket{}, func() (*s3.GetBucketLifecycleConfigurationOutput, error) {
		return s3Client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
			Bucket: scw.StringPtr(bucketName),
		})
	})
	if err != nil && isS3ErrCode(err, ErrCodeNoSuchLifecycleConfiguration, "") {
		return diag.FromErr(err)

	}

	lifecycleRules := make([]map[string]interface{}, 0)
	if len(lifecycle.Rules) > 0 {
		lifecycleRules = make([]map[string]interface{}, 0, len(lifecycle.Rules))

		for _, lifecycleRule := range lifecycle.Rules {
			log.Printf("[DEBUG] SCW bucket: %s, read lifecycle rule: %v", d.Id(), lifecycleRule)
			rule := make(map[string]interface{})

			// ID
			if lifecycleRule.ID != nil && *lifecycleRule.ID != "" {
				rule["id"] = *lifecycleRule.ID
			}
			filter := lifecycleRule.Filter
			if filter != nil {
				switch filter := filter.(type) {
				case *s3types.LifecycleRuleFilterMemberAnd:
					// Prefix
					if filter.Value.Prefix != nil && *filter.Value.Prefix != "" {
						rule["prefix"] = *filter.Value.Prefix
					}
					// Tag
					if len(filter.Value.Tags) > 0 {
						rule["tags"] = flattenObjectBucketTags(filter.Value.Tags)
					}
				case *s3types.LifecycleRuleFilterMemberPrefix:
					rule["prefix"] = filter.Value
				case *s3types.LifecycleRuleFilterMemberTag:
					rule["tags"] = flattenObjectBucketTags([]s3types.Tag{filter.Value})

				}
			} else {
				if lifecycleRule.Prefix != nil {
					rule["prefix"] = *lifecycleRule.Prefix
				}
			}

			// Enabled
			switch lifecycleRule.Status {
			case s3types.ExpirationStatusEnabled:
				rule["enabled"] = true
			case s3types.ExpirationStatusDisabled:
				rule["enabled"] = false
			}

			// AbortIncompleteMultipartUploadDays
			if lifecycleRule.AbortIncompleteMultipartUpload != nil {
				rule["abort_incomplete_multipart_upload_days"] = int(lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation)
			}

			// expiration
			if lifecycleRule.Expiration != nil {
				e := make(map[string]interface{})
				e["days"] = int(lifecycleRule.Expiration.Days)
				rule["expiration"] = []interface{}{e}
			}
			//// transition
			if len(lifecycleRule.Transitions) > 0 {
				transitions := make([]interface{}, 0, len(lifecycleRule.Transitions))
				for _, v := range lifecycleRule.Transitions {
					t := make(map[string]interface{})
					t["days"] = int(v.Days)
					t["storage_class"] = v.StorageClass
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
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: scw.StringPtr(bucketName),
	})

	if _, ok := err.(*s3types.NoSuchBucket); ok {
		return nil
	}

	if isS3ErrCode(err, ErrCodeBucketNotEmpty, "") {
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

func resourceScalewayObjectBucketVersioningUpdate(ctx context.Context, s3conn *s3.Client, d *schema.ResourceData) error {
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

func resourceScalewayS3BucketCorsUpdate(ctx context.Context, s3conn *s3.Client, d *schema.ResourceData) error {
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
			CORSConfiguration: &s3types.CORSConfiguration{
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
