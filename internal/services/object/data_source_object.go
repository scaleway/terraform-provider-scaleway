package object

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*ObjectDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ObjectDataSource)(nil)
)

func NewObjectDataSource() datasource.DataSource {
	return &ObjectDataSource{}
}

type ObjectDataSource struct {
	meta *meta.Meta
}

type objectDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Bucket         types.String `tfsdk:"bucket"`
	Key            types.String `tfsdk:"key"`
	File           types.String `tfsdk:"file"`
	Content        types.String `tfsdk:"content"`
	ContentBase64  types.String `tfsdk:"content_base64"`
	Hash           types.String `tfsdk:"hash"`
	StorageClass   types.String `tfsdk:"storage_class"`
	Metadata       types.Map    `tfsdk:"metadata"`
	ContentType    types.String `tfsdk:"content_type"`
	Tags           types.Map    `tfsdk:"tags"`
	Visibility     types.String `tfsdk:"visibility"`
	SSECustomerKey types.String `tfsdk:"sse_customer_key"`
	Region         types.String `tfsdk:"region"`
	ProjectID      types.String `tfsdk:"project_id"`
}

func (d *ObjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object"
}

func (d *ObjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Scaleway Object Storage object data source provides access to an object.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The regional ID of the object (region/bucket/key).",
			},
			"bucket": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The bucket's name or regional ID.",
			},
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Key of the object",
			},
			"file": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Path of the file to upload, defaults to an empty file",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("content"), path.MatchRoot("content_base64")),
				},
			},
			"content": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Content of the file to upload",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("file"), path.MatchRoot("content_base64")),
				},
			},
			"content_base64": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Content of the file to upload, should be base64 encoded",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("file"), path.MatchRoot("content")),
				},
			},
			"hash": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "File hash to trigger upload",
			},
			"storage_class": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Specifies the Scaleway Object Storage class to which you want the object to transition",
				Validators: []validator.String{
					stringvalidator.OneOf(TransitionSCWStorageClassValues()...),
				},
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Map of object's metadata, only lower case keys are allowed",
				Validators: []validator.Map{
					lowercaseMapKeysValidator{},
				},
			},
			"content_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The standard MIME type of the object's content",
			},
			"tags": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Map of object's tags",
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Visibility of the object, public-read or private",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(s3Types.ObjectCannedACLPrivate),
						string(s3Types.ObjectCannedACLPublicRead),
					),
				},
			},
			"sse_customer_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Customer's encryption keys to encrypt data (SSE-C)",
				Validators: []validator.String{
					stringvalidator.LengthBetween(32, 32),
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
	}
}

func (d *ObjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ObjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config objectDataSourceModel
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

	key := config.Key.ValueString()

	tflog.Debug(ctx, fmt.Sprintf("SCW object read for bucket=%s key=%s", bucket, key))

	_, err = s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read object", fmt.Errorf("couldn't read object %s/%s: %w", bucket, key, err).Error())

		return
	}

	config.ID = types.StringValue(regional.NewIDString(region, objectID(bucket, key)))
	config.Region = types.StringValue(region.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
