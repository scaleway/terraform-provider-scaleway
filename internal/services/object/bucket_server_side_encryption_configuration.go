package object

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
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
	"github.com/scaleway/scaleway-sdk-go/scw"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ resource.Resource                = (*BucketServerSideEncryptionConfigurationResource)(nil)
	_ resource.ResourceWithConfigure   = (*BucketServerSideEncryptionConfigurationResource)(nil)
	_ resource.ResourceWithImportState = (*BucketServerSideEncryptionConfigurationResource)(nil)
	_ resource.ResourceWithIdentity    = (*BucketServerSideEncryptionConfigurationResource)(nil)
)

func NewBucketServerSideEncryptionConfigurationResource() resource.Resource {
	return &BucketServerSideEncryptionConfigurationResource{}
}

type BucketServerSideEncryptionConfigurationResource struct {
	meta *meta.Meta
}

type bucketSSEConfigurationModel struct {
	ID        types.String `tfsdk:"id"`
	Bucket    types.String `tfsdk:"bucket"`
	Rule      types.Set    `tfsdk:"rule"`
	Region    types.String `tfsdk:"region"`
	ProjectID types.String `tfsdk:"project_id"`
}

type sseRuleModel struct {
	ApplyServerSideEncryptionByDefault types.List `tfsdk:"apply_server_side_encryption_by_default"`
	BucketKeyEnabled                   types.Bool `tfsdk:"bucket_key_enabled"`
}

type sseDefaultModel struct {
	KMSMasterKeyID types.String `tfsdk:"kms_master_key_id"`
	SSEAlgorithm   types.String `tfsdk:"sse_algorithm"`
}

type bucketSSEConfigurationIdentityModel = identityfw.RegionalIdentity

func (r *BucketServerSideEncryptionConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_bucket_server_side_encryption_configuration"
}

func (r *BucketServerSideEncryptionConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a server-side encryption configuration for a Scaleway Object Storage bucket.",
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
			},
			"region":     regional.SchemaAttributeComputed("The region you want to attach the resource to"),
			"project_id": projectIDFrameworkAttribute(),
		},
		Blocks: map[string]schema.Block{
			"rule": schema.SetNestedBlock{
				MarkdownDescription: "Set of server-side encryption configuration rules",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"apply_server_side_encryption_by_default": schema.ListNestedBlock{
							MarkdownDescription: "Single object for setting server-side encryption by default.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"kms_master_key_id": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										MarkdownDescription: "Scaleway KMS master key ID used for the SSE-KMS encryption. This can only be used when you set the value of sse_algorithm as 'aws:kms'. Will return an error if this element is absent while the sse_algorithm is 'aws:kms'.",
									},
									"sse_algorithm": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "Server-side encryption algorithm to use. Valid values are 'AES256', 'aws:kms'",
										Validators: []validator.String{
											stringvalidator.OneOfCaseInsensitive(
												string(awstypes.ServerSideEncryptionAes256),
												string(awstypes.ServerSideEncryptionAwsKms),
											),
										},
									},
								},
							},
						},
					},
					Attributes: map[string]schema.Attribute{
						"bucket_key_enabled": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Whether or not to use Scaleway Object Bucket Keys for SSE-KMS.",
						},
					},
				},
			},
		},
	}
}

func (r *BucketServerSideEncryptionConfigurationResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *BucketServerSideEncryptionConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketServerSideEncryptionConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketSSEConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.resolveBucketAndClient(ctx, plan.Bucket, plan.Region, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	rules, err := expandSSERules(ctx, plan.Rule, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to expand SSE rules", err.Error())

		return
	}

	_, err = s3Client.PutBucketEncryption(ctx, &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucket),
		ServerSideEncryptionConfiguration: &awstypes.ServerSideEncryptionConfiguration{
			Rules: rules,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating SSE configuration", fmt.Errorf("creating S3 Bucket (%s) Server-side Encryption Configuration: %w", bucket, err).Error())

		return
	}

	plan.ID = types.StringValue(regional.NewIDString(region, bucket))

	if _, err := findServerSideEncryptionConfiguration(ctx, s3Client, bucket); err != nil {
		resp.Diagnostics.AddError("Error waiting for SSE configuration", fmt.Errorf("waiting for S3 Bucket Server-side Encryption Configuration (%s) create: %w", plan.ID.ValueString(), err).Error())

		return
	}

	r.readInto(ctx, &plan, s3Client, region, bucket, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, bucket))...)
}

func (r *BucketServerSideEncryptionConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    bucketSSEConfigurationModel
		identity bucketSSEConfigurationIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		region scw.Region
		bucket string
		err    error
	)

	if identityAvailable {
		region, bucket = identityfw.ResolveRegionalIdentity(identity)
	} else {
		region, bucket, err = regional.ParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse SSE configuration ID", err.Error())

			return
		}
	}

	projectID, err := meta.ExtractFrameworkProjectID(state.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	r.readInto(ctx, &state, s3Client, region, bucket, &resp.Diagnostics, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, bucket))...)
}

func (r *BucketServerSideEncryptionConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketSSEConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.clientFromState(ctx, plan.ID, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	rules, err := expandSSERules(ctx, plan.Rule, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to expand SSE rules", err.Error())

		return
	}

	_, err = s3Client.PutBucketEncryption(ctx, &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucket),
		ServerSideEncryptionConfiguration: &awstypes.ServerSideEncryptionConfiguration{
			Rules: rules,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating SSE configuration", fmt.Errorf("updating S3 Bucket Server-side Encryption Configuration (%s): %w", plan.ID.ValueString(), err).Error())

		return
	}

	r.readInto(ctx, &plan, s3Client, region, bucket, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(region, bucket))...)
}

func (r *BucketServerSideEncryptionConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketSSEConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, bucket, s3Client, err := r.clientFromState(ctx, state.ID, state.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	_, err = s3Client.DeleteBucketEncryption(ctx, &s3.DeleteBucketEncryptionInput{
		Bucket: aws.String(bucket),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucket, ErrCodeServerSideEncryptionConfigurationNotFound) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Error deleting SSE configuration", fmt.Errorf("deleting S3 Bucket Server-side Encryption Configuration (%s): %w", state.ID.ValueString(), err).Error())

		return
	}
}

func (r *BucketServerSideEncryptionConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *BucketServerSideEncryptionConfigurationResource) resolveBucketAndClient(ctx context.Context, bucketAttr, regionAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
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

func (r *BucketServerSideEncryptionConfigurationResource) clientFromState(ctx context.Context, idAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
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

func (r *BucketServerSideEncryptionConfigurationResource) readInto(
	ctx context.Context,
	model *bucketSSEConfigurationModel,
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

	sse, err := findServerSideEncryptionConfiguration(ctx, s3Client, bucket)
	if err != nil {
		if !isNewResource {
			log.Printf("[WARN] S3 Bucket Server-side Encryption Configuration (%s) not found, removing from state", model.ID.ValueString())

			return
		}

		diags.AddError("Error reading SSE configuration", fmt.Errorf("reading S3 Bucket Server-side Encryption Configuration (%s): %w", model.ID.ValueString(), err).Error())

		return
	}

	ruleSet, ruleDiags := flattenSSERules(ctx, sse.Rules)
	diags.Append(ruleDiags...)

	model.Rule = ruleSet

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

func expandSSERules(ctx context.Context, ruleSet types.Set, diags *diag.Diagnostics) ([]awstypes.ServerSideEncryptionRule, error) {
	if ruleSet.IsNull() || ruleSet.IsUnknown() {
		return nil, nil
	}

	var rules []sseRuleModel

	ddiags := ruleSet.ElementsAs(ctx, &rules, false)
	diags.Append(ddiags...)

	if diags.HasError() {
		return nil, nil
	}

	result := make([]awstypes.ServerSideEncryptionRule, 0, len(rules))

	for _, rule := range rules {
		r := awstypes.ServerSideEncryptionRule{}

		if !rule.ApplyServerSideEncryptionByDefault.IsNull() && !rule.ApplyServerSideEncryptionByDefault.IsUnknown() && len(rule.ApplyServerSideEncryptionByDefault.Elements()) > 0 {
			var defaults []sseDefaultModel

			ediags := rule.ApplyServerSideEncryptionByDefault.ElementsAs(ctx, &defaults, false)
			diags.Append(ediags...)

			if diags.HasError() {
				return nil, nil
			}

			if len(defaults) > 0 {
				d := awstypes.ServerSideEncryptionByDefault{
					SSEAlgorithm: awstypes.ServerSideEncryption(defaults[0].SSEAlgorithm.ValueString()),
				}

				if kms := defaults[0].KMSMasterKeyID.ValueString(); kms != "" {
					d.KMSMasterKeyID = aws.String(kms)
				}

				r.ApplyServerSideEncryptionByDefault = &d
			}
		}

		if !rule.BucketKeyEnabled.IsNull() && !rule.BucketKeyEnabled.IsUnknown() && rule.BucketKeyEnabled.ValueBool() {
			r.BucketKeyEnabled = aws.Bool(true)
		}

		result = append(result, r)
	}

	return result, nil
}

func flattenSSERules(ctx context.Context, rules []awstypes.ServerSideEncryptionRule) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	models := make([]sseRuleModel, 0, len(rules))

	for _, rule := range rules {
		m := sseRuleModel{}

		if rule.ApplyServerSideEncryptionByDefault != nil {
			d := sseDefaultModel{
				SSEAlgorithm: types.StringValue(string(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm)),
			}

			if rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != nil {
				d.KMSMasterKeyID = types.StringValue(aws.ToString(rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID))
			} else {
				d.KMSMasterKeyID = types.StringNull()
			}

			defaultList, dd := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"kms_master_key_id": types.StringType,
					"sse_algorithm":     types.StringType,
				},
			}, []sseDefaultModel{d})
			diags.Append(dd...)

			m.ApplyServerSideEncryptionByDefault = defaultList
		} else {
			defaultList, dd := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"kms_master_key_id": types.StringType,
					"sse_algorithm":     types.StringType,
				},
			}, []sseDefaultModel{})
			diags.Append(dd...)

			m.ApplyServerSideEncryptionByDefault = defaultList
		}

		if rule.BucketKeyEnabled != nil {
			m.BucketKeyEnabled = types.BoolValue(aws.ToBool(rule.BucketKeyEnabled))
		} else {
			m.BucketKeyEnabled = types.BoolNull()
		}

		models = append(models, m)
	}

	ruleType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"apply_server_side_encryption_by_default": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"kms_master_key_id": types.StringType,
						"sse_algorithm":     types.StringType,
					},
				},
			},
			"bucket_key_enabled": types.BoolType,
		},
	}

	set, dd := types.SetValueFrom(ctx, ruleType, models)
	diags.Append(dd...)

	return set, diags
}
