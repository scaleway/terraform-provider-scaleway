package scaleway

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func resourceScalewayObject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayObjectCreate,
		ReadContext:   resourceScalewayObjectRead,
		UpdateContext: resourceScalewayObjectUpdate,
		DeleteContext: resourceScalewayObjectDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultObjectBucketTimeout),
			Create:  schema.DefaultTimeout(defaultObjectBucketTimeout),
			Read:    schema.DefaultTimeout(defaultObjectBucketTimeout),
			Update:  schema.DefaultTimeout(defaultObjectBucketTimeout),
			Delete:  schema.DefaultTimeout(defaultObjectBucketTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The bucket's name or regional ID.",
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Key of the object",
			},
			"file": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Path of the file to upload, defaults to an empty file",
				ConflictsWith: []string{"content", "content_base64"},
			},
			"content": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Content of the file to upload",
				ConflictsWith: []string{"file", "content_base64"},
			},
			"content_base64": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Content of the file to upload, should be base64 encoded",
				ConflictsWith: []string{"file", "content"},
			},
			"hash": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File hash to trigger upload",
			},
			"storage_class": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(TransitionSCWStorageClassValues(), false),
				Description:  "Specifies the Scaleway Object Storage class to which you want the object to transition",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Map of object's metadata, only lower case keys are allowed",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateDiagFunc: validateMapKeyLowerCase(),
			},
			"tags": {
				Optional:    true,
				Type:        schema.TypeMap,
				Description: "Map of object's tags",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"visibility": {
				Optional:    true,
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Visibility of the object, public-read or private",
				ValidateFunc: validation.StringInSlice([]string{
					s3.ObjectCannedACLPrivate,
					s3.ObjectCannedACLPublicRead,
				}, false),
			},
			"region":     regional.Schema(),
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayObjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(d, m.(*meta.Meta))
	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	regionalID := regional.ExpandID(d.Get("bucket"))
	bucket := regionalID.ID
	bucketRegion := regionalID.Region

	if bucketRegion != "" && bucketRegion != region {
		s3Client, err = s3ClientForceRegion(d, m.(*meta.Meta), bucketRegion.String())
		if err != nil {
			return diag.FromErr(err)
		}
		region = bucketRegion
	}

	key := d.Get("key").(string)

	req := &s3.PutObjectInput{
		ACL:          expandStringPtr(d.Get("visibility").(string)),
		Bucket:       expandStringPtr(bucket),
		Key:          expandStringPtr(key),
		StorageClass: expandStringPtr(d.Get("storage_class")),
		Metadata:     expandMapStringStringPtr(d.Get("metadata")),
	}

	if filePath, hasFile := d.GetOk("file"); hasFile {
		file, err := os.Open(filePath.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		req.Body = file
	} else if content, hasContent := d.GetOk("content"); hasContent {
		contentString := []byte(content.(string))
		req.Body = bytes.NewReader(contentString)
	} else if content, hasContent := d.GetOk("content_base64"); hasContent {
		contentString := []byte(content.(string))
		decoded := make([]byte, base64.StdEncoding.DecodedLen(len(contentString)))
		_, err = base64.StdEncoding.Decode(decoded, contentString)
		if err != nil {
			return diag.FromErr(err)
		}
		req.Body = bytes.NewReader(decoded)
	} else {
		req.Body = bytes.NewReader([]byte{})
	}

	_, err = s3Client.PutObjectWithContext(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	if rawTags, hasTags := d.GetOk("tags"); hasTags {
		_, err := s3Client.PutObjectTaggingWithContext(ctx, &s3.PutObjectTaggingInput{
			Bucket: expandStringPtr(bucket),
			Key:    expandStringPtr(key),
			Tagging: &s3.Tagging{
				TagSet: expandObjectBucketTags(rawTags),
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(regional.NewIDString(region, objectID(bucket, key)))

	return resourceScalewayObjectRead(ctx, d, m)
}

func resourceScalewayObjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, region, key, bucket, err := s3ClientWithRegionAndNestedName(d, m.(*meta.Meta), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	bucketUpdated := regional.ExpandID(d.Get("bucket")).ID
	keyUpdated := d.Get("key").(string)

	if d.HasChanges("file", "hash") {
		req := &s3.PutObjectInput{
			Bucket:       expandStringPtr(bucketUpdated),
			Key:          expandStringPtr(keyUpdated),
			StorageClass: expandStringPtr(d.Get("storage_class")),
			Metadata:     expandMapStringStringPtr(d.Get("metadata")),
			ACL:          expandStringPtr(d.Get("visibility").(string)),
		}

		if filePath, hasFile := d.GetOk("file"); hasFile {
			file, err := os.Open(filePath.(string))
			if err != nil {
				return diag.FromErr(err)
			}
			req.Body = file
		} else {
			req.Body = bytes.NewReader([]byte{})
		}
		_, err = s3Client.PutObjectWithContext(ctx, req)
	} else {
		_, err = s3Client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
			Bucket:       expandStringPtr(bucketUpdated),
			Key:          expandStringPtr(keyUpdated),
			StorageClass: expandStringPtr(d.Get("storage_class")),
			CopySource:   scw.StringPtr(fmt.Sprintf("%s/%s", bucket, key)),
			Metadata:     expandMapStringStringPtr(d.Get("metadata")),
			ACL:          expandStringPtr(d.Get("visibility").(string)),
		})
	}
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("key", "bucket") {
		_, err := s3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Key:    scw.StringPtr(key),
			Bucket: scw.StringPtr(bucket),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		_, err := s3Client.PutObjectTaggingWithContext(ctx, &s3.PutObjectTaggingInput{
			Bucket: expandStringPtr(bucketUpdated),
			Key:    expandStringPtr(key),
			Tagging: &s3.Tagging{
				TagSet: expandObjectBucketTags(d.Get("tags")),
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(regional.NewIDString(region, objectID(bucketUpdated, keyUpdated)))

	return resourceScalewayObjectCreate(ctx, d, m.(*meta.Meta))
}

func resourceScalewayObjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, region, key, bucket, err := s3ClientWithRegionAndNestedName(d, m.(*meta.Meta), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutRead))
	defer cancel()

	obj, err := s3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: expandStringPtr(bucket),
		Key:    expandStringPtr(key),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("region", region)
	_ = d.Set("bucket", regional.NewIDString(region, bucket))
	_ = d.Set("key", key)

	for k, v := range obj.Metadata {
		if k != strings.ToLower(k) {
			obj.Metadata[strings.ToLower(k)] = v
			delete(obj.Metadata, k)
		}
	}
	_ = d.Set("metadata", flattenMapStringStringPtr(obj.Metadata))

	tags, err := s3Client.GetObjectTaggingWithContext(ctx, &s3.GetObjectTaggingInput{
		Bucket: expandStringPtr(bucket),
		Key:    expandStringPtr(key),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("tags", flattenObjectBucketTags(tags.TagSet))

	acl, err := s3Client.GetObjectAclWithContext(ctx, &s3.GetObjectAclInput{
		Bucket: expandStringPtr(bucket),
		Key:    expandStringPtr(key),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if objectIsPublic(acl) {
		_ = d.Set("visibility", s3.ObjectCannedACLPublicRead)
	} else {
		_ = d.Set("visibility", s3.ObjectCannedACLPrivate)
	}

	return nil
}

func resourceScalewayObjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	s3Client, _, key, bucket, err := s3ClientWithRegionAndNestedName(d, m.(*meta.Meta), d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := &s3.DeleteObjectInput{
		Bucket: expandStringPtr(bucket),
		Key:    expandStringPtr(key),
	}

	_, err = s3Client.DeleteObjectWithContext(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func objectID(bucket, key string) string {
	return fmt.Sprintf("%s/%s", bucket, key)
}

func objectIsPublic(acl *s3.GetObjectAclOutput) bool {
	for _, grant := range acl.Grants {
		if grant.Grantee != nil &&
			*grant.Grantee.Type == s3.TypeGroup &&
			*grant.Grantee.URI == "http://acs.amazonaws.com/groups/global/AllUsers" {
			return true
		}
	}
	return false
}
