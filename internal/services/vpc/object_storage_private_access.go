package vpc

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*ObjectStoragePrivateAccessResource)(nil)
	_ resource.ResourceWithConfigure   = (*ObjectStoragePrivateAccessResource)(nil)
	_ resource.ResourceWithImportState = (*ObjectStoragePrivateAccessResource)(nil)
	_ resource.ResourceWithIdentity    = (*ObjectStoragePrivateAccessResource)(nil)
)

func NewObjectStoragePrivateAccessResource() resource.Resource {
	return &ObjectStoragePrivateAccessResource{}
}

type ObjectStoragePrivateAccessResource struct {
	vpcAPI *vpc.API
	meta   *meta.Meta
}

type objectStoragePrivateAccessResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Region            types.String `tfsdk:"region"`
	VpcID             types.String `tfsdk:"vpc_id"`
	PrivateNetworkIDs types.Set    `tfsdk:"private_network_ids"`
}

//go:embed descriptions/object_storage_private_access_resource.md
var objectStoragePrivateAccessResourceDescription string

func (r *ObjectStoragePrivateAccessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_object_storage_private_access"
}

func (r *ObjectStoragePrivateAccessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: objectStoragePrivateAccessResourceDescription,
		Description:         objectStoragePrivateAccessResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Object Storage private access, in the `{region}/{vpc_id}` format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region of the VPC. Defaults to the region configured on the provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vpc_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the VPC for which Object Storage private access is enabled.",
				PlanModifiers: []planmodifier.String{
					regional.LocalityPlanModifier(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"private_network_ids": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "IDs of the Private Networks that can reach Object Storage over the private network.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					verify.SetElemIsStringUUIDOrUUIDWithRegion(),
				},
				PlanModifiers: []planmodifier.Set{
					regional.LocalitySetPlanModifier(),
				},
			},
		},
	}
}

func (r *ObjectStoragePrivateAccessResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultRegional()
}

func (r *ObjectStoragePrivateAccessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.vpcAPI = vpc.NewAPI(m.ScwClient())
}

func (r *ObjectStoragePrivateAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan objectStoragePrivateAccessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, vpcID, err := regional.ResolveID(plan.VpcID.ValueString(), plan.Region.ValueString(), r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("vpc_id"), "Error resolving VPC ID", err.Error())

		return
	}

	if !plan.Region.IsUnknown() && !plan.Region.IsNull() && plan.Region.ValueString() != region.String() {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Region mismatch",
			fmt.Sprintf("region %q does not match the region of vpc_id (%s)", plan.Region.ValueString(), region),
		)

		return
	}

	privateNetworkIDs, diags := locality.ExpandFrameworkIDs(ctx, plan.PrivateNetworkIDs)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = r.vpcAPI.EnableObjectStoragePrivateAccess(&vpc.EnableObjectStoragePrivateAccessRequest{
		Region:            region,
		VpcID:             vpcID,
		PrivateNetworkIDs: privateNetworkIDs,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Object Storage private access",
			fmt.Sprintf("enabling Object Storage private access for VPC (%s): %s", vpcID, err),
		)

		return
	}

	plan.ID = types.StringValue(regional.NewIDString(region, vpcID))
	plan.Region = types.StringValue(region.String())

	state, readDiags := r.readObjectStoragePrivateAccess(ctx, region, vpcID, plan)
	resp.Diagnostics.Append(readDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state == nil {
		resp.Diagnostics.AddError(
			"Error creating Object Storage private access",
			fmt.Sprintf("Object Storage private access for VPC (%s) was not found after enable", vpcID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(region, vpcID))...)
}

func (r *ObjectStoragePrivateAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state objectStoragePrivateAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, vpcID, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Object Storage private access",
			fmt.Sprintf("parsing Object Storage private access ID (%s): %s", state.ID.ValueString(), err),
		)

		return
	}

	newState, diags := r.readObjectStoragePrivateAccess(ctx, region, vpcID, state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if newState == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(region, vpcID))...)
}

func (r *ObjectStoragePrivateAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state objectStoragePrivateAccessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, vpcID, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Object Storage private access",
			fmt.Sprintf("parsing Object Storage private access ID (%s): %s", state.ID.ValueString(), err),
		)

		return
	}

	if !plan.PrivateNetworkIDs.Equal(state.PrivateNetworkIDs) {
		privateNetworkIDs, diags := locality.ExpandFrameworkIDs(ctx, plan.PrivateNetworkIDs)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		_, err = r.vpcAPI.SetPrivateNetworksObjectStoragePrivateAccess(&vpc.SetPrivateNetworksObjectStoragePrivateAccessRequest{
			Region:            region,
			VpcID:             vpcID,
			PrivateNetworkIDs: privateNetworkIDs,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating Object Storage private access",
				fmt.Sprintf("setting Private Networks for Object Storage private access (%s): %s", vpcID, err),
			)

			return
		}
	}

	newState, readDiags := r.readObjectStoragePrivateAccess(ctx, region, vpcID, plan)
	resp.Diagnostics.Append(readDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if newState == nil {
		resp.Diagnostics.AddError(
			"Error updating Object Storage private access",
			fmt.Sprintf("Object Storage private access for VPC (%s) was not found after update", vpcID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(region, vpcID))...)
}

func (r *ObjectStoragePrivateAccessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state objectStoragePrivateAccessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, vpcID, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Object Storage private access",
			fmt.Sprintf("parsing Object Storage private access ID (%s): %s", state.ID.ValueString(), err),
		)

		return
	}

	listResp, err := r.vpcAPI.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
		Region:                            region,
		VpcID:                             &vpcID,
		ObjectStoragePrivateAccessEnabled: new(true),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError(
			"Error deleting Object Storage private access",
			fmt.Sprintf("listing Private Networks for VPC (%s): %s", vpcID, err),
		)

		return
	}

	for _, pn := range listResp.PrivateNetworks {
		err = r.vpcAPI.DeletePrivateNetworkObjectStoragePrivateAccess(&vpc.DeletePrivateNetworkObjectStoragePrivateAccessRequest{
			Region:           region,
			VpcID:            vpcID,
			PrivateNetworkID: pn.ID,
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting Object Storage private access",
				fmt.Sprintf("removing Private Network (%s) from Object Storage private access: %s", pn.ID, err),
			)

			return
		}
	}

	_, err = r.vpcAPI.DisableObjectStoragePrivateAccess(&vpc.DisableObjectStoragePrivateAccessRequest{
		Region: region,
		VpcID:  vpcID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError(
			"Error deleting Object Storage private access",
			fmt.Sprintf("disabling Object Storage private access for VPC (%s): %s", vpcID, err),
		)
	}
}

func (r *ObjectStoragePrivateAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, vpcID, err := regional.ParseID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Object Storage private access import ID must be {region}/{vpc_id}, got %q: %s", req.ID, err),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), regional.NewIDString(region, vpcID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region.String())...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vpc_id"), regional.NewIDString(region, vpcID))...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetRegionalIdentity(region, vpcID))...)
}

// readObjectStoragePrivateAccess loads the VPC flag and the Private Networks that have Object Storage private access
// A nil model means the private access is gone and should be removed from state
func (r *ObjectStoragePrivateAccessResource) readObjectStoragePrivateAccess(ctx context.Context, region scw.Region, vpcID string, current objectStoragePrivateAccessResourceModel) (*objectStoragePrivateAccessResourceModel, diag.Diagnostics) {
	res, err := r.vpcAPI.GetVPC(&vpc.GetVPCRequest{
		Region: region,
		VpcID:  vpcID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return nil, nil
		}

		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Error reading Object Storage private access",
			fmt.Sprintf("reading VPC (%s): %s", vpcID, err),
		)}
	}

	if !res.ObjectStoragePrivateAccessEnabled {
		return nil, nil
	}

	listResp, err := r.vpcAPI.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
		Region:                            region,
		VpcID:                             &vpcID,
		ObjectStoragePrivateAccessEnabled: new(true),
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Error reading Object Storage private access",
			fmt.Sprintf("listing Private Networks for VPC (%s): %s", vpcID, err),
		)}
	}

	knownIDs := map[string]string{}

	if !current.PrivateNetworkIDs.IsNull() && !current.PrivateNetworkIDs.IsUnknown() {
		var currentIDs []string

		diags := current.PrivateNetworkIDs.ElementsAs(ctx, &currentIDs, false)
		if diags.HasError() {
			return nil, diags
		}

		for _, id := range currentIDs {
			knownIDs[locality.ExpandID(id)] = id
		}
	}

	ids := make([]string, 0, len(listResp.PrivateNetworks))
	for _, pn := range listResp.PrivateNetworks {
		if pn == nil {
			continue
		}

		if id, ok := knownIDs[pn.ID]; ok {
			ids = append(ids, id)
		} else {
			ids = append(ids, regional.NewIDString(region, pn.ID))
		}
	}

	privateNetworkIDs, diags := types.SetValueFrom(ctx, types.StringType, ids)
	if diags.HasError() {
		return nil, diags
	}

	current.ID = types.StringValue(regional.NewIDString(region, vpcID))
	current.Region = types.StringValue(region.String())
	current.PrivateNetworkIDs = privateNetworkIDs

	if current.VpcID.IsNull() || current.VpcID.IsUnknown() || locality.ExpandID(current.VpcID.ValueString()) != vpcID {
		current.VpcID = types.StringValue(regional.NewIDString(region, vpcID))
	}

	return &current, nil
}
