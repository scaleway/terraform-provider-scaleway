package object

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	_ resource.Resource                = (*BucketLockConfigurationResource)(nil)
	_ resource.ResourceWithConfigure   = (*BucketLockConfigurationResource)(nil)
	_ resource.ResourceWithImportState = (*BucketLockConfigurationResource)(nil)
)

func NewBucketLockConfigurationResource() resource.Resource {
	return &BucketLockConfigurationResource{}
}

type BucketLockConfigurationResource struct {
	meta *meta.Meta
}

type bucketLockConfigurationModel struct {
	ID        types.String `tfsdk:"id"`
	Bucket    types.String `tfsdk:"bucket"`
	Rule      types.List   `tfsdk:"rule"`
	Region    types.String `tfsdk:"region"`
	ProjectID types.String `tfsdk:"project_id"`
}

type lockRuleModel struct {
	DefaultRetention types.List `tfsdk:"default_retention"`
}

type lockDefaultRetentionModel struct {
	Mode  types.String `tfsdk:"mode"`
	Days  types.Int64  `tfsdk:"days"`
	Years types.Int64  `tfsdk:"years"`
}

func (r *BucketLockConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_bucket_lock_configuration"
}

func (r *BucketLockConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Object Lock configuration for a Scaleway Object Storage bucket.",
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
			"region":     regional.SchemaAttributeComputed("The region you want to attach the resource to"),
			"project_id": projectIDFrameworkAttribute(),
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				MarkdownDescription: "Specifies the Object Lock rule for the specified object.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"default_retention": schema.ListNestedBlock{
							MarkdownDescription: "Default retention.",
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"mode": schema.StringAttribute{
										Required:            true,
										MarkdownDescription: "The default Object Lock retention mode you want to apply to new objects placed in the specified bucket.",
										Validators: []validator.String{
											stringvalidator.OneOf("GOVERNANCE", "COMPLIANCE"),
										},
									},
									"days": schema.Int64Attribute{
										Optional:            true,
										MarkdownDescription: "The number of days that you want to specify for the default retention period.",
										Validators: []validator.Int64{
											int64validator.ConflictsWith(path.MatchRoot("rule").AtListIndex(0).AtName("default_retention").AtListIndex(0).AtName("years")),
										},
									},
									"years": schema.Int64Attribute{
										Optional:            true,
										MarkdownDescription: "The number of years that you want to specify for the default retention period.",
										Validators: []validator.Int64{
											int64validator.ConflictsWith(path.MatchRoot("rule").AtListIndex(0).AtName("default_retention").AtListIndex(0).AtName("days")),
										},
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

func (r *BucketLockConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketLockConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketLockConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.resolveBucketAndClient(ctx, plan.Bucket, plan.Region, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	rule, err := expandBucketLockRule(ctx, plan.Rule, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to expand lock configuration rule", err.Error())

		return
	}

	_, err = s3Client.PutObjectLockConfiguration(ctx, &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3Types.ObjectLockConfiguration{
			ObjectLockEnabled: "Enabled",
			Rule:              rule,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating object bucket lock configuration", fmt.Errorf("error creating object bucket (%s) lock configuration: %w", bucket, err).Error())

		return
	}

	plan.ID = types.StringValue(regional.NewIDString(region, bucket))

	r.readInto(ctx, &plan, s3Client, region, bucket, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketLockConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketLockConfigurationModel
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

func (r *BucketLockConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketLockConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.clientFromState(ctx, plan.ID, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	rule, err := expandBucketLockRule(ctx, plan.Rule, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to expand lock configuration rule", err.Error())

		return
	}

	_, err = s3Client.PutObjectLockConfiguration(ctx, &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3Types.ObjectLockConfiguration{
			ObjectLockEnabled: s3Types.ObjectLockEnabledEnabled,
			Rule:              rule,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating object bucket lock configuration", fmt.Errorf("error updating Object bucket lock configuration (%s): %w", plan.ID.ValueString(), err).Error())

		return
	}

	r.readInto(ctx, &plan, s3Client, region, bucket, &resp.Diagnostics, true)

	// preserve resolved region from plan/state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketLockConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketLockConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, bucket, s3Client, err := r.clientFromState(ctx, state.ID, state.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	_, err = s3Client.PutObjectLockConfiguration(ctx, &s3.PutObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
		ObjectLockConfiguration: &s3Types.ObjectLockConfiguration{
			ObjectLockEnabled: s3Types.ObjectLockEnabledEnabled,
		},
	})

	if errors.As(err, new(*s3Types.NoSuchBucket)) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Error deleting object bucket lock configuration", fmt.Errorf("error deleting Object bucket lock configuration (%s): %w", state.ID.ValueString(), err).Error())

		return
	}
}

func (r *BucketLockConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *BucketLockConfigurationResource) resolveBucketAndClient(ctx context.Context, bucketAttr, regionAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
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

func (r *BucketLockConfigurationResource) clientFromState(ctx context.Context, idAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
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

func (r *BucketLockConfigurationResource) readInto(
	ctx context.Context,
	model *bucketLockConfigurationModel,
	s3Client *s3.Client,
	region scw.Region,
	bucket string,
	diags *diag.Diagnostics,
	isNewResource bool,
) {
	if region != "" {
		model.Region = types.StringValue(region.String())
	}

	output, err := s3Client.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{
		Bucket: aws.String(bucket),
	})
	if !isNewResource && errors.As(err, new(*s3Types.NoSuchBucket)) {
		tflog.Warn(ctx, fmt.Sprintf("Object Bucket Lock Configuration (%s) not found, removing from state", model.ID.ValueString()))
		diags.AddWarning("Bucket Lock Configuration not found", "Object Bucket Lock Configuration was not found, removing from state")

		return
	}

	if output == nil {
		if isNewResource {
			diags.AddError("Error reading object bucket lock configuration", fmt.Errorf("error reading object bucket lock configuration (%s): empty output", model.ID.ValueString()).Error())

			return
		}

		tflog.Warn(ctx, fmt.Sprintf("Object Bucket Lock Configuration (%s) not found, removing from state", model.ID.ValueString()))
		diags.AddWarning("Bucket Lock Configuration not found", "Object Bucket Lock Configuration was not found, removing from state")

		return
	}

	if model.Bucket.IsNull() {
		model.Bucket = types.StringValue(regional.NewIDString(region, bucket))
	}

	if output.ObjectLockConfiguration != nil {
		ruleList, ruleDiags := flattenBucketLockRule(ctx, output.ObjectLockConfiguration.Rule)
		diags.Append(ruleDiags...)

		model.Rule = ruleList
	} else {
		empty, _ := flattenBucketLockRule(ctx, nil)
		model.Rule = empty
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

func expandBucketLockRule(ctx context.Context, ruleList types.List, diags *diag.Diagnostics) (*s3Types.ObjectLockRule, error) {
	if ruleList.IsNull() || ruleList.IsUnknown() {
		return nil, nil
	}

	var rules []lockRuleModel

	ddiags := ruleList.ElementsAs(ctx, &rules, false)
	diags.Append(ddiags...)

	if diags.HasError() {
		return nil, nil
	}

	if len(rules) == 0 {
		return nil, nil
	}

	var retentionModels []lockDefaultRetentionModel

	rddiags := rules[0].DefaultRetention.ElementsAs(ctx, &retentionModels, false)
	diags.Append(rddiags...)

	if diags.HasError() {
		return nil, nil
	}

	if len(retentionModels) == 0 {
		return &s3Types.ObjectLockRule{}, nil
	}

	r := retentionModels[0]
	result := &s3Types.DefaultRetention{
		Mode: s3Types.ObjectLockRetentionMode(r.Mode.ValueString()),
	}

	if !r.Days.IsNull() && !r.Days.IsUnknown() && r.Days.ValueInt64() > 0 {
		result.Days = aws.Int32(int32(r.Days.ValueInt64()))
	}

	if !r.Years.IsNull() && !r.Years.IsUnknown() && r.Years.ValueInt64() > 0 {
		result.Years = aws.Int32(int32(r.Years.ValueInt64()))
	}

	return &s3Types.ObjectLockRule{DefaultRetention: result}, nil
}

func flattenBucketLockRule(ctx context.Context, rule *s3Types.ObjectLockRule) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	retentionModel := lockDefaultRetentionModel{}

	if rule != nil && rule.DefaultRetention != nil {
		retentionModel.Mode = types.StringValue(string(rule.DefaultRetention.Mode))

		if rule.DefaultRetention.Days != nil {
			retentionModel.Days = types.Int64Value(int64(aws.ToInt32(rule.DefaultRetention.Days)))
		}

		if rule.DefaultRetention.Years != nil {
			retentionModel.Years = types.Int64Value(int64(aws.ToInt32(rule.DefaultRetention.Years)))
		}
	}

	retentionList, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"mode":  types.StringType,
			"days":  types.Int64Type,
			"years": types.Int64Type,
		},
	}, []lockDefaultRetentionModel{retentionModel})
	diags.Append(d...)

	ruleModel := lockRuleModel{DefaultRetention: retentionList}
	ruleList, d := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"default_retention": retentionList.Type(ctx),
		},
	}, []lockRuleModel{ruleModel})
	diags.Append(d...)

	return ruleList, diags
}
