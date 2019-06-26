package scaleway

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/utils"
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
			},
			"region": regionSchema(),
		},
	}
}

func resourceScalewayStorageObjectBucketCreate(d *schema.ResourceData, m interface{}) error {
	bucketName := d.Get("name").(string)
	acl := d.Get("acl").(string)
	region := d.Get("region").(string)

	s3Client := m.(*Meta).s3Client

	_, err := s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		ACL:    aws.String(acl),
	})
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(utils.Region(region), bucketName))

	return resourceScalewayStorageObjectBucketRead(d, m)
}

func resourceScalewayStorageObjectBucketRead(d *schema.ResourceData, m interface{}) error {
	bucketName := d.Get("name").(string)

	s3Client := m.(*Meta).s3Client

	_, err := s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if serr, ok := err.(awserr.Error); ok && serr.Code() == s3.ErrCodeNoSuchBucket {
			l.Errorf("Bucket %q was not found - removing from state!", d.Get("name"))
			d.SetId("")
			return nil
		}
		return fmt.Errorf("couldn't read bucket: %s", err)
	}

	return nil
}

func resourceScalewayStorageObjectBucketUpdate(d *schema.ResourceData, m interface{}) error {

	if d.HasChange("acl") {
		bucketName := d.Get("name").(string)
		acl := d.Get("acl").(string)
		s3Client := m.(*Meta).s3Client

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
	bucketName := d.Get("name").(string)

	s3Client := m.(*Meta).s3Client

	_, err := s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}
