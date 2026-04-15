package vpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*VPCListResource)(nil)
	_ list.ListResourceWithConfigure    = (*VPCListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*VPCListResource)(nil)
)

type VPCListResource struct {
	meta   *meta.Meta
	vpcAPI *vpc.API
}

func (r *VPCListResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	m, ok := request.ProviderData.(*meta.Meta)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected List Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.meta = m
	r.vpcAPI = vpc.NewAPI(meta.ExtractScwClient(m))
}

func NewVPCListResource() list.ListResource {
	return &VPCListResource{}
}

func (r *VPCListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"routing_enabled": schema.BoolAttribute{
				Description: "Whether routing is enabled for VPC",
				Optional:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether the VPC is the default VPC",
				Optional:    true,
			},
			"name":            listscw.NameAttribute("Name of the vpc to list for"),
			"tags":            listscw.TagsAttribute("Tags of the VPC to list for"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID of the VPC to list for"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs of the VPC to list for"),
			"regions":         listscw.RegionsAttribute("Regions of the VPC to list for"),
		},
	}
}

func (r *VPCListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	resourceVPC := ResourceVPC()

	resp.ProtoV6Schema = translate.Schema(resourceVPC.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(resourceVPC.ProtoIdentitySchema(ctx)())
}

type ListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Regions        types.List   `tfsdk:"regions"`
	RoutingEnabled types.Bool   `tfsdk:"routing_enabled"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
}

func (m *ListResourceModel) GetTags() types.List {
	return m.Tags
}

func (m *ListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *ListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (r *VPCListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *VPCListResource) FetchVPCs(ctx context.Context, region scw.Region, project *string, tags []string, data ListResourceModel) ([]*vpc.VPC, error) {
	listRequest := &vpc.ListVPCsRequest{
		Region:         region,
		Name:           data.Name.ValueStringPointer(),
		Tags:           tags,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      project,
		IsDefault:      data.IsDefault.ValueBoolPointer(),
		RoutingEnabled: data.RoutingEnabled.ValueBoolPointer(),
	}

	response, err := r.vpcAPI.ListVPCs(listRequest, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Vpcs, nil
}

func (r *VPCListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ListResourceModel

	// Read list config data into the model
	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	tags, diags := listscw.ExtractTags(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	regions, err := listscw.ExtractRegions(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing regions", "An error was encountered when listing regions: "+err.Error()),
		})

		return
	}

	projects, err := listscw.ExtractProjects(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing projects", "An error was encountered when listing projects: "+err.Error()),
		})

		return
	}

	var targets []listscw.RegionalFetchTarget

	for _, r := range regions {
		for _, p := range projects {
			targets = append(targets, listscw.RegionalFetchTarget{Region: r, ProjectID: p})
		}
	}

	allVPCs, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*vpc.VPC, error) {
			return r.FetchVPCs(ctx, target.Region, &target.ProjectID, tags, data)
		},
		func(a, b *vpc.VPC) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Region != b.Region {
				return strings.Compare(string(a.Region), string(b.Region))
			}

			return strings.Compare(a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing VPCs", "Failed to list VPCs: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, rawVPC := range allVPCs {
			result := req.NewListResult(ctx)
			result.DisplayName = rawVPC.Name

			vpcResource := ResourceVPC()
			resourceData := vpcResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, rawVPC.Region, rawVPC.ID)
			if err != nil {
				result.Diagnostics.AddError(
					"Retrieving identity data",
					"An error was encountered when retrieving the identity data: "+err.Error(),
				)

				if !push(result) {
					return
				}

				continue
			}

			// Convert and set the identity and resource state into the result
			tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
			if errIdentityState != nil {
				result.Diagnostics.AddError(
					"Converting identity data",
					"An error was encountered when converting the identity data: "+errIdentityState.Error(),
				)
			}

			identitySetDiags := result.Identity.Set(ctx, *tfTypeIdentity)
			result.Diagnostics.Append(identitySetDiags...)

			diagsState := setVPCState(resourceData, rawVPC)
			if diagsState.HasError() {
				tflog.Error(ctx, "error from setting setVPCState")

				if !push(result) {
					return
				}

				continue
			}

			// Convert and set the resource state into the result
			tfTypeResource, errTfTypeResourceState := resourceData.TfTypeResourceState()
			if errTfTypeResourceState != nil {
				result.Diagnostics.AddError(
					"Converting resource state",
					"An error was encountered when converting the resource state: "+errTfTypeResourceState.Error(),
				)
			}

			resourceSetDiags := result.Resource.Set(ctx, *tfTypeResource)
			result.Diagnostics.Append(resourceSetDiags...)

			// Send the result to the stream.
			if !push(result) {
				return
			}
		}
	}
}
