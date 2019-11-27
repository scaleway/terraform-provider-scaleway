package scaleway

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
			"region": regionSchema(),
		},
	}
}

func resourceScalewayObjectBucketCreate(d *schema.ResourceData, m interface{}) error {
	bucketName := d.Get("name").(string)
	acl := d.Get("acl").(string)

	s3Client, region, err := getS3ClientWithRegion(d, m)

	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		ACL:    aws.String(acl),
	})
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, bucketName))

	return resourceScalewayObjectBucketRead(d, m)
}

func resourceScalewayObjectBucketRead(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := getS3ClientWithRegionAndName(m, d.Id())
	if err != nil {
		return err
	}

	d.Set("name", bucketName)

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

	return nil
}

func resourceScalewayObjectBucketUpdate(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := getS3ClientWithRegionAndName(m, d.Id())
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

	return resourceScalewayObjectBucketRead(d, m)
}

func resourceScalewayObjectBucketDelete(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := getS3ClientWithRegionAndName(m, d.Id())
	if err != nil {
		return err
	}

	_, err = s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}
