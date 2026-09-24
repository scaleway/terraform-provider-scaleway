package messageq

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	providertypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/user.md
var userDescription string

var (
	_ resource.Resource                = (*UserResource)(nil)
	_ resource.ResourceWithConfigure   = (*UserResource)(nil)
	_ resource.ResourceWithImportState = (*UserResource)(nil)
	_ resource.ResourceWithIdentity    = (*UserResource)(nil)
)

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	api  *messageqapi.API
	meta *meta.Meta
}

type userResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Region            types.String `tfsdk:"region"`
	DeploymentID      types.String `tfsdk:"deployment_id"`
	Name              types.String `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	PasswordWo        types.String `tfsdk:"password_wo"`
	PasswordWoVersion types.Int64  `tfsdk:"password_wo_version"`
}

type userResourceIdentityModel struct {
	Region       types.String `tfsdk:"region"`
	DeploymentID types.String `tfsdk:"deployment_id"`
	Name         types.String `tfsdk:"name"`
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_messageq_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: userDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the MessageQ user, in the `{region}/{deployment_id}/{name}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Deployment on which the user is created",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "MessageQ user name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "MessageQ user password. Only one of `password` or `password_wo` should be specified.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("password"),
						path.MatchRoot("password_wo"),
					),
				},
			},
			"password_wo": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "MessageQ user password in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. To update the `password_wo`, you must also update the `password_wo_version`.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("password"),
						path.MatchRoot("password_wo"),
					),
					stringvalidator.AlsoRequires(path.MatchRoot("password_wo_version")),
				},
			},
			"password_wo_version": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The version of the [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) password. To update the `password_wo`, you must also update the `password_wo_version`.",
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("password_wo")),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region of the MessageQ deployment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *UserResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.CompositeRegional("deployment_id", "name")
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = messageqapi.NewAPI(r.meta.ScwClient())
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var config userResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fallbackRegion, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	region, deploymentID, err := RegionAndIDFromAttr(plan.DeploymentID.ValueString(), fallbackRegion)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse deployment_id", err.Error())

		return
	}

	_, err = waitForDeployment(ctx, r.api, region, deploymentID, defaultDeploymentTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment", err.Error())

		return
	}

	password := plan.Password.ValueString()
	if !config.PasswordWo.IsNull() && !config.PasswordWo.IsUnknown() && config.PasswordWo.ValueString() != "" {
		password = config.PasswordWo.ValueString()
	}

	user, err := r.api.CreateUser(&messageqapi.CreateUserRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Username:     plan.Name.ValueString(),
		Password:     password,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MessageQ user", err.Error())

		return
	}

	// Match the former SDKv2 Create→Read wait so VCR cassettes stay aligned.
	_, err = waitForDeployment(ctx, r.api, region, deploymentID, defaultDeploymentTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment after user create", err.Error())

		return
	}

	listRes, err := r.api.ListUsers(&messageqapi.ListUsersRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         providertypes.ExpandStringPtr(user.Username),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read MessageQ user after create", err.Error())

		return
	}

	if len(listRes.Users) == 0 {
		resp.Diagnostics.AddError(
			"Failed to read MessageQ user after create",
			"user "+user.Username+" was not found after create",
		)

		return
	}

	state := userResourceModel{
		ID:                types.StringValue(fmt.Sprintf("%s/%s/%s", region, deploymentID, user.Username)),
		Region:            types.StringValue(region.String()),
		DeploymentID:      plan.DeploymentID,
		Name:              types.StringValue(user.Username),
		Password:          plan.Password,
		PasswordWoVersion: plan.PasswordWoVersion,
		PasswordWo:        types.StringNull(),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, userResourceIdentityModel{
		Region:       types.StringValue(region.String()),
		DeploymentID: types.StringValue(deploymentID),
		Name:         types.StringValue(user.Username),
	})...)
}

// ResourceUserParseID extracts region, deployment ID and username from the resource
// identifier, whose format is "region/deployment_id/name".
func ResourceUserParseID(resourceID string) (region scw.Region, deploymentID string, userName string, err error) {
	idParts := identity.ParseMultiPartID(resourceID, "region", "deployment_id", "name")
	if idParts["region"] == "" || idParts["deployment_id"] == "" || idParts["name"] == "" {
		return "", "", "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}

	return scw.Region(idParts["region"]), idParts["deployment_id"], idParts["name"], nil
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    userResourceModel
		identity userResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() &&
		!identity.Region.IsNull() && !identity.Region.IsUnknown() &&
		!identity.DeploymentID.IsNull() && !identity.DeploymentID.IsUnknown() &&
		!identity.Name.IsNull() && !identity.Name.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		region       scw.Region
		deploymentID string
		userName     string
		err          error
	)

	if identityAvailable {
		region = scw.Region(identity.Region.ValueString())
		deploymentID = identity.DeploymentID.ValueString()
		userName = identity.Name.ValueString()
	} else {
		region, deploymentID, userName, err = ResourceUserParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse MessageQ user ID", err.Error())

			return
		}
	}

	_, err = waitForDeployment(ctx, r.api, region, deploymentID, defaultDeploymentTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment", err.Error())

		return
	}

	res, err := r.api.ListUsers(&messageqapi.ListUsersRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         &userName,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to list MessageQ users", err.Error())

		return
	}

	if len(res.Users) == 0 {
		tflog.Warn(ctx, fmt.Sprintf("couldn't find user with name: [%s]", userName))
		resp.State.RemoveResource(ctx)

		return
	}

	user := res.Users[0]

	newState := userResourceModel{
		ID:                types.StringValue(fmt.Sprintf("%s/%s/%s", region, deploymentID, user.Username)),
		Region:            types.StringValue(region.String()),
		DeploymentID:      state.DeploymentID,
		Name:              types.StringValue(user.Username),
		Password:          state.Password,
		PasswordWoVersion: state.PasswordWoVersion,
		PasswordWo:        types.StringNull(),
	}

	if newState.DeploymentID.IsNull() || newState.DeploymentID.IsUnknown() || newState.DeploymentID.ValueString() == "" {
		newState.DeploymentID = types.StringValue(regional.NewIDString(region, deploymentID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, userResourceIdentityModel{
		Region:       types.StringValue(region.String()),
		DeploymentID: types.StringValue(deploymentID),
		Name:         types.StringValue(user.Username),
	})...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  userResourceModel
		state userResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var config userResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, deploymentID, userName, err := ResourceUserParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse MessageQ user ID", err.Error())

		return
	}

	_, err = waitForDeployment(ctx, r.api, region, deploymentID, defaultDeploymentTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment", err.Error())

		return
	}

	// A write-only password is never in state, so its rotation is driven by the version.
	// It takes precedence so that switching from `password` to `password_wo` applies the
	// new write-only value rather than the cleared `password`.
	var newPassword *string

	if !plan.PasswordWoVersion.Equal(state.PasswordWoVersion) {
		if !config.PasswordWo.IsNull() && !config.PasswordWo.IsUnknown() {
			newPassword = providertypes.ExpandStringPtr(config.PasswordWo.ValueString())
		}
	}

	if newPassword == nil && !plan.Password.Equal(state.Password) {
		newPassword = providertypes.ExpandStringPtr(plan.Password.ValueString())
	}

	if newPassword != nil {
		_, err = r.api.UpdateUser(&messageqapi.UpdateUserRequest{
			Region:       region,
			DeploymentID: deploymentID,
			Username:     userName,
			Password:     newPassword,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update MessageQ user", err.Error())

			return
		}
	}

	// Match the former SDKv2 Update→Read wait so VCR cassettes stay aligned.
	_, err = waitForDeployment(ctx, r.api, region, deploymentID, defaultDeploymentTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment after user update", err.Error())

		return
	}

	listRes, err := r.api.ListUsers(&messageqapi.ListUsersRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Name:         &userName,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read MessageQ user after update", err.Error())

		return
	}

	if len(listRes.Users) == 0 {
		resp.Diagnostics.AddError(
			"Failed to read MessageQ user after update",
			"user "+userName+" was not found after update",
		)

		return
	}

	newState := userResourceModel{
		ID:                state.ID,
		Region:            types.StringValue(region.String()),
		DeploymentID:      plan.DeploymentID,
		Name:              types.StringValue(userName),
		Password:          plan.Password,
		PasswordWoVersion: plan.PasswordWoVersion,
		PasswordWo:        types.StringNull(),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, userResourceIdentityModel{
		Region:       types.StringValue(region.String()),
		DeploymentID: types.StringValue(deploymentID),
		Name:         types.StringValue(userName),
	})...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, deploymentID, userName, err := ResourceUserParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse MessageQ user ID", err.Error())

		return
	}

	_, err = waitForDeployment(ctx, r.api, region, deploymentID, defaultDeploymentTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment", err.Error())

		return
	}

	err = r.api.DeleteUser(&messageqapi.DeleteUserRequest{
		Region:       region,
		DeploymentID: deploymentID,
		Username:     userName,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to delete MessageQ user", err.Error())
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, deploymentID, userName, err := ResourceUserParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region.String())...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deployment_id"), regional.NewIDString(region, deploymentID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), userName)...)

	resp.Diagnostics.Append(resp.Identity.Set(ctx, userResourceIdentityModel{
		Region:       types.StringValue(region.String()),
		DeploymentID: types.StringValue(deploymentID),
		Name:         types.StringValue(userName),
	})...)
}
