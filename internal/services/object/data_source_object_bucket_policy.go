package object

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

//go:embed descriptions/object_bucket_policy_data_source.md
var bucketPolicyDataSourceDescription string

var (
	_ datasource.DataSource              = (*BucketPolicyDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*BucketPolicyDataSource)(nil)
)

func NewBucketPolicyDataSource() datasource.DataSource {
	return &BucketPolicyDataSource{}
}

type BucketPolicyDataSource struct {
	meta *meta.Meta
}

type bucketPolicyDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Bucket    types.String `tfsdk:"bucket"`
	Policy    types.String `tfsdk:"policy"`
	Region    types.String `tfsdk:"region"`
	ProjectID types.String `tfsdk:"project_id"`
}

func (d *BucketPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_bucket_policy"
}

func (d *BucketPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: bucketPolicyDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The regional ID of the bucket policy, in the `{region}/{bucket}` format.",
			},
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "The bucket's name or regional ID.",
			},
			"policy": schema.StringAttribute{
				Computed:    true,
				Description: "The text of the policy.",
			},
			"region": regional.SchemaAttributeComputed("The region you want to attach the resource to"),
			"project_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The project ID the bucket belongs to.",
			},
		},
	}
}

func (d *BucketPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.meta = m
}

func (d *BucketPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bucketPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(config.ProjectID, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	regionalID := regional.ExpandID(config.Bucket.ValueString())
	bucket := regionalID.ID
	bucketRegion := regionalID.Region

	if bucketRegion != "" && bucketRegion != region {
		region = bucketRegion
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, d.meta, region.String(), projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	tflog.Debug(ctx, "[DEBUG] SCW bucket policy, read for bucket: "+bucket)

	pol, err := s3Client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if IsS3Err(err, ErrCodeNoSuchBucketPolicy, "") || IsS3Err(err, ErrCodeNoSuchBucket, "") {
			resp.Diagnostics.AddError(
				"couldn't read bucket policy",
				fmt.Sprintf("bucket %s doesn't exist or has no policy", bucket),
			)

			return
		}

		resp.Diagnostics.AddError(
			"couldn't read bucket policy",
			fmt.Errorf("couldn't read bucket %s policy: %w", bucket, err).Error(),
		)

		return
	}

	policyString := "{}"
	if pol != nil && pol.Policy != nil {
		policyString = aws.ToString(pol.Policy)
	}

	policyJSON, err := structure.NormalizeJsonString(policyString)
	if err != nil {
		resp.Diagnostics.AddError(
			"invalid policy JSON",
			fmt.Sprintf("policy (%s) is an invalid JSON: %s", policyString, err.Error()),
		)

		return
	}

	config.ID = types.StringValue(regional.NewIDString(region, bucket))
	config.Bucket = types.StringValue(regional.NewIDString(region, bucket))
	config.Region = types.StringValue(region.String())
	config.Policy = types.StringValue(policyJSON)

	config.ProjectID, resp.Diagnostics = appendProjectIDFromACL(ctx, s3Client, bucket, resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
