package object

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
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

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The bucket's name or regional ID.",
			},
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
									"sse_algorithm": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "Server-side encryption algorithm to use. Valid values are AES256",
										ValidateFunc: validation.StringInSlice([]string{string(awstypes.ServerSideEncryptionAes256)}, true),
									},
								},
							},
						},
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

	d.SetId(regional.NewIDString(region, bucketName))

	_, err = findServerSideEncryptionConfiguration(ctx, conn, bucketName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("waiting for S3 Bucket Server-side Encryption Configuration (%s) create: %w", d.Id(), err))
	}

	return resourceBucketServerSideEncryptionConfigurationRead(ctx, d, meta)
}

func resourceBucketServerSideEncryptionConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	s3Client, _, bucketName, err := s3ClientWithRegionAndName(ctx, d, meta, d.Id())
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

	_ = d.Set("bucket", bucketName)
	if err := d.Set("rule", flattenServerSideEncryptionRules(sse.Rules)); err != nil {
		return diag.FromErr(err)
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

func findServerSideEncryptionConfiguration(ctx context.Context, conn *s3.Client, bucketName string) (*awstypes.ServerSideEncryptionConfiguration, error) {
	input := s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	}

	output, err := conn.GetBucketEncryption(ctx, &input)

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucket, ErrCodeServerSideEncryptionConfigurationNotFound) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServerSideEncryptionConfiguration == nil {
		return nil, fmt.Errorf("no configuration found for bucket %q", bucketName)
	}

	return output.ServerSideEncryptionConfiguration, nil
}

func expandServerSideEncryptionByDefault(l []any) *awstypes.ServerSideEncryptionByDefault {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]any)
	if !ok {
		return nil
	}

	sse := &awstypes.ServerSideEncryptionByDefault{}

	if v, ok := tfMap["sse_algorithm"].(string); ok && v != "" {
		sse.SSEAlgorithm = awstypes.ServerSideEncryption(v)
	}

	return sse
}

func expandServerSideEncryptionRules(l []any) []awstypes.ServerSideEncryptionRule {
	var rules []awstypes.ServerSideEncryptionRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		rule := awstypes.ServerSideEncryptionRule{}

		if v, ok := tfMap["apply_server_side_encryption_by_default"].([]any); ok && len(v) > 0 && v[0] != nil {
			rule.ApplyServerSideEncryptionByDefault = expandServerSideEncryptionByDefault(v)
		}

		if v, ok := tfMap["blocked_encryption_types"].([]any); ok && len(v) > 0 {
			rule.BlockedEncryptionTypes = expandBlockedEncryptionTypes(v)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenServerSideEncryptionRules(rules []awstypes.ServerSideEncryptionRule) []any {
	results := make([]any, len(rules))

	for _, rule := range rules {
		m := make(map[string]any)

		if rule.ApplyServerSideEncryptionByDefault != nil {
			m["apply_server_side_encryption_by_default"] = flattenServerSideEncryptionByDefault(rule.ApplyServerSideEncryptionByDefault)
		}

		if rule.BlockedEncryptionTypes != nil {
			if flattened := flattenBlockedEncryptionTypes(rule.BlockedEncryptionTypes); flattened != nil {
				m["blocked_encryption_types"] = flattened
			}
		}

		results = append(results, m)
	}

	return results
}

func flattenServerSideEncryptionByDefault(sse *awstypes.ServerSideEncryptionByDefault) []any {
	if sse == nil {
		return nil
	}

	m := map[string]any{
		"sse_algorithm": sse.SSEAlgorithm,
	}

	return []any{m}
}

func expandBlockedEncryptionTypes(l []any) *awstypes.BlockedEncryptionTypes {
	if len(l) == 0 {
		return nil
	}

	var encryptionTypes []awstypes.EncryptionType
	for _, v := range l {
		encryptionTypes = append(encryptionTypes, awstypes.EncryptionType(v.(string)))
	}

	return &awstypes.BlockedEncryptionTypes{
		EncryptionType: encryptionTypes,
	}
}

func flattenBlockedEncryptionTypes(bet *awstypes.BlockedEncryptionTypes) []any {
	if bet == nil || len(bet.EncryptionType) == 0 {
		return nil
	}

	var result []any
	for _, et := range bet.EncryptionType {
		result = append(result, string(et))
	}

	return result
}
