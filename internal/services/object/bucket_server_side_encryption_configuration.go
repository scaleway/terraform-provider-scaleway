package object

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
)

func ResourceBucketServerSideEncryptionConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketServerSideEncryptionConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketServerSideEncryptionConfigurationRead,
		UpdateWithoutTimeout: resourceBucketServerSideEncryptionConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketServerSideEncryptionConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: bucketServerSideEncryptionConfigurationSchema,
		Identity:   identity.DefaultRegional(),
	}
}

func bucketServerSideEncryptionConfigurationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"bucket": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The bucket's name or regional ID.",
		},
		"region":     regional.Schema(),
		"project_id": account.ProjectIDSchema(),
		"rule": {
			Type:        schema.TypeSet,
			Required:    true,
			Description: "Set of server-side encryption configuration rules",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"apply_server_side_encryption_by_default": {
						Type:        schema.TypeList,
						MaxItems:    1,
						Optional:    true,
						Description: "Single object for setting server-side encryption by default.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"kms_master_key_id": {
									Type:     schema.TypeString,
									Optional: true,
									Computed: true,
									Description: "Scaleway KMS master key ID used for the SSE-KMS encryption. " +
										"This can only be used when you set the value of sse_algorithm as 'aws:kms'. " +
										"Will return an error if this element is absent while the sse_algorithm is 'aws:kms'.",
								},
								"sse_algorithm": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Server-side encryption algorithm to use. Valid values are 'AES256', 'aws:kms'",
									ValidateFunc: validation.StringInSlice([]string{
										string(awstypes.ServerSideEncryptionAes256),
										string(awstypes.ServerSideEncryptionAwsKms),
									}, true),
								},
							},
						},
					},
					"bucket_key_enabled": {
						Type:        schema.TypeBool,
						Optional:    true,
						Computed:    true,
						Description: "Whether or not to use Scaleway Object Bucket Keys for SSE-KMS.",
					},
				},
			},
		},
	}
}

func resourceBucketServerSideEncryptionConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn, region, err := s3ClientWithRegion(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	bucketName := d.Get("bucket").(string)

	input := s3.PutBucketEncryptionInput{
		Bucket: &bucketName,
		ServerSideEncryptionConfiguration: &awstypes.ServerSideEncryptionConfiguration{
			Rules: expandServerSideEncryptionRules(d.Get("rule").(*schema.Set).List()),
		},
	}

	_, err = conn.PutBucketEncryption(ctx, &input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("creating S3 Bucket (%s) Server-side Encryption Configuration: %w", bucketName, err))
	}

	err = identity.SetRegionalIdentity(d, region, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = findServerSideEncryptionConfiguration(ctx, conn, bucketName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("waiting for S3 Bucket Server-side Encryption Configuration (%s) create: %w", d.Id(), err))
	}

	return resourceBucketServerSideEncryptionConfigurationRead(ctx, d, meta)
}

func resourceBucketServerSideEncryptionConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	s3Client, region, bucketName, err := s3ClientWithRegionAndName(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	sse, err := findServerSideEncryptionConfiguration(ctx, s3Client, bucketName)
	if err != nil {
		if !d.IsNewResource() {
			log.Printf("[WARN] S3 Bucket Server-side Encryption Configuration (%s) not found, removing from state", d.Id())
			d.SetId("")

			return diags
		}

		return diag.FromErr(fmt.Errorf("reading S3 Bucket Server-side Encryption Configuration (%s): %w", d.Id(), err))
	}

	err = identity.SetRegionalIdentity(d, region, bucketName)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("bucket", bucketName)
	_ = d.Set("region", region)

	if err := d.Set("rule", flattenServerSideEncryptionRules(sse.Rules)); err != nil {
		return diag.FromErr(err)
	}

	acl, err := s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		if bucketFound, _ := addReadBucketErrorDiagnostic(&diags, err, "acl", ""); !bucketFound {
			return diags
		}
	} else if acl != nil && acl.Owner != nil {
		_ = d.Set("project_id", NormalizeOwnerID(acl.Owner.ID))
	}

	return diags
}

func resourceBucketServerSideEncryptionConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucketName),
		ServerSideEncryptionConfiguration: &awstypes.ServerSideEncryptionConfiguration{
			Rules: expandServerSideEncryptionRules(d.Get("rule").(*schema.Set).List()),
		},
	}

	_, err = s3Client.PutBucketEncryption(ctx, &input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("updating S3 Bucket Server-side Encryption Configuration (%s): %w", d.Id(), err))
	}

	return resourceBucketServerSideEncryptionConfigurationRead(ctx, d, meta)
}

func resourceBucketServerSideEncryptionConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, d, meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := s3.DeleteBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	}

	_, err = s3Client.DeleteBucketEncryption(ctx, &input)

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucket, ErrCodeServerSideEncryptionConfigurationNotFound) {
		return diags
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("deleting S3 Bucket Server-side Encryption Configuration (%s): %w", d.Id(), err))
	}

	// Don't wait for the SSE configuration to disappear as the bucket now always has one.

	return diags
}
