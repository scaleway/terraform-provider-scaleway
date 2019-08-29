package scaleway

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"

	api "github.com/nicolai86/scaleway-sdk"
)

func resourceScalewayBucket() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: `This resource is deprecated and will be removed in the next major version.

Please use scaleway_object_bucket instead.`,

		Create: resourceScalewayBucketCreate,
		Read:   resourceScalewayBucketRead,
		Delete: resourceScalewayBucketDelete,
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
		},
	}
}

func resourceScalewayBucketRead(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	_, err := scaleway.ListObjects(d.Get("name").(string))
	if err != nil {
		if serr, ok := err.(api.APIError); ok && serr.StatusCode == 404 {
			log.Printf("[DEBUG] Bucket %q was not found - removing from state!", d.Get("name").(string))
			d.SetId("")
			return nil
		}
	}

	return err
}

func resourceScalewayBucketCreate(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	container, err := scaleway.CreateBucket(&api.CreateBucketRequest{
		Name:         d.Get("name").(string),
		Organization: scaleway.Organization,
	})
	if err != nil {
		return err
	}

	d.SetId(container.Name)
	return nil
}

func resourceScalewayBucketDelete(d *schema.ResourceData, m interface{}) error {
	scaleway := m.(*Meta).deprecatedClient

	err := scaleway.DeleteBucket(d.Id())
	if err != nil {
		return err
	}
	return nil
}
