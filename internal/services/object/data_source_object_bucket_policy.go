package object

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

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
		MarkdownDescription: "The Scaleway Object Storage bucket policy data source provides access to a bucket's policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The regional ID of the bucket (region/name).",
			},
			"bucket": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The bucket's name or regional ID.",
			},
			"policy": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The text of the policy.",
			},
			"region": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The region you want to attach the resource to",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(regional.AllRegions()),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The project ID the bucket belongs to. Defaults to the provider's project ID.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
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

	regionalID := regional.ExpandID(config.Bucket.ValueString())
	bucket := regionalID.ID

	tflog.Debug(ctx, "bucket name: "+bucket)

	if regionalID.Region != "" && regionalID.Region != region {
		region = regionalID.Region
	}

	projectID, err := meta.ExtractFrameworkProjectID(config.ProjectID, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, d.meta, region.String(), projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	config.Region = types.StringValue(region.String())

	pol, err := s3Client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucketPolicy, ErrCodeNoSuchBucket) {
			resp.Diagnostics.AddError("Bucket policy not found", fmt.Sprintf("bucket %s doesn't exist or has no policy", bucket))

			return
		}

		resp.Diagnostics.AddError("Couldn't read bucket policy", fmt.Errorf("couldn't read bucket %s policy: %w", bucket, err).Error())

		return
	}

	policyString := "{}"
	if pol.Policy != nil {
		policyString = aws.ToString(pol.Policy)
	}

	policyJSON, err := normalizePolicyJSON(policyString)
	if err != nil {
		resp.Diagnostics.AddError("Invalid policy JSON", fmt.Errorf("policy (%s) is an invalid JSON: %w", policyString, err).Error())

		return
	}

	config.Policy = types.StringValue(policyJSON)

	acl, err := s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		resp.Diagnostics.AddError("Couldn't read bucket acl", err.Error())

		return
	}

	if acl != nil && acl.Owner != nil {
		config.ProjectID = types.StringValue(aws.ToString(NormalizeOwnerID(acl.Owner.ID)))
	}

	config.ID = types.StringValue(regional.NewIDString(region, bucket))
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}