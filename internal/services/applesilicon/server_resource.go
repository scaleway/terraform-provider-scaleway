package applesilicon

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	framework "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/server.md
var serverDescription string

var (
	_ resource.Resource                = (*serverResource)(nil)
	_ resource.ResourceWithConfigure   = (*serverResource)(nil)
	_ resource.ResourceWithImportState = (*serverResource)(nil)
	_ resource.ResourceWithIdentity    = (*serverResource)(nil)
)

func NewServerResource() resource.Resource {
	return &serverResource{}
}

type serverResource struct {
	api  *applesilicon.API
	meta *meta.Meta
}

type serverResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	EnableVpc       types.Bool   `tfsdk:"enable_vpc"`
	Commitment      types.String `tfsdk:"commitment"`
	RunnerIDs       types.List   `tfsdk:"runner_ids"`
	OsID            types.String `tfsdk:"os_id"`
	PublicBandwidth types.Int64  `tfsdk:"public_bandwidth"`
	PrivateNetwork  types.Set    `tfsdk:"private_network"`
	IP              types.String `tfsdk:"ip"`
	PrivateIPs      types.List   `tfsdk:"private_ips"`
	VncURL          types.String `tfsdk:"vnc_url"`
	State           types.String `tfsdk:"state"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	DeletableAt     types.String `tfsdk:"deletable_at"`
	VpcStatus       types.String `tfsdk:"vpc_status"`
	Password        types.String `tfsdk:"password"`
	Username        types.String `tfsdk:"username"`
	Zone            types.String `tfsdk:"zone"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	ProjectID       types.String `tfsdk:"project_id"`
}

type serverResourceIdentityModel = framework.ZonalIdentity

type privateNetworkBlockModel struct {
	ID        types.String `tfsdk:"id"`
	IpamIPIDs types.List   `tfsdk:"ipam_ip_ids"`
	Vlan      types.Int64  `tfsdk:"vlan"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type privateIPsBlockModel struct {
	ID      types.String `tfsdk:"id"`
	Address types.String `tfsdk:"address"`
}

func privateNetworkObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":          types.StringType,
			"ipam_ip_ids": types.ListType{ElemType: types.StringType},
			"vlan":        types.Int64Type,
			"status":      types.StringType,
			"created_at":  types.StringType,
			"updated_at":  types.StringType,
		},
	}
}

func privateIPsObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":      types.StringType,
			"address": types.StringType,
		},
	}
}

func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apple_silicon_server"
}

func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: serverDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Name of the server",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of the server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enable_vpc": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether or not to enable VPC access",
			},
			"commitment": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("duration_24h"),
				MarkdownDescription: "The commitment period of the server",
				Validators: []validator.String{
					verify.ValidateEnumFramework[applesilicon.CommitmentType](),
				},
			},
			"runner_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "List of runner ids attach to the server",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(verify.IsStringUUIDOrUUIDWithZone()),
				},
			},
			"os_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The OS ID of the server",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithZone(),
				},
			},
			"public_bandwidth": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The public bandwidth of the server in bits per second",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "IPv4 address of the server",
			},
			"vnc_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "VNC url use to connect remotely to the desktop GUI",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The state of the server",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time of the creation of the server",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time of the last update of the server",
			},
			"deletable_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The minimal date and time on which you can delete this server due to Apple licence",
			},
			"vpc_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The VPC status of the server",
			},
			"password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The password of the server",
			},
			"username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The username of the server",
			},
			"zone": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The zone you want to attach the resource to",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization_id you want to attach the resource to",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project_id you want to attach the resource to",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_ips": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of private IPv4 and IPv6 addresses associated with the server",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the IP address resource",
						},
						"address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The private IP address",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"private_network": schema.SetNestedBlock{
				MarkdownDescription: "The private networks to attach to the server",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The private network ID",
							Validators: []validator.String{
								verify.IsStringUUIDOrUUIDWithRegion(),
							},
						},
						"ipam_ip_ids": schema.ListAttribute{
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "List of IPAM IP IDs to attach to the server",
							Validators: []validator.List{
								listvalidator.ValueStringsAre(verify.IsStringUUIDOrUUIDWithRegion()),
							},
						},
						"vlan": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The VLAN ID associated to the private network",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The private network status",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The date and time of the creation of the private network",
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The date and time of the last update of the private network",
						},
					},
				},
			},
		},
	}
}

func (r *serverResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultZonal()
}

func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = applesilicon.NewAPI(m.ScwClient())
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := meta.ExtractFrameworkZone(data.Zone, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve zone", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	createReq := &applesilicon.CreateServerRequest{
		Name:           scwtypes.ExpandOrGenerateString(data.Name.ValueString(), "m1"),
		Type:           data.Type.ValueString(),
		ProjectID:      projectID,
		EnableVpc:      data.EnableVpc.ValueBool(),
		CommitmentType: applesilicon.CommitmentType(data.Commitment.ValueString()),
		Zone:           zone,
	}

	if !data.OsID.IsNull() && !data.OsID.IsUnknown() {
		createReq.OsID = scwtypes.ExpandRawID(data.OsID, "os_id", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	runnerIDs := expandRunnerIDs(ctx, data.RunnerIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(runnerIDs) > 0 {
		createReq.AppliedRunnerConfigurations = &applesilicon.AppliedRunnerConfigurations{
			RunnerConfigurationIDs: runnerIDs,
		}
	}

	if !data.PublicBandwidth.IsNull() && !data.PublicBandwidth.IsUnknown() {
		createReq.PublicBandwidthBps = uint64(data.PublicBandwidth.ValueInt64())
	}

	res, err := r.api.CreateServer(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create apple silicon server", err.Error())

		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetZonalIdentity(res.Zone, res.ID))...)

	_, err = waitForAppleSiliconServer(ctx, r.api, zone, res.ID, defaultAppleSiliconServerTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for apple silicon server", err.Error())

		return
	}

	if !data.PrivateNetwork.IsNull() && !data.PrivateNetwork.IsUnknown() {
		pnMap := expandPrivateNetworks(ctx, data.PrivateNetwork, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		privateNetworkAPI := applesilicon.NewPrivateNetworkAPI(r.meta.ScwClient())
		_, err = privateNetworkAPI.SetServerPrivateNetworks(&applesilicon.PrivateNetworkAPISetServerPrivateNetworksRequest{
			Zone:                       zone,
			ServerID:                   res.ID,
			PerPrivateNetworkIpamIPIDs: pnMap,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to set private networks", err.Error())

			return
		}

		_, err = waitForAppleSiliconPrivateNetworkServer(ctx, privateNetworkAPI, zone, res.ID, defaultAppleSiliconServerTimeout)
		if err != nil {
			resp.Diagnostics.AddError("Failed to wait for private network server", err.Error())

			return
		}
	}

	res, err = r.api.GetServer(&applesilicon.GetServerRequest{
		Zone:     zone,
		ServerID: res.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read apple silicon server", err.Error())

		return
	}

	newState, err := flattenServer(ctx, r.meta, res, zone, req, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to flatten server", err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    serverResourceModel
		identity serverResourceIdentityModel
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

	var resourceID string
	if identityAvailable {
		resourceID = identity.ID.ValueString()
	} else {
		resourceID = state.ID.ValueString()
	}

	zone, id, err := zonal.ParseID(resourceID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse server ID", err.Error())

		return
	}

	res, err := r.api.GetServer(&applesilicon.GetServerRequest{
		Zone:     zone,
		ServerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) || httperrors.Is403(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read apple silicon server", err.Error())

		return
	}

	newState, err := flattenServer(ctx, r.meta, res, zone, req, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to flatten server", err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetZonalIdentity(res.Zone, res.ID))...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverResourceModel
	var state serverResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse server ID", err.Error())

		return
	}

	updateReq := &applesilicon.UpdateServerRequest{
		Zone:     zone,
		ServerID: id,
	}

	if !plan.Name.Equal(state.Name) {
		updateReq.Name = scwtypes.ExpandRawID(plan.Name, "name", &resp.Diagnostics)
	}

	if !plan.Commitment.Equal(state.Commitment) {
		updateReq.CommitmentType = &applesilicon.CommitmentTypeValue{
			CommitmentType: applesilicon.CommitmentType(plan.Commitment.ValueString()),
		}
	}

	if !plan.EnableVpc.Equal(state.EnableVpc) {
		updateReq.EnableVpc = new(bool)
		*updateReq.EnableVpc = plan.EnableVpc.ValueBool()
	}

	if !plan.PublicBandwidth.Equal(state.PublicBandwidth) {
		bw := uint64(plan.PublicBandwidth.ValueInt64())
		updateReq.PublicBandwidthBps = &bw
	}

	if !plan.RunnerIDs.Equal(state.RunnerIDs) {
		runnerIDs := expandRunnerIDs(ctx, plan.RunnerIDs, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.AppliedRunnerConfigurations = &applesilicon.AppliedRunnerConfigurations{
			RunnerConfigurationIDs: runnerIDs,
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = r.api.UpdateServer(updateReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update apple silicon server", err.Error())

		return
	}

	err = waitForTerminalVPCState(ctx, r.api, zone, id, defaultAppleSiliconServerTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for VPC terminal state", err.Error())

		return
	}

	privateNetworkAPI := applesilicon.NewPrivateNetworkAPI(r.meta.ScwClient())

	if !plan.PrivateNetwork.Equal(state.PrivateNetwork) && plan.EnableVpc.ValueBool() {
		pnMap := expandPrivateNetworks(ctx, plan.PrivateNetwork, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err = privateNetworkAPI.SetServerPrivateNetworks(&applesilicon.PrivateNetworkAPISetServerPrivateNetworksRequest{
			Zone:                       zone,
			ServerID:                   id,
			PerPrivateNetworkIpamIPIDs: pnMap,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to set private networks", err.Error())

			return
		}
	}

	_, err = waitForAppleSiliconPrivateNetworkServer(ctx, privateNetworkAPI, zone, id, defaultAppleSiliconServerTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for private network server", err.Error())

		return
	}

	res, err := r.api.GetServer(&applesilicon.GetServerRequest{
		Zone:     zone,
		ServerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read apple silicon server after update", err.Error())

		return
	}

	newState, err := flattenServer(ctx, r.meta, res, zone, req, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError("Failed to flatten server", err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetZonalIdentity(res.Zone, res.ID))...)
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, id, err := zonal.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse server ID", err.Error())

		return
	}

	err = detachAllPrivateNetworksFromServer(ctx, r.meta, zone, id)
	if err != nil {
		resp.Diagnostics.AddError("Failed to detach private networks", err.Error())

		return
	}

	_, err = waitForAppleSiliconServer(ctx, r.api, zone, id, defaultAppleSiliconServerTimeout)
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to wait for apple silicon server before deletion", err.Error())

		return
	}

	err = r.api.DeleteServer(&applesilicon.DeleteServerRequest{
		Zone:     zone,
		ServerID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is403(err) {
		resp.Diagnostics.AddError("Failed to delete apple silicon server", err.Error())

		return
	}
}

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func getOsIDFromReference(ctx context.Context, reference any, diags *diag.Diagnostics) types.String {
	var osID basetypes.StringValue

	switch req := reference.(type) {
	case resource.CreateRequest:
		d := req.Config.GetAttribute(ctx, path.Root("os_id"), &osID)
		diags.Append(d...)
	case resource.ReadRequest:
		d := req.State.GetAttribute(ctx, path.Root("os_id"), &osID)
		diags.Append(d...)
	case resource.UpdateRequest:
		d := req.Config.GetAttribute(ctx, path.Root("os_id"), &osID)
		diags.Append(d...)
	}

	return osID
}

func getBoolFromReference(ctx context.Context, reference any, attrName string, diags *diag.Diagnostics) types.Bool {
	var val basetypes.BoolValue

	switch req := reference.(type) {
	case resource.CreateRequest:
		d := req.Config.GetAttribute(ctx, path.Root(attrName), &val)
		diags.Append(d...)
	case resource.ReadRequest:
		d := req.State.GetAttribute(ctx, path.Root(attrName), &val)
		diags.Append(d...)
	case resource.UpdateRequest:
		d := req.Config.GetAttribute(ctx, path.Root(attrName), &val)
		diags.Append(d...)
	}

	return val
}

func flattenServer(ctx context.Context, m *meta.Meta, res *applesilicon.Server, zone scw.Zone, reference any, diags *diag.Diagnostics) (serverResourceModel, error) {
	osIDFromReference := getOsIDFromReference(ctx, reference, diags)

	model := serverResourceModel{
		ID:             types.StringValue(zonal.NewIDString(zone, res.ID)),
		Name:           types.StringValue(res.Name),
		Type:           types.StringValue(res.Type),
		EnableVpc:      getBoolFromReference(ctx, reference, "enable_vpc", diags),
		State:          types.StringValue(res.Status.String()),
		IP:             types.StringValue(res.IP.String()),
		VncURL:         types.StringValue(res.VncURL),
		VpcStatus:      types.StringValue(res.VpcStatus.String()),
		Zone:           types.StringValue(res.Zone.String()),
		OrganizationID: types.StringValue(res.OrganizationID),
		ProjectID:      types.StringValue(res.ProjectID),
		Password:       types.StringValue(res.SudoPassword),
		Username:       types.StringValue(res.SSHUsername),
	}

	if !osIDFromReference.IsNull() && !osIDFromReference.IsUnknown() {
		if scwtypes.IDUsesZonedFormat(ctx, reference, path.Root("os_id"), diags) {
			if res.Os != nil {
				model.OsID = types.StringValue(zonal.NewIDString(zone, res.Os.ID))
			}
		} else {
			if res.Os != nil {
				model.OsID = types.StringValue(res.Os.ID)
			}
		}
	}

	model.PublicBandwidth = types.Int64Value(int64(res.PublicBandwidthBps))

	if res.CreatedAt != nil {
		model.CreatedAt = types.StringValue(res.CreatedAt.Format(time.RFC3339))
	} else {
		model.CreatedAt = types.StringNull()
	}

	if res.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(res.UpdatedAt.Format(time.RFC3339))
	} else {
		model.UpdatedAt = types.StringNull()
	}

	if res.DeletableAt != nil {
		model.DeletableAt = types.StringValue(res.DeletableAt.Format(time.RFC3339))
	} else {
		model.DeletableAt = types.StringNull()
	}

	switch res.VpcStatus {
	case applesilicon.ServerPrivateNetworkStatusVpcDisabled:
		model.EnableVpc = types.BoolValue(false)
	case applesilicon.ServerPrivateNetworkStatusVpcEnabled:
		model.EnableVpc = types.BoolValue(true)
	}

	model.Commitment = types.StringValue(applesilicon.CommitmentTypeDuration24h.String())
	if res.Commitment != nil {
		switch res.Commitment.Type {
		case applesilicon.CommitmentTypeNone, applesilicon.CommitmentTypeDuration24h:
			model.Commitment = types.StringValue(applesilicon.CommitmentTypeDuration24h.String())
		case applesilicon.CommitmentTypeRenewedMonthly:
			model.Commitment = types.StringValue(applesilicon.CommitmentTypeRenewedMonthly.String())
		}
	}

	runnerIDsList, d := scwtypes.FlattenStringList(ctx, "runner_ids", res.AppliedRunnerConfigurationIDs, reference)
	diags.Append(d...)
	model.RunnerIDs = runnerIDsList

	privateNetworkAPI := applesilicon.NewPrivateNetworkAPI(m.ScwClient())
	listPrivateNetworks, err := privateNetworkAPI.ListServerPrivateNetworks(&applesilicon.PrivateNetworkAPIListServerPrivateNetworksRequest{
		Zone:     res.Zone,
		ServerID: &res.ID,
	})
	if err != nil {
		return model, fmt.Errorf("failed to list server private networks: %w", err)
	}

	pnRegion, err := res.Zone.Region()
	if err != nil {
		return model, fmt.Errorf("failed to get region from zone: %w", err)
	}

	pnSet, d := flattenPrivateNetworks(ctx, pnRegion, listPrivateNetworks.ServerPrivateNetworks)
	diags.Append(d...)
	model.PrivateNetwork = pnSet

	privateNetworkIDs := make([]string, 0, listPrivateNetworks.TotalCount)
	for _, pn := range listPrivateNetworks.ServerPrivateNetworks {
		privateNetworkIDs = append(privateNetworkIDs, pn.PrivateNetworkID)
	}

	allPrivateIPs := make([]privateIPsBlockModel, 0, listPrivateNetworks.TotalCount)
	authorized := true

	for _, privateNetworkID := range privateNetworkIDs {
		opts := &ipam.GetResourcePrivateIPsOptions{
			ResourceType:     new(ipamAPI.ResourceTypeAppleSiliconPrivateNic),
			PrivateNetworkID: &privateNetworkID,
			ProjectID:        &res.ProjectID,
		}

		privateIPs, err := ipam.GetResourcePrivateIPs(ctx, m, pnRegion, opts)

		switch {
		case err == nil:
			for _, ip := range privateIPs {
				allPrivateIPs = append(allPrivateIPs, privateIPsBlockModel{
					ID:      types.StringValue(ip["id"].(string)),
					Address: types.StringValue(ip["address"].(string)),
				})
			}
		case httperrors.Is403(err):
			authorized = false

			diags.Append(diag.NewWarningDiagnostic(
				"Unauthorized to read server's private IPs, please check your IAM permissions",
				err.Error(),
			))
		default:
			diags.Append(diag.NewWarningDiagnostic(
				fmt.Sprintf("Unable to get private IP for server %q", res.Name),
				err.Error(),
			))
		}

		if !authorized {
			break
		}
	}

	if authorized {
		privateIPsList, d := types.ListValueFrom(ctx, privateIPsObjectType(), allPrivateIPs)
		diags.Append(d...)
		model.PrivateIPs = privateIPsList
	} else {
		model.PrivateIPs = types.ListNull(privateIPsObjectType())
	}

	return model, nil
}

func expandPrivateNetworks(ctx context.Context, pnSet types.Set, diags *diag.Diagnostics) map[string]*[]string {
	privateNetworks := make(map[string]*[]string)

	var pnModels []privateNetworkBlockModel
	diags.Append(pnSet.ElementsAs(ctx, &pnModels, false)...)
	if diags.HasError() {
		return nil
	}

	for _, pn := range pnModels {
		id := locality.ExpandID(pn.ID.ValueString())

		ipamIPIDs := &[]string{}
		if !pn.IpamIPIDs.IsNull() && !pn.IpamIPIDs.IsUnknown() {
			var ipamIDs []string
			diags.Append(pn.IpamIPIDs.ElementsAs(ctx, &ipamIDs, false)...)
			if diags.HasError() {
				return nil
			}

			rawIDs := make([]string, 0, len(ipamIDs))
			for _, ipID := range ipamIDs {
				rawIDs = append(rawIDs, locality.ExpandID(ipID))
			}

			ipamIPIDs = &rawIDs
		}

		privateNetworks[id] = ipamIPIDs
	}

	return privateNetworks
}

func flattenPrivateNetworks(ctx context.Context, region scw.Region, privateNetworks []*applesilicon.ServerPrivateNetwork) (types.Set, diag.Diagnostics) {
	if len(privateNetworks) == 0 {
		return types.SetNull(privateNetworkObjectType()), nil
	}

	models := make([]privateNetworkBlockModel, 0, len(privateNetworks))
	for _, pn := range privateNetworks {
		model := privateNetworkBlockModel{
			ID:     types.StringValue(regional.NewIDString(region, pn.PrivateNetworkID)),
			Status: types.StringValue(string(pn.Status)),
		}

		if pn.Vlan != nil {
			model.Vlan = types.Int64Value(int64(*pn.Vlan))
		} else {
			model.Vlan = types.Int64Null()
		}

		if pn.CreatedAt != nil {
			model.CreatedAt = types.StringValue(pn.CreatedAt.Format(time.RFC3339))
		} else {
			model.CreatedAt = types.StringNull()
		}

		if pn.UpdatedAt != nil {
			model.UpdatedAt = types.StringValue(pn.UpdatedAt.Format(time.RFC3339))
		} else {
			model.UpdatedAt = types.StringNull()
		}

		if len(pn.IpamIPIDs) > 0 {
			ipamIDs := make([]string, 0, len(pn.IpamIPIDs))
			for _, id := range pn.IpamIPIDs {
				ipamIDs = append(ipamIDs, regional.NewIDString(region, id))
			}

			listVal, d := types.ListValueFrom(ctx, types.StringType, ipamIDs)
			if d.HasError() {
				return types.SetNull(privateNetworkObjectType()), d
			}

			model.IpamIPIDs = listVal
		} else {
			model.IpamIPIDs = types.ListNull(types.StringType)
		}

		models = append(models, model)
	}

	return types.SetValueFrom(ctx, privateNetworkObjectType(), models)
}

func expandRunnerIDs(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	rawIDs := scwtypes.ExpandStringList(ctx, list, diags)
	if diags.HasError() {
		return nil
	}

	result := make([]string, 0, len(rawIDs))
	for _, id := range rawIDs {
		result = append(result, locality.ExpandID(id))
	}

	return result
}

func detachAllPrivateNetworksFromServer(ctx context.Context, m *meta.Meta, zone scw.Zone, serverID string) error {
	privateNetworkAPI := applesilicon.NewPrivateNetworkAPI(meta.ExtractScwClient(m))

	listPrivateNetwork, err := privateNetworkAPI.ListServerPrivateNetworks(&applesilicon.PrivateNetworkAPIListServerPrivateNetworksRequest{
		Zone:     zone,
		ServerID: &serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	for _, pn := range listPrivateNetwork.ServerPrivateNetworks {
		err := privateNetworkAPI.DeleteServerPrivateNetwork(&applesilicon.PrivateNetworkAPIDeleteServerPrivateNetworkRequest{
			Zone:             zone,
			ServerID:         serverID,
			PrivateNetworkID: pn.PrivateNetworkID,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}
	}

	_, err = waitForAppleSiliconPrivateNetworkServer(ctx, privateNetworkAPI, zone, serverID, defaultAppleSiliconServerTimeout)
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	return nil
}
