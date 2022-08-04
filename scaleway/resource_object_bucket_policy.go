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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayObjectBucketPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayObjectBucketPolicyCreate,
		ReadContext:   resourceScalewayObjectBucketPolicyRead,
		UpdateContext: resourceScalewayObjectBuckePolicyUpdate,
		DeleteContext: resourceScalewayObjectBucketPolicyDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultObjectBucketTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the bucket.",
			},
			"policy": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The text of the policy.",
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayObjectBucketPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, bucketName, err := parseRegionalID(d.Get("bucket_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	req := &s3.PutBucketPolicyInput{
		Bucket: scw.StringPtr(bucketName),
	}
	_, err = s3Client.PutBucketPolicyWithContext(ctx, req)
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

func resourceScalewayObjectBucketPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
func resourceScalewayObjectBucketPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, bucketName, err := s3ClientWithRegionAndName(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", bucketName)
	_ = d.Set("region", region)

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
		return s3Client.GetBucketLifecycleConfigurationWithContext(ctx, &s3.GetBucketLifecycleConfigurationInput{
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

func resourceScalewayObjectBucketPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
