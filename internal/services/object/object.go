package object

import (
	"bytes"
	"context"
	"crypto/md5" //nolint:gosec
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/object.md
var objectDescription string

func ResourceObject() *schema.Resource {
	return &schema.Resource{
		Description:   objectDescription,
		CreateContext: resourceObjectCreate,
		ReadContext:   resourceObjectRead,
		UpdateContext: resourceObjectUpdate,
		DeleteContext: resourceObjectDelete,
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
		SchemaFunc: objectSchema,
	}
}

func objectSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"bucket": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "The bucket's name or regional ID.",
			DiffSuppressFunc: dsf.Locality,
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
		"content_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The standard MIME type of the object's content (e.g., 'application/json', 'text/plain'). This specifies how the object should be interpreted by clients. See RFC 9110: https://www.rfc-editor.org/rfc/rfc9110.html#name-content-type",
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
				string(s3Types.ObjectCannedACLPrivate),
				string(s3Types.ObjectCannedACLPublicRead),
			}, false),
		},
		"sse_customer_key": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			Description:   "Customer's encryption keys to encrypt data (SSE-C). Only one of `sse_customer_key` or `sse_customer_key_wo` should be specified.",
			ValidateFunc:  validation.StringLenBetween(32, 32),
			ConflictsWith: []string{"sse_customer_key_wo"},
		},
		"sse_customer_key_wo": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Customer's encryption keys to encrypt data (SSE-C) in [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) mode. Only one of `sse_customer_key` or `sse_customer_key_wo` should be specified. `sse_customer_key_wo` will not be set in the Terraform state. To update the `sse_customer_key_wo`, you must also update the `sse_customer_key_wo_version`. Important: Objects encrypted with `sse_customer_key_wo` cannot be read back by Terraform since the encryption key is not stored in state. This means attributes like `content_type`, `metadata`, and `tags` will not be populated from the actual object. Additionally, data sources cannot use `sse_customer_key_wo` because data sources cannot have write-only attributes - the key would be exposed in the state, defeating the purpose of write-only mode.",
			WriteOnly:     true,
			ConflictsWith: []string{"sse_customer_key"},
			RequiredWith:  []string{"sse_customer_key_wo_version"},
		},
		"sse_customer_key_wo_version": {
			Type:         schema.TypeInt,
			Optional:     true,
			Description:  "The version of the [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) SSE customer key. To update the `sse_customer_key_wo`, you must also update the `sse_customer_key_wo_version`.",
			RequiredWith: []string{"sse_customer_key_wo"},
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
	}
}

func resourceObjectCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, region, err := s3ClientWithRegion(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	regionalID := regional.ExpandID(d.Get("bucket"))
	bucket := regionalID.ID
	bucketRegion := regionalID.Region

	if bucketRegion != "" && bucketRegion != region {
		s3Client, err = s3ClientForceRegion(ctx, d, m, bucketRegion.String())
		if err != nil {
			return diag.FromErr(err)
		}

		region = bucketRegion
	}

	key := d.Get("key").(string)
	storageClassStr := d.Get("storage_class").(string)
	storageClass := s3Types.StorageClass(storageClassStr)

	req := &s3.PutObjectInput{
		Bucket:       types.ExpandStringPtr(bucket),
		Key:          types.ExpandStringPtr(key),
		StorageClass: storageClass,
		Metadata:     types.ExpandMapStringString(d.Get("metadata")),
	}

	if contentType, ok := d.GetOk("content_type"); ok {
		req.ContentType = types.ExpandStringPtr(contentType)
	}

	visibilityStr := types.ExpandStringPtr(d.Get("visibility").(string))
	if visibilityStr != nil {
		req.ACL = s3Types.ObjectCannedACL(*visibilityStr)
	}

	encryptionKeyStr := getEncryptionKeyStr(d)

	if encryptionKeyStr != "" {
		digestMD5, encryption, err := EncryptCustomerKey(encryptionKeyStr)
		if err != nil {
			return diag.FromErr(err)
		}

		req.SSECustomerAlgorithm = scw.StringPtr("AES256")
		req.SSECustomerKeyMD5 = &digestMD5
		req.SSECustomerKey = encryption
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

	_, err = s3Client.PutObject(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	if rawTags, hasTags := d.GetOk("tags"); hasTags {
		_, err := s3Client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
			Bucket: types.ExpandStringPtr(bucket),
			Key:    types.ExpandStringPtr(key),
			Tagging: &s3Types.Tagging{
				TagSet: ExpandObjectBucketTags(rawTags),
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(regional.NewIDString(region, objectID(bucket, key)))

	return resourceObjectRead(ctx, d, m)
}

func EncryptCustomerKey(encryptionKeyStr string) (string, *string, error) {
	encryptionKey := []byte(encryptionKeyStr)
	h := md5.New() //nolint:gosec

	_, err := h.Write(encryptionKey)
	if err != nil {
		return "", nil, err
	}

	digest := h.Sum(nil)
	digestMD5 := base64.StdEncoding.EncodeToString(digest)
	encryption := aws.String(base64.StdEncoding.EncodeToString(encryptionKey))

	return digestMD5, encryption, nil
}

func resourceObjectUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, region, key, bucket, err := s3ClientWithRegionAndNestedName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	bucketUpdated := regional.ExpandID(d.Get("bucket")).ID
	keyUpdated := d.Get("key").(string)

	if d.HasChanges("file", "hash") {
		req := &s3.PutObjectInput{
			Bucket:       types.ExpandStringPtr(bucketUpdated),
			Key:          types.ExpandStringPtr(keyUpdated),
			StorageClass: s3Types.StorageClass(d.Get("storage_class").(string)),
			Metadata:     types.ExpandMapStringString(d.Get("metadata")),
			ACL:          s3Types.ObjectCannedACL(d.Get("visibility").(string)),
		}

		if contentType, ok := d.GetOk("content_type"); ok {
			req.ContentType = types.ExpandStringPtr(contentType)
		}

		encryptionKeyStr := getEncryptionKeyStr(d)

		if encryptionKeyStr != "" {
			digestMD5, encryption, err := EncryptCustomerKey(encryptionKeyStr)
			if err != nil {
				return diag.FromErr(err)
			}

			req.SSECustomerAlgorithm = scw.StringPtr("AES256")
			req.SSECustomerKeyMD5 = &digestMD5
			req.SSECustomerKey = encryption
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

		_, err = s3Client.PutObject(ctx, req)
	} else {
		req := &s3.CopyObjectInput{
			Bucket:       types.ExpandStringPtr(bucketUpdated),
			Key:          types.ExpandStringPtr(keyUpdated),
			StorageClass: s3Types.StorageClass(d.Get("storage_class").(string)),
			CopySource:   scw.StringPtr(fmt.Sprintf("%s/%s", bucket, key)),
			Metadata:     types.ExpandMapStringString(d.Get("metadata")),
			ACL:          s3Types.ObjectCannedACL(d.Get("visibility").(string)),
		}

		if contentType, ok := d.GetOk("content_type"); ok {
			req.ContentType = types.ExpandStringPtr(contentType)
		}

		encryptionKeyStr := getEncryptionKeyStr(d)

		if encryptionKeyStr != "" {
			digestMD5, encryption, err := EncryptCustomerKey(encryptionKeyStr)
			if err != nil {
				return diag.FromErr(err)
			}

			req.CopySourceSSECustomerAlgorithm = scw.StringPtr("AES256")
			req.CopySourceSSECustomerKeyMD5 = &digestMD5
			req.CopySourceSSECustomerKey = encryption
		}

		_, err = s3Client.CopyObject(ctx, req)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("key", "bucket") {
		_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Key:    scw.StringPtr(key),
			Bucket: scw.StringPtr(bucket),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		_, err := s3Client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
			Bucket: types.ExpandStringPtr(bucketUpdated),
			Key:    types.ExpandStringPtr(key),
			Tagging: &s3Types.Tagging{
				TagSet: ExpandObjectBucketTags(d.Get("tags")),
			},
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(regional.NewIDString(region, objectID(bucketUpdated, keyUpdated)))

	return resourceObjectCreate(ctx, d, m)
}

func resourceObjectRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, region, key, bucket, err := s3ClientWithRegionAndNestedName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutRead))
	defer cancel()

	_ = d.Set("region", region)
	_ = d.Set("bucket", regional.NewIDString(region, bucket))
	_ = d.Set("key", key)

	req := &s3.HeadObjectInput{
		Bucket: types.ExpandStringPtr(bucket),
		Key:    types.ExpandStringPtr(key),
	}

	tags, err := s3Client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: types.ExpandStringPtr(bucket),
		Key:    types.ExpandStringPtr(key),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("tags", flattenObjectBucketTags(tags.TagSet))

	acl, err := s3Client.GetObjectAcl(ctx, &s3.GetObjectAclInput{
		Bucket: types.ExpandStringPtr(bucket),
		Key:    types.ExpandStringPtr(key),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if objectIsPublic(acl) {
		_ = d.Set("visibility", s3Types.ObjectCannedACLPublicRead)
	} else {
		_ = d.Set("visibility", s3Types.ObjectCannedACLPrivate)
	}

	if _, ok := d.GetOk("sse_customer_key_wo_version"); ok {
		// Object was encrypted with write-only key, but we can't read it back
		// since the actual key is not stored in state. This is expected behavior.
		// Skip the HeadObject call and return partial state
		_ = d.Set("content_type", "")
		_ = d.Set("metadata", map[string]string{})

		return nil
	} else if encryption, ok := d.GetOk("sse_customer_key"); ok {
		req.SSECustomerKey = aws.String(base64.StdEncoding.EncodeToString([]byte(encryption.(string))))
		req.SSECustomerAlgorithm = scw.StringPtr("AES256")
	}

	obj, err := s3Client.HeadObject(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	for k, v := range obj.Metadata {
		if k != strings.ToLower(k) {
			obj.Metadata[strings.ToLower(k)] = v
			delete(obj.Metadata, k)
		}
	}

	_ = d.Set("metadata", types.FlattenMap(obj.Metadata))
	_ = d.Set("content_type", &obj.ContentType)

	return nil
}

func resourceObjectDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	s3Client, _, key, bucket, err := s3ClientWithRegionAndNestedName(ctx, d, m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()

	req := &s3.DeleteObjectInput{
		Bucket: types.ExpandStringPtr(bucket),
		Key:    types.ExpandStringPtr(key),
	}

	_, err = s3Client.DeleteObject(ctx, req)
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
			grant.Grantee.Type == s3Types.TypeGroup &&
			*grant.Grantee.URI == AllUsersURI {
			return true
		}
	}

	return false
}

func getEncryptionKeyStr(d *schema.ResourceData) string {
	var encryptionKeyStr string
	if _, ok := d.GetOk("sse_customer_key_wo_version"); ok {
		encryptionKeyStr = d.GetRawConfig().GetAttr("sse_customer_key_wo").AsString()
	} else if encryptionKey, ok := d.GetOk("sse_customer_key"); ok {
		encryptionKeyStr = encryptionKey.(string)
	}

	return encryptionKeyStr
}

func validateMapKeyLowerCase() schema.SchemaValidateDiagFunc {
	return func(i any, _ cty.Path) diag.Diagnostics {
		m := types.ExpandMapStringStringPtr(i)
		for k := range m {
			if strings.ToLower(k) != k {
				return diag.Diagnostics{diag.Diagnostic{
					Severity:      diag.Error,
					AttributePath: cty.IndexStringPath(k),
					Summary:       "Invalid map content",
					Detail:        fmt.Sprintf("key (%s) should be lowercase", k),
				}}
			}
		}

		return nil
	}
}
