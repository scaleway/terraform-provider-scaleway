package messageq

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	messageqapi "github.com/scaleway/scaleway-sdk-go/api/messageq/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	providertypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/deployment.md
var deploymentDescription string

var (
	_ resource.Resource                = (*DeploymentResource)(nil)
	_ resource.ResourceWithConfigure   = (*DeploymentResource)(nil)
	_ resource.ResourceWithImportState = (*DeploymentResource)(nil)
	_ resource.ResourceWithIdentity    = (*DeploymentResource)(nil)
)

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	api  *messageqapi.API
	meta *meta.Meta
}

type deploymentResourceModel struct {
	Tags              types.List   `tfsdk:"tags"`
	Endpoints         types.List   `tfsdk:"endpoints"`
	Volume            types.Object `tfsdk:"volume"`
	PrivateNetwork    types.Object `tfsdk:"private_network"`
	ID                types.String `tfsdk:"id"`
	Region            types.String `tfsdk:"region"`
	ProjectID         types.String `tfsdk:"project_id"`
	Name              types.String `tfsdk:"name"`
	Version           types.String `tfsdk:"version"`
	NodeType          types.String `tfsdk:"node_type"`
	UserName          types.String `tfsdk:"user_name"`
	Password          types.String `tfsdk:"password"`
	PasswordWo        types.String `tfsdk:"password_wo"`
	Status            types.String `tfsdk:"status"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	NodeCount         types.Int64  `tfsdk:"node_count"`
	PasswordWoVersion types.Int64  `tfsdk:"password_wo_version"`
}

type deploymentResourceIdentityModel = framework.RegionalIdentity

type deploymentVolumeModel struct {
	Type     types.String `tfsdk:"type"`
	SizeInGB types.Int64  `tfsdk:"size_in_gb"`
}

type deploymentPrivateNetworkModel struct {
	PrivateNetworkID types.String `tfsdk:"private_network_id"`
}

func volumeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":       types.StringType,
		"size_in_gb": types.Int64Type,
	}
}

func privateNetworkAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"private_network_id": types.StringType,
	}
}

func endpointServiceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"port": types.Int64Type,
		"url":  types.StringType,
	}
}

func endpointAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                 types.StringType,
		"services":           types.ListType{ElemType: types.ObjectType{AttrTypes: endpointServiceAttrTypes()}},
		"public":             types.BoolType,
		"private_network_id": types.StringType,
	}
}

func (r *DeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_messageq_deployment"
}

func (r *DeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: deploymentDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the MessageQ deployment, in the `{region}/{id}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region the MessageQ deployment is in.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the MessageQ deployment belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Name of the MessageQ deployment",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of tags to apply",
			},
			"version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "MessageQ version to use",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"node_count": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Number of nodes. Can be updated via Upgrade without recreating the deployment",
			},
			"node_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of node",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Bootstrap username for the deployment. Prefer scaleway_messageq_user for additional users",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Bootstrap password for the deployment user. Only one of `password` or `password_wo` should be specified. Prefer scaleway_messageq_user for password rotation",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("password_wo")),
				},
			},
			// WriteOnly is incompatible with RequiresReplace, so password_wo_version carries
			// the ForceNew that makes a rotation attempt visible in the plan.
			"password_wo": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Bootstrap password for the deployment user in [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) mode. Only one of `password` or `password_wo` should be specified. `password_wo` will not be set in the Terraform state. Bootstrap credentials are immutable, so changing `password_wo_version` recreates the deployment: use scaleway_messageq_user to rotate a password in place",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("password")),
					stringvalidator.AlsoRequires(path.MatchRoot("password_wo_version")),
				},
			},
			"password_wo_version": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The version of the [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) password. To update the `password_wo`, you must also update the `password_wo_version`",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("password_wo")),
				},
			},
			"private_network": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Private network configuration",
				Attributes: map[string]schema.Attribute{
					"private_network_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "UUID of the Private Network",
						Validators: []validator.String{
							verify.IsStringUUIDOrUUIDWithRegion(),
						},
					},
				},
			},
			"volume": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Volume configuration",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Volume type (sbs_5k, sbs_15k)",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
						Validators: []validator.String{
							verify.ValidateEnumFramework[messageqapi.VolumeType](),
						},
					},
					"size_in_gb": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "Volume size in GB. Can be updated via Upgrade without recreating the deployment",
					},
				},
			},
			"endpoints": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of endpoints",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Endpoint ID",
						},
						"services": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "List of services",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Service name",
									},
									"port": schema.Int64Attribute{
										Computed:            true,
										MarkdownDescription: "Service port",
									},
									"url": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Service URL",
									},
								},
							},
						},
						"public": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the endpoint is public",
						},
						"private_network_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Private network ID if applicable",
						},
					},
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the deployment",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of deployment creation (RFC 3339 format)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of deployment last update (RFC 3339 format)",
			},
		},
	}
}

func (r *DeploymentResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultRegional()
}

func (r *DeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var config deploymentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

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

	name := plan.Name.ValueString()
	if name == "" {
		name = providertypes.NewRandomName("messageq")
	}

	createReq := &messageqapi.CreateDeploymentRequest{
		Region:    region,
		ProjectID: projectID,
		Name:      name,
		Version:   plan.Version.ValueString(),
		NodeCount: uint32(plan.NodeCount.ValueInt64()),
		NodeType:  plan.NodeType.ValueString(),
	}

	createReq.Tags = expandStringList(ctx, plan.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.UserName.IsNull() && !plan.UserName.IsUnknown() && plan.UserName.ValueString() != "" {
		createReq.UserName = providertypes.ExpandStringPtr(plan.UserName.ValueString())
	}

	password := plan.Password.ValueString()
	if !config.PasswordWoVersion.IsNull() && !config.PasswordWoVersion.IsUnknown() {
		password = config.PasswordWo.ValueString()
	}

	if password != "" {
		createReq.Password = &password
	}

	pnID := ""

	if !plan.Volume.IsNull() && !plan.Volume.IsUnknown() {
		var volume deploymentVolumeModel
		resp.Diagnostics.Append(plan.Volume.As(ctx, &volume, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		createReq.Volume = &messageqapi.Volume{
			Type:      messageqapi.VolumeType(volume.Type.ValueString()),
			SizeBytes: ExpandVolumeSizeBytes(int(volume.SizeInGB.ValueInt64())),
		}
	}

	if !plan.PrivateNetwork.IsNull() && !plan.PrivateNetwork.IsUnknown() {
		var pn deploymentPrivateNetworkModel
		resp.Diagnostics.Append(plan.PrivateNetwork.As(ctx, &pn, basetypes.ObjectAsOptions{})...)

		if resp.Diagnostics.HasError() {
			return
		}

		pnID = locality.ExpandID(pn.PrivateNetworkID.ValueString())
	}

	createReq.Endpoints = ExpandEndpointSpecsFromPrivateNetwork(pnID)

	deployment, err := r.api.CreateDeployment(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MessageQ deployment", err.Error())

		return
	}

	deployment, err = waitForDeployment(ctx, r.api, region, deployment.ID, defaultDeploymentTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment", err.Error())

		return
	}

	// Match the former SDKv2 Create→Read wait so VCR cassettes stay aligned.
	deployment, err = waitForDeployment(ctx, r.api, region, deployment.ID, defaultDeploymentReadTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed reading MessageQ deployment after create", err.Error())

		return
	}

	state := flattenDeployment(ctx, deployment, plan.PrivateNetwork, &resp.Diagnostics)
	state.Password = plan.Password
	state.PasswordWoVersion = plan.PasswordWoVersion
	state.UserName = plan.UserName
	state.PrivateNetwork = plan.PrivateNetwork

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(deployment.Region, deployment.ID))...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    deploymentResourceModel
		identity deploymentResourceIdentityModel
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

	resourceID := state.ID.ValueString()
	if identityAvailable {
		resourceID = identity.ID.ValueString()
	}

	region, id, err := regional.ParseID(resourceID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse MessageQ deployment ID", err.Error())

		return
	}

	deployment, err := waitForDeployment(ctx, r.api, region, id, defaultDeploymentReadTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read MessageQ deployment", err.Error())

		return
	}

	newState := flattenDeployment(ctx, deployment, state.PrivateNetwork, &resp.Diagnostics)
	newState.Password = state.Password
	newState.PasswordWoVersion = state.PasswordWoVersion
	newState.UserName = state.UserName
	newState.PrivateNetwork = state.PrivateNetwork

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(deployment.Region, deployment.ID))...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  deploymentResourceModel
		state deploymentResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse MessageQ deployment ID", err.Error())

		return
	}

	_, err = waitForDeployment(ctx, r.api, region, id, defaultDeploymentTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment", err.Error())

		return
	}

	if !plan.Name.Equal(state.Name) || !plan.Tags.Equal(state.Tags) {
		updateReq := &messageqapi.UpdateDeploymentRequest{
			Region:       region,
			DeploymentID: id,
		}

		if !plan.Name.Equal(state.Name) {
			updateReq.Name = providertypes.ExpandStringPtr(plan.Name.ValueString())
		}

		if !plan.Tags.Equal(state.Tags) {
			tags := expandStringList(ctx, plan.Tags, &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			if tags == nil {
				tags = []string{}
			}

			updateReq.Tags = &tags
		}

		_, err := r.api.UpdateDeployment(updateReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update MessageQ deployment", err.Error())

			return
		}

		_, err = waitForDeployment(ctx, r.api, region, id, defaultDeploymentTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Failed waiting for MessageQ deployment update", err.Error())

			return
		}
	}

	// UpgradeDeployment accepts precisely one of NodeCount or VolumeSizeBytes.
	if !plan.NodeCount.Equal(state.NodeCount) {
		nodeCount := uint32(plan.NodeCount.ValueInt64())

		_, err := r.api.UpgradeDeployment(&messageqapi.UpgradeDeploymentRequest{
			Region:       region,
			DeploymentID: id,
			NodeCount:    &nodeCount,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to upgrade MessageQ deployment node count", err.Error())

			return
		}

		_, err = waitForDeployment(ctx, r.api, region, id, defaultDeploymentTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Failed waiting for MessageQ deployment upgrade", err.Error())

			return
		}
	}

	volumeSizeChanged, sizeBytes := volumeSizeChange(ctx, plan.Volume, state.Volume, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if volumeSizeChanged {
		_, err := r.api.UpgradeDeployment(&messageqapi.UpgradeDeploymentRequest{
			Region:          region,
			DeploymentID:    id,
			VolumeSizeBytes: &sizeBytes,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to upgrade MessageQ deployment volume", err.Error())

			return
		}

		_, err = waitForDeployment(ctx, r.api, region, id, defaultDeploymentTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Failed waiting for MessageQ deployment volume upgrade", err.Error())

			return
		}
	}

	if !plan.PrivateNetwork.Equal(state.PrivateNetwork) {
		diags := r.updatePrivateNetwork(ctx, plan.PrivateNetwork, region, id)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	deployment, err := waitForDeployment(ctx, r.api, region, id, defaultDeploymentReadTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read MessageQ deployment after update", err.Error())

		return
	}

	newState := flattenDeployment(ctx, deployment, plan.PrivateNetwork, &resp.Diagnostics)
	newState.Password = plan.Password
	newState.PasswordWoVersion = plan.PasswordWoVersion
	newState.UserName = plan.UserName
	newState.PrivateNetwork = plan.PrivateNetwork

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(deployment.Region, deployment.ID))...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse MessageQ deployment ID", err.Error())

		return
	}

	_, err = waitForDeployment(ctx, r.api, region, id, defaultDeploymentTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment before delete", err.Error())

		return
	}

	_, err = r.api.DeleteDeployment(&messageqapi.DeleteDeploymentRequest{
		Region:       region,
		DeploymentID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete MessageQ deployment", err.Error())

		return
	}

	// Wait until the deployment is fully gone after Delete.
	_, err = waitForDeployment(ctx, r.api, region, id, defaultDeploymentTimeout)
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed waiting for MessageQ deployment deletion", err.Error())
	}
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *DeploymentResource) updatePrivateNetwork(
	ctx context.Context,
	privateNetwork types.Object,
	region scw.Region,
	id string,
) diag.Diagnostics {
	var diags diag.Diagnostics

	deployment, err := waitForDeployment(ctx, r.api, region, id, defaultDeploymentTimeout)
	if err != nil {
		diags.AddError("Failed waiting for MessageQ deployment", err.Error())

		return diags
	}

	desiredPrivate := false
	pnID := ""

	if !privateNetwork.IsNull() && !privateNetwork.IsUnknown() {
		var pn deploymentPrivateNetworkModel
		diags.Append(privateNetwork.As(ctx, &pn, basetypes.ObjectAsOptions{})...)

		if diags.HasError() {
			return diags
		}

		desiredPrivate = true
		pnID = locality.ExpandID(pn.PrivateNetworkID.ValueString())
	}

	// Public endpoints are never deleted here: only private endpoints are reconciled
	// against the desired private_network_id. Adding a private network therefore leaves
	// the existing public endpoint in place, which flattenDeployment hides from the
	// `endpoints` attribute.
	var deletedEndpointIDs []string

	hasDesiredEndpoint := false

	for _, endpoint := range deployment.Endpoints {
		if endpoint == nil {
			continue
		}

		if endpoint.Public != nil && endpoint.PrivateNetwork == nil {
			if !desiredPrivate {
				hasDesiredEndpoint = true
			}

			continue
		}

		if endpoint.PrivateNetwork == nil {
			continue
		}

		if desiredPrivate && endpoint.PrivateNetwork.PrivateNetworkID == pnID {
			hasDesiredEndpoint = true

			continue
		}

		err := r.api.DeleteEndpoint(&messageqapi.DeleteEndpointRequest{
			Region:     region,
			EndpointID: endpoint.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			diags.AddError("Failed to delete MessageQ endpoint", err.Error())

			return diags
		}

		deletedEndpointIDs = append(deletedEndpointIDs, endpoint.ID)
	}

	// DeleteEndpoint returns immediately but the removal is asynchronous, so the deleted
	// endpoints must be gone before creating the new one.
	if len(deletedEndpointIDs) > 0 {
		err := waitForEndpointsDeleted(ctx, r.api, region, id, deletedEndpointIDs, defaultDeploymentTimeout)
		if err != nil {
			diags.AddError("Failed waiting for MessageQ endpoints deletion", err.Error())

			return diags
		}
	}

	// Create the desired endpoint only if it doesn't already exist.
	if !hasDesiredEndpoint {
		var spec *messageqapi.EndpointSpec
		if desiredPrivate {
			spec = &messageqapi.EndpointSpec{
				PrivateNetwork: &messageqapi.EndpointSpecPrivateNetworkDetails{
					PrivateNetworkID: pnID,
				},
			}
		} else {
			spec = &messageqapi.EndpointSpec{
				Public: &messageqapi.EndpointSpecPublicDetails{},
			}
		}

		_, err := r.api.CreateEndpoint(&messageqapi.CreateEndpointRequest{
			Region:       region,
			DeploymentID: id,
			EndpointSpec: spec,
		}, scw.WithContext(ctx))
		if err != nil {
			diags.AddError("Failed to create MessageQ endpoint", err.Error())

			return diags
		}

		_, err = waitForDeployment(ctx, r.api, region, id, defaultDeploymentTimeout)
		if err != nil {
			diags.AddError("Failed waiting for MessageQ deployment after endpoint create", err.Error())

			return diags
		}
	}

	return diags
}

func volumeSizeChange(
	ctx context.Context,
	planVolume, stateVolume types.Object,
	diags *diag.Diagnostics,
) (bool, scw.Size) {
	if planVolume.IsNull() || planVolume.IsUnknown() {
		return false, 0
	}

	var planVol deploymentVolumeModel
	diags.Append(planVolume.As(ctx, &planVol, basetypes.ObjectAsOptions{})...)

	if diags.HasError() {
		return false, 0
	}

	if stateVolume.IsNull() || stateVolume.IsUnknown() {
		return true, ExpandVolumeSizeBytes(int(planVol.SizeInGB.ValueInt64()))
	}

	var stateVol deploymentVolumeModel
	diags.Append(stateVolume.As(ctx, &stateVol, basetypes.ObjectAsOptions{})...)

	if diags.HasError() {
		return false, 0
	}

	if planVol.SizeInGB.Equal(stateVol.SizeInGB) {
		return false, 0
	}

	return true, ExpandVolumeSizeBytes(int(planVol.SizeInGB.ValueInt64()))
}

func flattenDeployment(
	ctx context.Context,
	deployment *messageqapi.Deployment,
	privateNetwork types.Object,
	diags *diag.Diagnostics,
) deploymentResourceModel {
	model := deploymentResourceModel{
		ID:        types.StringValue(regional.NewIDString(deployment.Region, deployment.ID)),
		Region:    types.StringValue(deployment.Region.String()),
		ProjectID: types.StringValue(deployment.ProjectID),
		Name:      types.StringValue(deployment.Name),
		Version:   types.StringValue(deployment.Version),
		NodeCount: types.Int64Value(int64(deployment.NodeCount)),
		NodeType:  types.StringValue(deployment.NodeType),
		Status:    types.StringValue(string(deployment.Status)),
	}

	tagList, d := flattenStringList(ctx, deployment.Tags)
	diags.Append(d...)

	model.Tags = tagList

	if deployment.CreatedAt != nil {
		model.CreatedAt = types.StringValue(deployment.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringNull()
	}

	if deployment.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(deployment.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	model.Volume = flattenVolume(deployment.Volume, diags)
	model.Endpoints = flattenFilteredEndpoints(ctx, deployment.Endpoints, privateNetwork, diags)

	model.Password = types.StringNull()
	model.PasswordWo = types.StringNull()
	model.PasswordWoVersion = types.Int64Null()
	model.UserName = types.StringNull()
	model.PrivateNetwork = types.ObjectNull(privateNetworkAttrTypes())

	return model
}

func flattenVolume(volume *messageqapi.Volume, diags *diag.Diagnostics) types.Object {
	if volume == nil {
		return types.ObjectNull(volumeAttrTypes())
	}

	obj, d := types.ObjectValue(volumeAttrTypes(), map[string]attr.Value{
		"type":       types.StringValue(string(volume.Type)),
		"size_in_gb": types.Int64Value(int64(BytesToGB(volume.SizeBytes))),
	})
	diags.Append(d...)

	return obj
}

// The API may briefly return both public and private endpoints while a switch is
// in progress, and it keeps the public endpoint alive once a private one is added
// (see updatePrivateNetwork). In state, expose only the endpoint that matches the
// Terraform config, falling back to every endpoint when nothing matches so the
// attribute is never silently emptied.
func flattenFilteredEndpoints(
	ctx context.Context,
	allEndpoints []*messageqapi.Endpoint,
	privateNetwork types.Object,
	diags *diag.Diagnostics,
) types.List {
	var filteredEndpoints []*messageqapi.Endpoint

	if !privateNetwork.IsNull() && !privateNetwork.IsUnknown() {
		var pn deploymentPrivateNetworkModel
		diags.Append(privateNetwork.As(ctx, &pn, basetypes.ObjectAsOptions{})...)

		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: endpointAttrTypes()})
		}

		desiredPNID := locality.ExpandID(pn.PrivateNetworkID.ValueString())
		filteredEndpoints = nil

		for _, ep := range allEndpoints {
			if ep == nil || ep.PrivateNetwork == nil {
				continue
			}

			if ep.PrivateNetwork.PrivateNetworkID == desiredPNID {
				filteredEndpoints = append(filteredEndpoints, ep)
			}
		}

		if len(filteredEndpoints) == 0 {
			filteredEndpoints = allEndpoints
		}
	} else {
		filteredEndpoints = nil

		for _, ep := range allEndpoints {
			if ep == nil || ep.Public == nil || ep.PrivateNetwork != nil {
				continue
			}

			filteredEndpoints = append(filteredEndpoints, ep)
		}

		if len(filteredEndpoints) == 0 {
			filteredEndpoints = allEndpoints
		}
	}

	return flattenEndpointsList(filteredEndpoints, diags)
}

func flattenEndpointsList(endpoints []*messageqapi.Endpoint, diags *diag.Diagnostics) types.List {
	elemType := types.ObjectType{AttrTypes: endpointAttrTypes()}

	if len(endpoints) == 0 {
		return types.ListNull(elemType)
	}

	values := make([]attr.Value, 0, len(endpoints))

	for _, endpoint := range endpoints {
		if endpoint == nil {
			continue
		}

		services := types.ListNull(types.ObjectType{AttrTypes: endpointServiceAttrTypes()})

		if len(endpoint.Services) > 0 {
			serviceValues := make([]attr.Value, 0, len(endpoint.Services))

			for _, service := range endpoint.Services {
				svc, d := types.ObjectValue(endpointServiceAttrTypes(), map[string]attr.Value{
					"name": types.StringValue(service.Name),
					"port": types.Int64Value(int64(service.Port)),
					"url":  types.StringValue(service.URL),
				})
				diags.Append(d...)

				serviceValues = append(serviceValues, svc)
			}

			listVal, d := types.ListValue(types.ObjectType{AttrTypes: endpointServiceAttrTypes()}, serviceValues)
			diags.Append(d...)

			services = listVal
		}

		public := false
		privateNetworkID := types.StringNull()

		if endpoint.Public != nil {
			public = true
		}

		if endpoint.PrivateNetwork != nil {
			public = false
			privateNetworkID = types.StringValue(endpoint.PrivateNetwork.PrivateNetworkID)
		}

		obj, d := types.ObjectValue(endpointAttrTypes(), map[string]attr.Value{
			"id":                 types.StringValue(endpoint.ID),
			"services":           services,
			"public":             types.BoolValue(public),
			"private_network_id": privateNetworkID,
		})
		diags.Append(d...)

		values = append(values, obj)
	}

	listVal, d := types.ListValue(elemType, values)
	diags.Append(d...)

	return listVal
}

func expandStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)

	return result
}

func flattenStringList(ctx context.Context, items []string) (types.List, diag.Diagnostics) {
	if len(items) == 0 {
		return types.ListNull(types.StringType), nil
	}

	return types.ListValueFrom(ctx, types.StringType, items)
}
