package scaleway

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceScalewayStorageObjectBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayStorageObjectBucketCreate,
		Read:   resourceScalewayStorageObjectBucketRead,
		Update: resourceScalewayStorageObjectBucketUpdate,
		Delete: resourceScalewayStorageObjectBucketDelete,
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
					"private",
					"public-read",
					"public-read-write",
					"aws-exec-read",
					"authenticated-read",
					"bucket-owner-read",
					"bucket-owner-full-control",
					"log-delivery-write",
				}, false),
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayStorageObjectBucketCreate(d *schema.ResourceData, m interface{}) error {
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

	return resourceScalewayStorageObjectBucketRead(d, m)
}

func resourceScalewayStorageObjectBucketRead(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := getS3ClientWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

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

func resourceScalewayStorageObjectBucketUpdate(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := getS3ClientWithRegionAndID(m, d.Id())
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

	return resourceScalewayStorageObjectBucketRead(d, m)
}

func resourceScalewayStorageObjectBucketDelete(d *schema.ResourceData, m interface{}) error {
	s3Client, _, bucketName, err := getS3ClientWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	_, err = s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}
