package scaleway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
			"region": regionSchema(),
			"versioning": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

	_, err = s3Client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
		Bucket: scw.StringPtr(bucketName),
		ACL:    scw.StringPtr(acl),
	})
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
			l.Errorf("Couldn't update bucket ACL: %s", err)
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

	if d.HasChange("cors_rule") {
		if err := resourceScalewayS3BucketCorsUpdate(ctx, s3Client, d); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayObjectBucketRead(ctx, d, meta)
}

func resourceScalewayObjectBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s3Client, region, bucketName, err := s3ClientWithRegionAndName(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("name", bucketName)

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
			l.Errorf("Bucket %q was not found - removing from state!", bucketName)
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
		if s3err, ok := err.(awserr.Error); !ok || s3err.Code() != "NoSuchTagSet" {
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

	if err != nil && !isS3Err(err, "NoSuchCORSConfiguration", "") {
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
	l.Debugf("S3 put bucket versioning: %#v", i)

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
		l.Debugf("S3 bucket: %s, delete CORS", bucketName)

		_, err := s3conn.DeleteBucketCorsWithContext(ctx, &s3.DeleteBucketCorsInput{
			Bucket: scw.StringPtr(bucketName),
		})

		if err != nil {
			return fmt.Errorf("error deleting S3 CORS: %s", err)
		}
	} else {
		// Put CORS
		rules := expandBucketCORS(rawCors, bucketName)
		corsInput := &s3.PutBucketCorsInput{
			Bucket: scw.StringPtr(bucketName),
			CORSConfiguration: &s3.CORSConfiguration{
				CORSRules: rules,
			},
		}
		l.Debugf("S3 bucket: %s, put CORS: %#v", bucketName, corsInput)

		_, err := s3conn.PutBucketCorsWithContext(ctx, corsInput)
		if err != nil {
			return fmt.Errorf("error putting S3 CORS: %s", err)
		}
	}

	return nil
}
