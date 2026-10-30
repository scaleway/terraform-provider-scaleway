package object

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/bucket_server_side_encryption_configuration_data_source.md
var bucketServerSideEncryptionConfigurationDataSourceDescription string

var (
	_ datasource.DataSource              = (*BucketServerSideEncryptionConfigurationDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*BucketServerSideEncryptionConfigurationDataSource)(nil)
)

func NewBucketServerSideEncryptionConfigurationDataSource() datasource.DataSource {
	return &BucketServerSideEncryptionConfigurationDataSource{}
}

type BucketServerSideEncryptionConfigurationDataSource struct {
	meta *meta.Meta
}

type bucketSSEConfigurationDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Bucket    types.String `tfsdk:"bucket"`
	ConfigID  types.String `tfsdk:"bucket_server_side_encryption_configuration_id"`
	Rule      types.Set    `tfsdk:"rule"`
	Region    types.String `tfsdk:"region"`
	ProjectID types.String `tfsdk:"project_id"`
}

func (d *BucketServerSideEncryptionConfigurationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_bucket_server_side_encryption_configuration"
}

func (d *BucketServerSideEncryptionConfigurationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: bucketServerSideEncryptionConfigurationDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the bucket server side encryption configuration",
			},
			"bucket": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The bucket's name or regional ID.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("bucket_server_side_encryption_configuration_id")),
				},
			},
			"bucket_server_side_encryption_configuration_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the bucket server side encryption configuration",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("bucket")),
				},
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
		Blocks: map[string]schema.Block{
			"rule": schema.SetNestedBlock{
				MarkdownDescription: "Set of server-side encryption configuration rules",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"apply_server_side_encryption_by_default": schema.ListNestedBlock{
							MarkdownDescription: "Single object for setting server-side encryption by default.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"kms_master_key_id": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Scaleway KMS master key ID used for the SSE-KMS encryption.",
									},
									"sse_algorithm": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Server-side encryption algorithm to use.",
									},
								},
							},
						},
					},
					Attributes: map[string]schema.Attribute{
						"bucket_key_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether or not to use Scaleway Object Bucket Keys for SSE-KMS.",
						},
					},
				},
			},
		},
	}
}

func (d *BucketServerSideEncryptionConfigurationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BucketServerSideEncryptionConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bucketSSEConfigurationDataSourceModel
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

	var (
		bucket string
		id     string
	)

	switch {
	case config.ConfigID.ValueString() != "":
		// Lookup by regional configuration ID (= bucket regional ID).
		id = config.ConfigID.ValueString()

		parsedRegion, parsedBucket, perr := regional.ParseID(id)
		if perr != nil {
			resp.Diagnostics.AddError("Failed to parse configuration ID", perr.Error())

			return
		}

		region = parsedRegion
		bucket = parsedBucket
	case config.Bucket.ValueString() != "":
		regionalID := regional.ExpandID(config.Bucket.ValueString())
		bucket = regionalID.ID

		if regionalID.Region != "" && regionalID.Region != region {
			region = regionalID.Region
		}

		id = regional.NewIDString(region, bucket)
	default:
		resp.Diagnostics.AddError("Missing lookup attribute", "Either bucket or bucket_server_side_encryption_configuration_id must be specified.")

		return
	}

	config.ID = types.StringValue(id)
	config.Region = types.StringValue(region.String())

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, d.meta, region.String(), projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	sse, err := findServerSideEncryptionConfiguration(ctx, s3Client, bucket)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SSE configuration", fmt.Errorf("no server side encryption configuration found for bucket %s: %w", bucket, err).Error())

		return
	}

	config.Bucket = types.StringValue(bucket)

	ruleSet, ruleDiags := flattenSSERules(ctx, sse.Rules)
	resp.Diagnostics.Append(ruleDiags...)
	config.Rule = ruleSet

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
