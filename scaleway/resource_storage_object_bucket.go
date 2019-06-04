package scaleway

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
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
			"policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "private",
				Description: "Policy of the bucket: either 'public' or 'private'. Private by default.",
			},
		},
	}
}

func resourceScalewayStorageObjectBucketCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	s3Client := m.(*Meta).s3Client

	err := s3Client.MakeBucket(name, "")
	if err != nil {
		return err
	}

	d.SetId(name)
	return nil
}

func resourceScalewayStorageObjectBucketRead(d *schema.ResourceData, m interface{}) error {
	bucketName := d.Get("name").(string)
	s3Client := m.(*Meta).s3Client

	exists, err := s3Client.BucketExists(bucketName)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("[DEBUG] Bucket %q was not found - removing from state!", d.Get("name").(string))
		d.SetId("")
		return nil
	}

	return nil
}

func resourceScalewayStorageObjectBucketUpdate(d *schema.ResourceData, m interface{}) error {

	if d.HasChange("policy") {
		bucketName := d.Get("name").(string)
		policy := d.Get("policy").(string)
		s3Client := m.(*Meta).s3Client

		//policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::my-bucketname/*"],"Sid": ""}]}`

		err := s3Client.SetBucketPolicy(bucketName, policy)
		if err != nil {
			return err
		}
	}

	return resourceScalewayStorageObjectBucketRead(d, m)

}

func resourceScalewayStorageObjectBucketDelete(d *schema.ResourceData, m interface{}) error {
	bucketName := d.Get("name").(string)
	s3Client := m.(*Meta).s3Client

	return s3Client.RemoveBucket(bucketName)
}
