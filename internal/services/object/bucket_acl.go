package object

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/object_bucket_acl_resource.md
var bucketACLResourceDescription string

const (
	BucketACLSeparator = "/"

	// AWS predefined group URIs for bucket ACL grants
	AuthenticatedUsersURI = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
	AllUsersURI           = "http://acs.amazonaws.com/groups/global/AllUsers"
)

var (
	_ resource.Resource                = (*BucketACLResource)(nil)
	_ resource.ResourceWithConfigure   = (*BucketACLResource)(nil)
	_ resource.ResourceWithImportState = (*BucketACLResource)(nil)
)

func NewBucketACLResource() resource.Resource {
	return &BucketACLResource{}
}

type BucketACLResource struct {
	meta *meta.Meta
}

type bucketACLResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Bucket              types.String `tfsdk:"bucket"`
	Region              types.String `tfsdk:"region"`
	ProjectID           types.String `tfsdk:"project_id"`
	ACL                 types.String `tfsdk:"acl"`
	ExpectedBucketOwner types.String `tfsdk:"expected_bucket_owner"`
	AccessControlPolicy types.List   `tfsdk:"access_control_policy"`
}

type bucketACLAccessControlPolicyModel struct {
	Grant types.Set  `tfsdk:"grant"`
	Owner types.List `tfsdk:"owner"`
}

type bucketACLGrantModel struct {
	Grantee    types.List   `tfsdk:"grantee"`
	Permission types.String `tfsdk:"permission"`
}

type bucketACLGranteeModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	URI         types.String `tfsdk:"uri"`
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
}

type bucketACLOwnerModel struct {
	DisplayName types.String `tfsdk:"display_name"`
	ID          types.String `tfsdk:"id"`
}

func bucketACLGranteeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"display_name": types.StringType,
			"uri":          types.StringType,
			"id":           types.StringType,
			"type":         types.StringType,
		},
	}
}

func bucketACLGrantObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"grantee":    types.ListType{ElemType: bucketACLGranteeObjectType()},
			"permission": types.StringType,
		},
	}
}

func bucketACLOwnerObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"display_name": types.StringType,
			"id":           types.StringType,
		},
	}
}

func bucketACLAccessControlPolicyObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"grant": types.SetType{ElemType: bucketACLGrantObjectType()},
			"owner": types.ListType{ElemType: bucketACLOwnerObjectType()},
		},
	}
}

func (r *BucketACLResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_bucket_acl"
}

func (r *BucketACLResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: bucketACLResourceDescription,
		Description:         bucketACLResourceDescription, // Required for "TestProviderResourceDescriptionsAreNotEmpty"
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The regional ID of the bucket ACL, in the `{region}/{bucket}` or `{region}/{bucket}/{acl}` format.",
				PlanModifiers: []planmodifier.String{
					bucketACLIDPlanModifier{},
				},
			},
			"bucket": schema.StringAttribute{
				Required:    true,
				Description: "The bucket's name or regional ID.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 127),
				},
				PlanModifiers: []planmodifier.String{
					regional.LocalityPlanModifier(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"acl": schema.StringAttribute{
				Optional:    true,
				Description: "ACL of the bucket: either 'private', 'public-read', 'public-read-write' or 'authenticated-read'.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(s3Types.ObjectCannedACLPrivate),
						string(s3Types.ObjectCannedACLPublicRead),
						string(s3Types.ObjectCannedACLPublicReadWrite),
						string(s3Types.ObjectCannedACLAuthenticatedRead),
					),
				},
			},
			"expected_bucket_owner": schema.StringAttribute{
				Optional:    true,
				Description: "The project ID as owner.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region":     regional.SchemaAttributeComputed("The region you want to attach the resource to"),
			"project_id": account.ProjectIDSchemaFw("The project ID the bucket belongs to."),
		},
		// Blocks are used instead of nested attributes to preserve the SDKv2
		// configuration syntax (`access_control_policy { grant { ... } }`).
		// Blocks cannot be Computed, so the access_control_policy is stored in
		// state only when configured (see findBucketACL).
		Blocks: map[string]schema.Block{
			"access_control_policy": schema.ListNestedBlock{
				Description: "A configuration block that sets the ACL permissions for an object per grantee.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"grant": schema.SetNestedBlock{
							Description: "Grant",
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"grantee": schema.ListNestedBlock{
										Description: "Configuration block for the project being granted permissions.",
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"display_name": schema.StringAttribute{
													Computed:    true,
													Description: "Display name of the grantee to grant access to.",
												},
												"uri": schema.StringAttribute{
													Optional:    true,
													Description: "The uri of the grantee if you are granting permissions to a predefined group.",
													Validators: []validator.String{
														stringvalidator.OneOf(
															AuthenticatedUsersURI,
															AllUsersURI,
														),
													},
												},
												"id": schema.StringAttribute{
													Optional:    true,
													Description: "The project ID owner of the grantee.",
													Validators: []validator.String{
														verify.IsStringUUID(),
													},
												},
												"type": schema.StringAttribute{
													Optional:    true,
													Description: "Type of grantee. Valid values: `CanonicalUser`, `Group`",
													Validators: []validator.String{
														stringvalidator.OneOf(
															string(s3Types.TypeCanonicalUser),
															string(s3Types.TypeGroup),
														),
													},
												},
											},
										},
									},
								},
								Attributes: map[string]schema.Attribute{
									"permission": schema.StringAttribute{
										Required:    true,
										Description: "Logging permissions assigned to the grantee for the bucket.",
										Validators: []validator.String{
											stringvalidator.OneOf(
												string(s3Types.PermissionFullControl),
												string(s3Types.PermissionRead),
												string(s3Types.PermissionWrite),
												string(s3Types.PermissionReadAcp),
												string(s3Types.PermissionWriteAcp),
											),
										},
									},
								},
							},
						},
						"owner": schema.ListNestedBlock{
							Description: "Configuration block of the bucket project owner's display organization ID.",
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"display_name": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Description: "The project ID of the grantee.",
										Validators: []validator.String{
											verify.IsStringUUID(),
										},
									},
									"id": schema.StringAttribute{
										Required:    true,
										Description: "The display ID of the project.",
										Validators: []validator.String{
											verify.IsStringUUID(),
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

// ConfigValidators reports an error when both `acl` and `access_control_policy`
// are configured. It replaces the SDKv2 `ConflictsWith` behavior.
// It also enforces the SDKv2 `Required` semantics of the `owner` block, which
// Framework blocks (always optional) cannot express on their own.
func (r *BucketACLResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("acl"),
			path.MatchRoot("access_control_policy"),
		),
		bucketACLOwnerRequiredValidator{},
	}
}

type bucketACLOwnerRequiredValidator struct{}

func (bucketACLOwnerRequiredValidator) Description(_ context.Context) string {
	return "owner is required inside access_control_policy"
}

func (bucketACLOwnerRequiredValidator) MarkdownDescription(_ context.Context) string {
	return "owner is required inside access_control_policy"
}

func (bucketACLOwnerRequiredValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config bucketACLResourceModel
	if req.Config.Get(ctx, &config).HasError() {
		return
	}

	if config.AccessControlPolicy.IsNull() || len(config.AccessControlPolicy.Elements()) == 0 {
		return
	}

	var acps []bucketACLAccessControlPolicyModel
	if config.AccessControlPolicy.ElementsAs(ctx, &acps, false).HasError() {
		return
	}

	if len(acps) > 0 && acps[0].Owner.IsNull() {
		resp.Diagnostics.AddError(
			"Missing required owner block",
			"owner must be configured inside access_control_policy.",
		)
	}
}

func (r *BucketACLResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BucketACLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketACLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	regionalID := regional.ExpandID(plan.Bucket.ValueString())
	bucket := regionalID.ID

	if regionalID.Region != "" && regionalID.Region != region {
		region = regionalID.Region
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region.String(), projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	acl := ""

	if !plan.ACL.IsNull() && plan.ACL.ValueString() != "" {
		acl = plan.ACL.ValueString()
		input.ACL = s3Types.BucketCannedACL(acl)
	}

	if !plan.ExpectedBucketOwner.IsNull() && plan.ExpectedBucketOwner.ValueString() != "" {
		input.ExpectedBucketOwner = aws.String(plan.ExpectedBucketOwner.ValueString())
	}

	if !plan.AccessControlPolicy.IsNull() && len(plan.AccessControlPolicy.Elements()) > 0 {
		accessControlPolicy, diags := expandBucketACLAccessControlPolicy(ctx, plan.AccessControlPolicy)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		input.AccessControlPolicy = accessControlPolicy
	}

	out, err := s3Client.PutBucketAcl(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("error putting Object Storage ACL", err.Error())

		return
	}

	tflog.Debug(ctx, fmt.Sprintf("output: %v", out))

	plan.ID = types.StringValue(bucketACLCreateResourceID(region, bucket, acl))

	_, _, found, diags := r.findBucketACL(ctx, &plan, true)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Object Bucket ACL not found",
			fmt.Sprintf("Object Bucket ACL (%s) not found after creation", plan.ID.ValueString()),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketACLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketACLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucketName, found, diags := r.findBucketACL(ctx, &state, false)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.State.RemoveResource(ctx)

		return
	}

	// The bucket attribute keeps the value from the configuration during create
	// and update; reads normalize it to a regional ID.
	state.Bucket = types.StringValue(regional.NewIDString(scw.Region(region), bucketName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BucketACLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state bucketACLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, acl, err := locality.ParseLocalizedNestedOwnerID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse bucket ACL ID", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(state.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	input := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if !plan.ACL.Equal(state.ACL) {
		acl = ""

		if !plan.ACL.IsNull() && plan.ACL.ValueString() != "" {
			acl = plan.ACL.ValueString()
			input.ACL = s3Types.BucketCannedACL(acl)
		}
	}

	// Only re-send the access control policy when it is configured: blocks
	// cannot be Computed, so dropping it from state (e.g. when refreshing a
	// state written by the SDKv2 provider with only `acl` set) must not
	// touch the API.
	if !plan.AccessControlPolicy.IsNull() {
		accessControlPolicy, diags := expandBucketACLAccessControlPolicy(ctx, plan.AccessControlPolicy)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		input.AccessControlPolicy = accessControlPolicy
	}

	if input.ACL != "" || input.AccessControlPolicy != nil {
		_, err = s3Client.PutBucketAcl(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				"error updating object bucket ACL",
				fmt.Sprintf("error updating object bucket ACL (%s): %s", state.ID.ValueString(), err),
			)

			return
		}
	}

	// Set the new ACL value back in the resource ID.
	plan.ID = types.StringValue(bucketACLCreateResourceID(scw.Region(region), bucket, acl))

	_, _, found, diags := r.findBucketACL(ctx, &plan, true)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !found {
		resp.Diagnostics.AddError(
			"Object Bucket ACL not found",
			fmt.Sprintf("Object Bucket ACL (%s) not found after update", plan.ID.ValueString()),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BucketACLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketACLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, bucket, _, err := locality.ParseLocalizedNestedOwnerID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse bucket ACL ID", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(state.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create S3 client", err.Error())

		return
	}

	_, err = s3Client.PutBucketAcl(ctx, &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		ACL:    s3Types.BucketCannedACL(s3Types.ObjectCannedACLPrivate),
	})
	if err != nil {
		resp.Diagnostics.AddError("error putting bucket ACL", err.Error())

		return
	}

	resp.Diagnostics.AddWarning(
		"Deleting Object Bucket ACL resource resets ACL to private",
		"Deleting Object Bucket ACL resource resets the bucket's ACL to its default value: private.\nIf you wish to set it to something else, you should recreate a Bucket ACL resource with the `acl` field filled accordingly.",
	)
}

func (r *BucketACLResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// bucketACLIDPlanModifier keeps the prior ID in the plan when the `acl`
// attribute is unchanged. When `acl` changes, the ID must be recomputed
// (it embeds the ACL value), so the plan value stays unknown.
type bucketACLIDPlanModifier struct{}

func (bucketACLIDPlanModifier) Description(_ context.Context) string {
	return "Keeps the prior ID unless the acl attribute changes."
}

func (bucketACLIDPlanModifier) MarkdownDescription(_ context.Context) string {
	return "Keeps the prior ID unless the acl attribute changes."
}

func (bucketACLIDPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}

	var planACL, stateACL types.String

	if req.Plan.GetAttribute(ctx, path.Root("acl"), &planACL).HasError() {
		return
	}

	if req.State.GetAttribute(ctx, path.Root("acl"), &stateACL).HasError() {
		return
	}

	if planACL.Equal(stateACL) {
		resp.PlanValue = req.StateValue
	}
}

// findBucketACL reads the bucket ACL and updates the given state with the result.
// It returns false when the bucket does not exist.
func (r *BucketACLResource) findBucketACL(ctx context.Context, state *bucketACLResourceModel, isNewResource bool) (region, bucket string, found bool, diags diag.Diagnostics) {
	region, bucket, acl, err := locality.ParseLocalizedNestedOwnerID(state.ID.ValueString())
	if err != nil {
		diags.AddError("Failed to parse bucket ACL ID", err.Error())

		return "", "", false, diags
	}

	projectID, err := meta.ExtractFrameworkProjectID(state.ProjectID, r.meta.ScwClient())
	if err != nil {
		diags.AddError("Failed to resolve project ID", err.Error())

		return "", "", false, diags
	}

	s3Client, err := NewS3ClientFromMetaWithProjectID(ctx, r.meta, region, projectID)
	if err != nil {
		diags.AddError("Failed to create S3 client", err.Error())

		return "", "", false, diags
	}

	input := &s3.GetBucketAclInput{
		Bucket: aws.String(bucket),
	}

	if !state.ExpectedBucketOwner.IsNull() && state.ExpectedBucketOwner.ValueString() != "" {
		input.ExpectedBucketOwner = aws.String(state.ExpectedBucketOwner.ValueString())
	}

	output, err := s3Client.GetBucketAcl(ctx, input)

	if !isNewResource && errors.As(err, new(*s3Types.NoSuchBucket)) {
		tflog.Warn(ctx, fmt.Sprintf("[WARN] Object Bucket ACL (%s) not found, removing from state", state.ID.ValueString()))

		return "", "", false, nil
	}

	if err != nil {
		diags.AddError(
			"error getting object storage bucket ACL",
			fmt.Sprintf("error getting object storage bucket ACL (%s): %s", state.ID.ValueString(), err),
		)

		return "", "", false, diags
	}

	if output == nil {
		diags.AddError(
			"error getting object bucket ACL",
			fmt.Sprintf("error getting object bucket ACL (%s): empty output", state.ID.ValueString()),
		)

		return "", "", false, diags
	}

	if acl != "" {
		state.ACL = types.StringValue(acl)
	} else {
		state.ACL = types.StringNull()
	}

	accessControlPolicy, flattenDiags := flattenBucketACLAccessControlPolicy(ctx, output)
	diags.Append(flattenDiags...)

	if diags.HasError() {
		return "", "", false, diags
	}

	// Blocks cannot be Computed: store the access_control_policy only when it
	// is configured, or when importing (only the ID is known at that point).
	// An acl-only resource therefore does not keep the ACL grants in state,
	// unlike the previous SDKv2 implementation.
	if !state.AccessControlPolicy.IsNull() || state.Bucket.IsNull() {
		state.AccessControlPolicy = accessControlPolicy
	}

	if (state.ProjectID.IsNull() || state.ProjectID.IsUnknown()) && output.Owner != nil && output.Owner.ID != nil {
		state.ProjectID = types.StringValue(aws.ToString(NormalizeOwnerID(output.Owner.ID)))
	}

	state.Region = types.StringValue(region)

	return region, bucket, true, diags
}

// bucketACLCreateResourceID creates an ID string with the bucket name and optional ACL.
func bucketACLCreateResourceID(region scw.Region, bucket, acl string) string {
	if acl == "" {
		return regional.NewIDString(region, bucket)
	}

	return regional.NewIDString(region, strings.Join([]string{bucket, acl}, BucketACLSeparator))
}

func expandBucketACLAccessControlPolicy(ctx context.Context, l types.List) (*s3Types.AccessControlPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics

	var acps []bucketACLAccessControlPolicyModel
	diags.Append(l.ElementsAs(ctx, &acps, false)...)

	if diags.HasError() {
		return nil, diags
	}

	if len(acps) == 0 {
		return nil, diags
	}

	acp := acps[0]
	result := &s3Types.AccessControlPolicy{}

	if !acp.Grant.IsNull() && len(acp.Grant.Elements()) > 0 {
		grants, grantDiags := expandBucketACLGrants(ctx, acp.Grant)
		diags.Append(grantDiags...)

		if diags.HasError() {
			return nil, diags
		}

		result.Grants = grants
	}

	if !acp.Owner.IsNull() && len(acp.Owner.Elements()) > 0 {
		owner, ownerDiags := expandBucketACLOwner(ctx, acp.Owner)
		diags.Append(ownerDiags...)

		if diags.HasError() {
			return nil, diags
		}

		result.Owner = owner
	}

	return result, diags
}

func expandBucketACLGrants(ctx context.Context, s types.Set) ([]s3Types.Grant, diag.Diagnostics) {
	var diags diag.Diagnostics

	var grants []bucketACLGrantModel
	diags.Append(s.ElementsAs(ctx, &grants, false)...)

	if diags.HasError() {
		return nil, diags
	}

	results := make([]s3Types.Grant, 0, len(grants))

	for _, grant := range grants {
		g := s3Types.Grant{}

		if !grant.Grantee.IsNull() && len(grant.Grantee.Elements()) > 0 {
			grantee, granteeDiags := expandBucketACLGrantee(ctx, grant.Grantee)
			diags.Append(granteeDiags...)

			if diags.HasError() {
				return nil, diags
			}

			g.Grantee = grantee
		}

		if !grant.Permission.IsNull() && !grant.Permission.IsUnknown() {
			g.Permission = s3Types.Permission(grant.Permission.ValueString())
		}

		results = append(results, g)
	}

	return results, diags
}

func expandBucketACLGrantee(ctx context.Context, l types.List) (*s3Types.Grantee, diag.Diagnostics) {
	var diags diag.Diagnostics

	var grantees []bucketACLGranteeModel
	diags.Append(l.ElementsAs(ctx, &grantees, false)...)

	if diags.HasError() {
		return nil, diags
	}

	if len(grantees) == 0 {
		return nil, diags
	}

	grantee := grantees[0]
	result := &s3Types.Grantee{}

	if !grantee.ID.IsNull() && !grantee.ID.IsUnknown() {
		result.ID = buildBucketOwnerID(aws.String(grantee.ID.ValueString()))
	}

	if !grantee.URI.IsNull() && !grantee.URI.IsUnknown() {
		result.URI = aws.String(grantee.URI.ValueString())
	}

	if !grantee.Type.IsNull() && !grantee.Type.IsUnknown() {
		result.Type = s3Types.Type(grantee.Type.ValueString())
	}

	if result.Type == s3Types.TypeCanonicalUser && result.ID == nil {
		diags.AddError("Invalid grantee", "id is required when grantee type is CanonicalUser")

		return nil, diags
	}

	if result.Type == s3Types.TypeGroup && result.URI == nil {
		diags.AddError("Invalid grantee", "uri is required when grantee type is Group")

		return nil, diags
	}

	return result, diags
}

func expandBucketACLOwner(ctx context.Context, l types.List) (*s3Types.Owner, diag.Diagnostics) {
	var diags diag.Diagnostics

	var owners []bucketACLOwnerModel
	diags.Append(l.ElementsAs(ctx, &owners, false)...)

	if diags.HasError() {
		return nil, diags
	}

	if len(owners) == 0 {
		return nil, diags
	}

	owner := owners[0]
	result := &s3Types.Owner{}

	if !owner.DisplayName.IsNull() && !owner.DisplayName.IsUnknown() {
		result.DisplayName = buildBucketOwnerID(aws.String(owner.DisplayName.ValueString()))
	}

	if !owner.ID.IsNull() && !owner.ID.IsUnknown() {
		result.ID = buildBucketOwnerID(aws.String(owner.ID.ValueString()))
	}

	return result, diags
}

func flattenBucketACLAccessControlPolicy(ctx context.Context, output *s3.GetBucketAclOutput) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrs := map[string]attr.Value{
		"grant": types.SetNull(bucketACLGrantObjectType()),
		"owner": types.ListNull(bucketACLOwnerObjectType()),
	}

	if len(output.Grants) > 0 {
		grants, grantDiags := flattenBucketACLGrants(ctx, output.Grants)
		diags.Append(grantDiags...)

		if diags.HasError() {
			return types.ListNull(bucketACLAccessControlPolicyObjectType()), diags
		}

		attrs["grant"] = grants
	}

	if output.Owner != nil {
		owner, ownerDiags := flattenBucketACLOwner(ctx, output.Owner)
		diags.Append(ownerDiags...)

		if diags.HasError() {
			return types.ListNull(bucketACLAccessControlPolicyObjectType()), diags
		}

		attrs["owner"] = owner
	}

	acpObject, d := types.ObjectValue(bucketACLAccessControlPolicyObjectType().AttrTypes, attrs)
	diags.Append(d...)

	if diags.HasError() {
		return types.ListNull(bucketACLAccessControlPolicyObjectType()), diags
	}

	acpList, d := types.ListValueFrom(ctx, bucketACLAccessControlPolicyObjectType(), []attr.Value{acpObject})
	diags.Append(d...)

	return acpList, diags
}

func flattenBucketACLGrants(ctx context.Context, grants []s3Types.Grant) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	elems := make([]attr.Value, 0, len(grants))

	for _, grant := range grants {
		if grant.Grantee == nil && grant.Permission == "" {
			continue
		}

		attrs := map[string]attr.Value{
			"grantee":    types.ListNull(bucketACLGranteeObjectType()),
			"permission": types.StringNull(),
		}

		if grant.Grantee != nil {
			grantee, d := flattenBucketACLGrantee(ctx, grant.Grantee)
			diags.Append(d...)

			if diags.HasError() {
				return types.SetNull(bucketACLGrantObjectType()), diags
			}

			attrs["grantee"] = grantee
		}

		if grant.Permission != "" {
			attrs["permission"] = types.StringValue(string(grant.Permission))
		}

		grantObject, d := types.ObjectValue(bucketACLGrantObjectType().AttrTypes, attrs)
		diags.Append(d...)

		if diags.HasError() {
			return types.SetNull(bucketACLGrantObjectType()), diags
		}

		elems = append(elems, grantObject)
	}

	set, d := types.SetValue(bucketACLGrantObjectType(), elems)
	diags.Append(d...)

	return set, diags
}

func flattenBucketACLGrantee(ctx context.Context, grantee *s3Types.Grantee) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrs := map[string]attr.Value{
		"display_name": types.StringNull(),
		"uri":          types.StringNull(),
		"id":           types.StringNull(),
		"type":         types.StringNull(),
	}

	if grantee.DisplayName != nil {
		attrs["display_name"] = types.StringValue(aws.ToString(NormalizeOwnerID(grantee.DisplayName)))
	}

	if grantee.URI != nil {
		attrs["uri"] = types.StringValue(*grantee.URI)
	}

	if grantee.ID != nil {
		attrs["id"] = types.StringValue(aws.ToString(NormalizeOwnerID(grantee.ID)))
	}

	if grantee.Type != "" {
		attrs["type"] = types.StringValue(string(grantee.Type))
	}

	granteeObject, d := types.ObjectValue(bucketACLGranteeObjectType().AttrTypes, attrs)
	diags.Append(d...)

	if diags.HasError() {
		return types.ListNull(bucketACLGranteeObjectType()), diags
	}

	granteeList, d := types.ListValueFrom(ctx, bucketACLGranteeObjectType(), []attr.Value{granteeObject})
	diags.Append(d...)

	return granteeList, diags
}

func flattenBucketACLOwner(ctx context.Context, owner *s3Types.Owner) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrs := map[string]attr.Value{
		"display_name": types.StringNull(),
		"id":           types.StringNull(),
	}

	if owner.DisplayName != nil {
		attrs["display_name"] = types.StringValue(aws.ToString(NormalizeOwnerID(owner.DisplayName)))
	}

	if owner.ID != nil {
		attrs["id"] = types.StringValue(aws.ToString(NormalizeOwnerID(owner.ID)))
	}

	ownerObject, d := types.ObjectValue(bucketACLOwnerObjectType().AttrTypes, attrs)
	diags.Append(d...)

	if diags.HasError() {
		return types.ListNull(bucketACLOwnerObjectType()), diags
	}

	ownerList, d := types.ListValueFrom(ctx, bucketACLOwnerObjectType(), []attr.Value{ownerObject})
	diags.Append(d...)

	return ownerList, diags
}
