package object

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

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

type bucketPolicyModel struct {
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
		MarkdownDescription: "Manages a Scaleway Object Storage bucket policy.",
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
				},
			},
			"policy": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The text of the policy.",
				PlanModifiers: []planmodifier.String{
					equivalentPolicyPlanModifier{},
				},
			},
			"region":     regional.SchemaAttributeComputed("The region you want to attach the resource to"),
			"project_id": projectIDFrameworkAttribute(),
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
	var plan bucketPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.resolveBucketAndClient(ctx, plan.Bucket, plan.Region, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	policy, err := normalizePolicyJSON(plan.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid policy JSON", err.Error())

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] SCW bucket: %s, put policy: %s", bucket, policy))

	params := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	err = retryPutBucketPolicy(ctx, s3Client, params)
	if err != nil {
		resp.Diagnostics.AddError("Error putting SCW bucket policy", err.Error())

		return
	}

	plan.ID = types.StringValue(regional.NewIDString(region, bucket))

	r.readInto(ctx, &plan, s3Client, region, bucket, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketPolicyModel
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

func (r *BucketPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, s3Client, err := r.resolveBucketAndClient(ctx, plan.Bucket, plan.Region, plan.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	policy, err := normalizePolicyJSON(plan.Policy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid policy JSON", err.Error())

		return
	}

	params := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	err = retryPutBucketPolicy(ctx, s3Client, params)
	if err != nil {
		resp.Diagnostics.AddError("Error putting SCW bucket policy", err.Error())

		return
	}

	plan.ID = types.StringValue(regional.NewIDString(region, bucket))

	r.readInto(ctx, &plan, s3Client, region, bucket, &resp.Diagnostics, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, bucket, s3Client, err := r.clientFromState(ctx, state.ID, state.ProjectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve S3 client", err.Error())

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("scw object bucket: %s, delete policy", bucket))

	_, err = s3Client.DeleteBucketPolicy(ctx, &s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucket) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Error deleting SCW Object policy", err.Error())

		return
	}
}

func (r *BucketPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// resolveBucketAndClient resolves the region and bucket name from the bucket
// attribute (which may be a bare name or a regional ID), builds an S3 client
// scoped to the resolved project ID.
func (r *BucketPolicyResource) resolveBucketAndClient(ctx context.Context, bucketAttr, regionAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
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

// clientFromState builds an S3 client for a resource being read/deleted,
// deriving the region and bucket name from the stored regional ID.
func (r *BucketPolicyResource) clientFromState(ctx context.Context, idAttr, projectIDAttr types.String) (scw.Region, string, *s3.Client, error) {
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

func (r *BucketPolicyResource) readInto(
	ctx context.Context,
	model *bucketPolicyModel,
	s3Client *s3.Client,
	region scw.Region,
	bucket string,
	diags *diag.Diagnostics,
	isNewResource bool,
) {
	model.Region = types.StringValue(region.String())

	pol, err := s3Client.GetBucketPolicy(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})

	if !isNewResource && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchBucketPolicy, ErrCodeNoSuchBucket) {
		tflog.Warn(ctx, fmt.Sprintf("[WARN] SCW Bucket Policy (%s) not found, removing from state", model.ID.ValueString()))
		diags.AddWarning("Bucket Policy not found", "Bucket policy was not found, removing from state")

		return
	}

	var apiPolicy string
	if err == nil && pol.Policy != nil {
		apiPolicy = aws.ToString(pol.Policy)
	}

	policyToSet, err := SecondJSONUnlessEquivalent(model.Policy.ValueString(), apiPolicy)
	if err != nil {
		diags.AddError("Failed to normalize policy", err.Error())

		return
	}

	policyToSet, err = normalizePolicyJSON(policyToSet)
	if err != nil {
		diags.AddError("Invalid policy JSON", err.Error())

		return
	}

	model.Policy = types.StringValue(policyToSet)

	if model.Bucket.IsNull() {
		model.Bucket = types.StringValue(regional.NewIDString(region, bucket))
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

func retryPutBucketPolicy(ctx context.Context, s3Client *s3.Client, params *s3.PutBucketPolicyInput) error {
	var lastErr error

	for retry := 0; retry < 6; retry++ {
		_, err := s3Client.PutBucketPolicy(ctx, params)
		if err == nil {
			return nil
		}

		lastErr = err

		if tfawserr.ErrCodeEquals(err, "MalformedPolicy") {
			time.Sleep(10 * time.Second)

			continue
		}

		return err
	}

	return lastErr
}

// equivalentPolicyPlanModifier suppresses diffs between semantically equivalent
// bucket policies, mirroring the SDKv2 SuppressEquivalentPolicyDiffs.
type equivalentPolicyPlanModifier struct{}

func (m equivalentPolicyPlanModifier) Description(_ context.Context) string {
	return "Suppresses diffs between semantically equivalent JSON policies."
}

func (m equivalentPolicyPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m equivalentPolicyPlanModifier) PlanModifyString(
	_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse,
) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}

	if SuppressEquivalentPolicyDiffs("", req.StateValue.ValueString(), req.PlanValue.ValueString(), nil) {
		resp.PlanValue = req.StateValue
	}
}
