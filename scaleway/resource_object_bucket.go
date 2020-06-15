package scaleway

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayObjectBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayObjectBucketCreate,
		Read:   resourceScalewayObjectBucketRead,
		Update: resourceScalewayObjectBucketUpdate,
		Delete: resourceScalewayObjectBucketDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"region": regionSchema(),
		},
	}
}

func resourceScalewayObjectBucketCreate(d *schema.ResourceData, m interface{}) error {
	bucketName := d.Get("name").(string)
	acl := d.Get("acl").(string)

	s3Client, region, err := s3ClientWithRegion(d, m)
	if err != nil {
		return err
	}

	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		ACL:    aws.String(acl),
	})
	if err != nil {
		return err
	}

	tagsSet := make([]*s3.Tag, 0)

	for key, value := range d.Get("tags").(map[string]interface{}) {
		tagsSet = append(tagsSet, &s3.Tag{
			Key:   &key,
			Value: scw.StringPtr(value.(string)),
		})
	}

	if len(tagsSet) > 0 {
		_, err = s3Client.PutBucketTagging(&s3.PutBucketTaggingInput{
			Bucket: aws.String(bucketName),
			Tagging: &s3.Tagging{
				TagSet: tagsSet,
			},
		})
		if err != nil {
			return err
		}
	}

	d.SetId(newRegionalId(region, bucketName))

	return resourceScalewayObjectBucketRead(d, m)
}

func resourceScalewayObjectBucketRead(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(m, d.Id())
	if err != nil {
		return err
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

	_, err = s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if serr, ok := err.(awserr.Error); ok && serr.Code() == s3.ErrCodeNoSuchBucket {
			l.Errorf("Bucket %q was not found - removing from state!", bucketName)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("couldn't read bucket: %s", err)
	}

	var tagsSet []*s3.Tag

	tagsResponse, err := s3Client.GetBucketTagging(&s3.GetBucketTaggingInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if serr, ok := err.(awserr.Error); !ok || serr.Code() != "NoSuchTagSet" {
			return fmt.Errorf("couldn't read tags from bucket: %s", err)
		}
	} else {
		tagsSet = tagsResponse.TagSet
	}

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

	_ = d.Set("tags", tags)

	return nil
}

func resourceScalewayObjectBucketUpdate(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(m, d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("acl") {
		acl := d.Get("acl").(string)

		_, err := s3Client.PutBucketAcl(&s3.PutBucketAclInput{
			Bucket: aws.String(bucketName),
			ACL:    aws.String(acl),
		})
		if err != nil {
			l.Errorf("Couldn't update bucket ACL: %s", err)
			return fmt.Errorf("couldn't update bucket ACL: %s", err)
		}
	}

	if d.HasChange("tags") {
		tagsSet := make([]*s3.Tag, 0)

		for key, value := range d.Get("tags").(map[string]interface{}) {
			tagsSet = append(tagsSet, &s3.Tag{
				Key:   &key,
				Value: scw.StringPtr(value.(string)),
			})
		}

		_, err = s3Client.PutBucketTagging(&s3.PutBucketTaggingInput{
			Bucket: aws.String(bucketName),
			Tagging: &s3.Tagging{
				TagSet: tagsSet,
			},
		})
	}

	return resourceScalewayObjectBucketRead(d, m)
}

func resourceScalewayObjectBucketDelete(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(m, d.Id())
	if err != nil {
		return err
	}

	_, err = s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}
