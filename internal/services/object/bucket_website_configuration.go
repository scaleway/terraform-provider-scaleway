package object

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ resource.Resource                = (*BucketWebsiteConfigurationResource)(nil)
	_ resource.ResourceWithConfigure   = (*BucketWebsiteConfigurationResource)(nil)
	_ resource.ResourceWithImportState = (*BucketWebsiteConfigurationResource)(nil)
)

func NewBucketWebsiteConfigurationResource() resource.Resource {
	return &BucketWebsiteConfigurationResource{}
}

type BucketWebsiteConfigurationResource struct {
	meta *meta.Meta
}

type bucketWebsiteConfigurationModel struct {
	ID              types.String `tfsdk:"id"`
	Bucket          types.String `tfsdk:"bucket"`
	IndexDocument   types.List   `tfsdk:"index_document"`
	ErrorDocument   types.List   `tfsdk:"error_document"`
	WebsiteEndpoint types.String `tfsdk:"website_endpoint"`
	WebsiteDomain   types.String `tfsdk:"website_domain"`
	Region          types.String `tfsdk:"region"`
	ProjectID       types.String `tfsdk:"project_id"`
}

type websiteIndexDocumentModel struct {
	Suffix types.String `tfsdk:"suffix"`
}

type websiteErrorDocumentModel struct {
	Key types.String `tfsdk:"key"`
}

func (r *BucketWebsiteConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_bucket_website_configuration"
}

func (r *BucketWebsiteConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a static website configuration for a Scaleway Object Storage bucket.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The regional ID of the bucket (region/name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The bucket's name or regional ID.",
				PlanModifiers: []planmodifier.String{
					regional.LocalityPlanModifier(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
				},
			},
			"website_endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The domain of the website endpoint.",
			},
			"website_domain": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The website endpoint.",
			},
			"region":     regional.SchemaAttributeComputed("The region you want to attach the resource to"),
			"project_id": projectIDFrameworkAttribute(),
		},
		Blocks: map[string]schema.Block{
			"index_document": schema.ListNestedBlock{
				MarkdownDescription: "The name of the index document for the website.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"suffix": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Suffix that will be added to the index.",
						},
					},
				},
			},
			"error_document": schema.ListNestedBlock{
				MarkdownDescription: "The name of the error document for the website.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Key for the object to use as an error document.",
						},
					},
				},
			},
		},
	}
}

func (r *BucketWebsiteConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
}

func (r *BucketWebsiteConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketWebsiteConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.resolveBucketAndClient(ctx, plan.Bucket, plan.Region, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	websiteConfig, err := expandWebsiteConfiguration(ctx, plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to expand website configuration", err.Error())

		return
	}

	_, err = s3Client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		resp.Diagnostics.AddError("Couldn't read bucket", err.Error())

		return
	}

	_, err = s3Client.PutBucketWebsite(ctx, &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfig,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating object bucket website configuration", fmt.Errorf("error creating object bucket (%s) website configuration: %w", bucket, err).Error())

		return
	}

	plan.ID = types.StringValue(regional.NewIDString(region, bucket))

	r.readInto(ctx, &plan, s3Client, region, bucket, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketWebsiteConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketWebsiteConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.clientFromState(ctx, state.ID, state.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	r.readInto(ctx, &state, s3Client, region, bucket, &resp.Diagnostics, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketWebsiteConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketWebsiteConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.clientFromState(ctx, plan.ID, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	websiteConfig, err := expandWebsiteConfiguration(ctx, plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to expand website configuration", err.Error())

		return
	}

	_, err = s3Client.PutBucketWebsite(ctx, &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfig,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating object bucket website configuration", fmt.Errorf("error updating Object bucket website configuration (%s): %w", plan.ID.ValueString(), err).Error())

		return
	}

	r.readInto(ctx, &plan, s3Client, region, bucket, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketWebsiteConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketWebsiteConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, bucket, s3Client, err := r.clientFromState(ctx, state.ID, state.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	_, err = s3Client.DeleteBucketWebsite(ctx, &s3.DeleteBucketWebsiteInput{
		Bucket: aws.String(bucket),
	})

	if ErrCodeEquals(err, ErrCodeNoSuchBucket, ErrCodeNoSuchWebsiteConfiguration) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Error deleting object bucket website configuration", fmt.Errorf("error deleting Object bucket website configuration (%s): %w", state.ID.ValueString(), err).Error())

		return
	}
}

func (r *BucketWebsiteConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *BucketWebsiteConfigurationResource) resolveBucketAndClient(ctx context.Context, bucketAttr, regionAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
	region, err := meta.ExtractFrameworkRegion(regionAttr, r.meta.ScwClient())
	if err != nil {
		return "", "", nil, err
	}

	regionalID := regional.ExpandID(bucketAttr.ValueString())
	bucket := regionalID.ID

	if regionalID.Region != "" && regionalID.Region != region {
		region = regionalID.Region
	}

	projectID, err := meta.ExtractFrameworkProjectID(projectIDAttr, r.meta.ScwClient())
	if err != nil {
		return "", "", nil, err
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		return "", "", nil, err
	}

	return region, bucket, s3Client, nil
}

func (r *BucketWebsiteConfigurationResource) clientFromState(ctx context.Context, idAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
	region, bucket, err := regional.ParseID(idAttr.ValueString())
	if err != nil {
		return "", "", nil, err
	}

	projectID, err := meta.ExtractFrameworkProjectID(projectIDAttr, r.meta.ScwClient())
	if err != nil {
		return "", "", nil, err
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		return "", "", nil, err
	}

	return region, bucket, s3Client, nil
}

func (r *BucketWebsiteConfigurationResource) readInto(
	ctx context.Context,
	model *bucketWebsiteConfigurationModel,
	s3Client *s3.Client,
	region scw.Region,
	bucket string,
	diags *diag.Diagnostics,
	isNewResource bool,
) {
	model.Region = types.StringValue(region.String())

	if model.Bucket.IsNull() {
		model.Bucket = types.StringValue(regional.NewIDString(region, bucket))
	}

	_, err := s3Client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if IsS3Err(err, ErrCodeNoSuchBucket, "") {
			tflog.Error(ctx, fmt.Sprintf("Bucket %q was not found - removing from state!", bucket))
			diags.AddWarning("Bucket not found", "Bucket was not found, removing from state")

			return
		}

		diags.AddError("Couldn't read bucket", err.Error())

		return
	}

	output, err := s3Client.GetBucketWebsite(ctx, &s3.GetBucketWebsiteInput{
		Bucket: aws.String(bucket),
	})
	if !isNewResource && ErrCodeEquals(err, ErrCodeNoSuchBucket, ErrCodeNoSuchWebsiteConfiguration) {
		tflog.Debug(ctx, fmt.Sprintf("[WARN] Object Bucket Website Configuration (%s) not found, removing from state", model.ID.ValueString()))
		diags.AddWarning("Website configuration not found", "Object Bucket Website Configuration was not found, removing from state")

		return
	}

	if output == nil {
		if isNewResource {
			diags.AddError("Error reading object bucket website configuration", fmt.Errorf("error reading object bucket website configuration (%s): empty output", model.ID.ValueString()).Error())

			return
		}

		tflog.Info(ctx, fmt.Sprintf("[WARN] object Bucket Website Configuration (%s) not found, removing from state", model.ID.ValueString()))
		diags.AddWarning("Website configuration not found", "Object Bucket Website Configuration was not found, removing from state")

		return
	}

	indexList, indexDiags := flattenWebsiteIndexDocument(ctx, output.IndexDocument)
	diags.Append(indexDiags...)

	model.IndexDocument = indexList

	errorList, errorDiags := flattenWebsiteErrorDocument(ctx, output.ErrorDocument)
	diags.Append(errorDiags...)

	model.ErrorDocument = errorList

	websiteEndpoint := WebsiteEndpoint(bucket, region)
	if websiteEndpoint != nil {
		model.WebsiteEndpoint = types.StringValue(websiteEndpoint.Endpoint)
		model.WebsiteDomain = types.StringValue(websiteEndpoint.Domain)
	}

	acl, err := s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		diags.AddError("Couldn't read bucket acl", err.Error())

		return
	}

	if acl != nil && acl.Owner != nil {
		model.ProjectID = types.StringValue(aws.ToString(NormalizeOwnerID(acl.Owner.ID)))
	}
}

func expandWebsiteConfiguration(ctx context.Context, model bucketWebsiteConfigurationModel, diags *diag.Diagnostics) (*s3Types.WebsiteConfiguration, error) {
	var indexDocs []websiteIndexDocumentModel

	ddiags := model.IndexDocument.ElementsAs(ctx, &indexDocs, false)
	diags.Append(ddiags...)

	if diags.HasError() {
		return nil, nil
	}

	config := &s3Types.WebsiteConfiguration{}

	if len(indexDocs) > 0 {
		suffix := indexDocs[0].Suffix.ValueString()
		config.IndexDocument = &s3Types.IndexDocument{
			Suffix: aws.String(suffix),
		}
	}

	if !model.ErrorDocument.IsNull() && !model.ErrorDocument.IsUnknown() && len(model.ErrorDocument.Elements()) > 0 {
		var errorDocs []websiteErrorDocumentModel

		ediags := model.ErrorDocument.ElementsAs(ctx, &errorDocs, false)
		diags.Append(ediags...)

		if diags.HasError() {
			return nil, nil
		}

		if len(errorDocs) > 0 {
			key := errorDocs[0].Key.ValueString()
			config.ErrorDocument = &s3Types.ErrorDocument{
				Key: aws.String(key),
			}
		}
	}

	return config, nil
}

func flattenWebsiteIndexDocument(ctx context.Context, i *s3Types.IndexDocument) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if i == nil {
		list, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"suffix": types.StringType,
			},
		}, []websiteIndexDocumentModel{})
		diags.Append(d...)

		return list, diags
	}

	m := websiteIndexDocumentModel{}
	if i.Suffix != nil {
		m.Suffix = types.StringValue(aws.ToString(i.Suffix))
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"suffix": types.StringType,
		},
	}, []websiteIndexDocumentModel{m})
	diags.Append(d...)

	return list, diags
}

func flattenWebsiteErrorDocument(ctx context.Context, e *s3Types.ErrorDocument) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if e == nil {
		list, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"key": types.StringType,
			},
		}, []websiteErrorDocumentModel{})
		diags.Append(d...)

		return list, diags
	}

	m := websiteErrorDocumentModel{}
	if e.Key != nil {
		m.Key = types.StringValue(aws.ToString(e.Key))
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"key": types.StringType,
		},
	}, []websiteErrorDocumentModel{m})
	diags.Append(d...)

	return list, diags
}
