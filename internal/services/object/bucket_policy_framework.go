package object

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// Reuses the embedded description from bucket_policy.go (bucketPolicyResourceDescription).

var (
	_ resource.Resource                = (*BucketPolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*BucketPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*BucketPolicyResource)(nil)
)

func NewBucketPolicyResource() resource.Resource {
	return &BucketPolicyResource{}
}

type BucketPolicyResource struct {
	meta *meta.Meta
}

type bucketPolicyResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Bucket    types.String `tfsdk:"bucket"`
	Policy    types.String `tfsdk:"policy"`
	Region    types.String `tfsdk:"region"`
	ProjectID types.String `tfsdk:"project_id"`
}

func (r *BucketPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_bucket_policy"
}

func (r *BucketPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: bucketPolicyResourceDescription,
		Description:         bucketPolicyResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The regional ID of the bucket policy, in the `{region}/{bucket}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "The bucket's name or regional ID.",
			},
			"policy": schema.StringAttribute{
				Required:    true,
				Description: "The text of the policy.",
				PlanModifiers: []planmodifier.String{
					policyEquivalentPlanModifier{},
				},
			},
			"region": regional.SchemaAttributeComputed("The region you want to attach the resource to"),
			"project_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The project ID the bucket belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *BucketPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.putPolicy(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.putPolicy(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucketName, projectID, err := r.resolveFromState(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve bucket policy state", err.Error())

		return
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("scw object bucket: %s, delete policy", bucketName))

	_, err = s3Client.DeleteBucketPolicy(ctx, &s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucketName),
	})

	if IsS3Err(err, ErrCodeNoSuchBucket, "") || err == nil {
		return
	}

	resp.Diagnostics.AddError("error deleting SCW Object policy", err.Error())
}

func (r *BucketPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucketName, projectID, err := r.resolveFromState(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve bucket policy state", err.Error())

		return
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	tflog.Debug(ctx, "[DEBUG] SCW bucket policy, read for bucket: "+state.ID.ValueString())

	pol, err := s3Client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucketName),
	})

	if IsS3Err(err, ErrCodeNoSuchBucketPolicy, "") || IsS3Err(err, ErrCodeNoSuchBucket, "") {
		tflog.Warn(ctx, fmt.Sprintf("[WARN] SCW Bucket Policy (%s) not found, removing from state", state.ID.ValueString()))
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to read bucket policy", err.Error())

		return
	}

	remotePolicy := ""
	if pol != nil && pol.Policy != nil {
		remotePolicy = aws.ToString(pol.Policy)
	}

	policyToSet, err := SecondJSONUnlessEquivalent(state.Policy.ValueString(), remotePolicy)
	if err != nil {
		resp.Diagnostics.AddError("while setting policy", "encountered: "+err.Error())

		return
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		resp.Diagnostics.AddError("invalid policy JSON", fmt.Sprintf("policy (%s) is an invalid JSON: %s", policyToSet, err.Error()))

		return
	}

	state.Policy = types.StringValue(policyToSet)

	// Refresh project_id from the bucket ACL owner.
	acl, err := s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		resp.Diagnostics.AddWarning("Failed to read bucket ACL", err.Error())
	} else if acl != nil && acl.Owner != nil && acl.Owner.ID != nil {
		state.ProjectID = types.StringValue(aws.ToString(NormalizeOwnerID(acl.Owner.ID)))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// putPolicy applies (create or update) the bucket policy and returns the resulting state.
// It mirrors the SDKv2 Create/Update behavior, including the bucket-region override and
// the retry-on-MalformedPolicy logic.
func (r *BucketPolicyResource) putPolicy(ctx context.Context, plan *bucketPolicyResourceModel) (bucketPolicyResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	region, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		diags.AddError("Failed to resolve region", err.Error())

		return bucketPolicyResourceModel{}, diags
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		diags.AddError("Failed to resolve project ID", err.Error())

		return bucketPolicyResourceModel{}, diags
	}

	regionalID := regional.ExpandID(plan.Bucket.ValueString())
	bucket := regionalID.ID
	bucketRegion := regionalID.Region

	if bucketRegion != "" && bucketRegion != region {
		region = bucketRegion
	}

	policy, err := structure.NormalizeJsonString(plan.Policy.ValueString())
	if err != nil {
		diags.AddError("invalid policy JSON", fmt.Sprintf("policy (%s) is an invalid JSON: %s", plan.Policy.ValueString(), err.Error()))

		return bucketPolicyResourceModel{}, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] SCW bucket: %s, put policy: %s", bucket, policy))

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		diags.AddError("Failed to create S3 client", err.Error())

		return bucketPolicyResourceModel{}, diags
	}

	params := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	err = retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		_, err := s3Client.PutBucketPolicy(ctx, params)
		if IsS3Err(err, "MalformedPolicy", "") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if TimedOut(err) {
		_, err = s3Client.PutBucketPolicy(ctx, params)
	}

	if err != nil {
		diags.AddError("error putting SCW bucket policy", err.Error())

		return bucketPolicyResourceModel{}, diags
	}

	// State mirrors the planned value: we just applied this policy, so the policy
	// attribute is set from the plan (not re-read from the API). Re-reading here
	// would risk a post-apply inconsistency error under VCR replay, where the
	// recorded GetBucketPolicy response carries the recorded (different) bucket
	// name. The next Read (refresh) reconciles any drift, and the policy plan
	// modifier suppresses semantically-equivalent diffs.
	state := bucketPolicyResourceModel{
		ID:        types.StringValue(regional.NewIDString(region, bucket)),
		Bucket:    types.StringValue(regionalID.String()),
		Policy:    types.StringValue(policy),
		Region:    types.StringValue(region.String()),
		ProjectID: plan.ProjectID,
	}

	// Refresh only the computed project_id from the bucket ACL owner.
	acl, err := s3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		diags.AddWarning("Failed to read bucket ACL", err.Error())
	} else if acl != nil && acl.Owner != nil && acl.Owner.ID != nil {
		state.ProjectID = types.StringValue(aws.ToString(NormalizeOwnerID(acl.Owner.ID)))
	}

	return state, diags
}

// resolveFromState extracts the region, bucket name and project ID from the resource state.
func (r *BucketPolicyResource) resolveFromState(_ context.Context, state *bucketPolicyResourceModel) (scw.Region, string, string, error) {
	region, bucketName, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		// Fall back to parsing the bucket attribute if the ID is not a regional ID.
		regionalID := regional.ExpandID(state.Bucket.ValueString())
		bucketName = regionalID.ID

		region = regionalID.Region
		if region == "" {
			region, err = meta.ExtractFrameworkRegion(state.Region, r.meta.ScwClient())
			if err != nil {
				return "", "", "", err
			}
		}
	}

	projectID, err := meta.ExtractFrameworkProjectID(state.ProjectID, r.meta.ScwClient())
	if err != nil {
		return "", "", "", err
	}

	return region, bucketName, projectID, nil
}

// policyEquivalentPlanModifier suppresses plan diffs when the configured policy JSON
// is semantically equivalent to the prior state. This is the Plugin Framework
// counterpart of the SDKv2 SuppressEquivalentPolicyDiffs diff suppressor.
type policyEquivalentPlanModifier struct{}

func (policyEquivalentPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs for semantically equivalent bucket policy JSON."
}

func (policyEquivalentPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Suppresses diffs for semantically equivalent bucket policy JSON."
}

func (policyEquivalentPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	old := req.StateValue.ValueString()
	newP := req.PlanValue.ValueString()

	if strings.TrimSpace(old) == "" && strings.TrimSpace(newP) == "" ||
		strings.TrimSpace(old) == "{}" && strings.TrimSpace(newP) == "" ||
		strings.TrimSpace(old) == "" && strings.TrimSpace(newP) == "{}" ||
		strings.TrimSpace(old) == "{}" && strings.TrimSpace(newP) == "{}" {
		resp.PlanValue = req.StateValue

		return
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(old, newP)
	if err != nil {
		return
	}

	if equivalent {
		resp.PlanValue = req.StateValue
	}
}
