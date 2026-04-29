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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*PrivateNetworkListResource)(nil)
	_ list.ListResourceWithConfigure    = (*PrivateNetworkListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*PrivateNetworkListResource)(nil)
)

type PrivateNetworkListResource struct {
	meta   *meta.Meta
	vpcAPI *vpc.API
}

func (r *PrivateNetworkListResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func NewPrivateNetworkListResource() list.ListResource {
	return &PrivateNetworkListResource{}
}

func (r *PrivateNetworkListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"vpc_id": schema.StringAttribute{
				Description: "VPC ID to filter for",
				Optional:    true,
			},
			"name":            listscw.NameAttribute("Name of the private network to filter for"),
			"tags":            listscw.TagsAttribute("Tags of the private network to filter for"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter for."),
			"regions":         listscw.RegionsAttribute("Regions to filter for."),
		},
	}
}

func (r *PrivateNetworkListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	pnResource := ResourcePrivateNetwork()

	resp.ProtoV6Schema = translate.Schema(pnResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(pnResource.ProtoIdentitySchema(ctx)())
}

type PrivateNetworkListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Regions        types.List   `tfsdk:"regions"`
	VpcID          types.String `tfsdk:"vpc_id"`
}

func (m *PrivateNetworkListResourceModel) GetTags() types.List    { return m.Tags }
func (m *PrivateNetworkListResourceModel) GetRegions() types.List { return m.Regions }
func (m *PrivateNetworkListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (r *PrivateNetworkListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_private_network"
}

func (r *PrivateNetworkListResource) FetchPrivateNetworks(ctx context.Context, region scw.Region, project *string, tags []string, data PrivateNetworkListResourceModel) ([]*vpc.PrivateNetwork, error) {
	var vpcID *string
	if !data.VpcID.IsNull() {
		vpcID = new(locality.ExpandID(data.VpcID.ValueString()))
	}

	listRequest := &vpc.ListPrivateNetworksRequest{
		Region:         region,
		Name:           data.Name.ValueStringPointer(),
		Tags:           tags,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      project,
		VpcID:          vpcID,
	}

	response, err := r.vpcAPI.ListPrivateNetworks(listRequest, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.PrivateNetworks, nil
}

func (r *PrivateNetworkListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data PrivateNetworkListResourceModel

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

	allPNs, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*vpc.PrivateNetwork, error) {
			return r.FetchPrivateNetworks(ctx, target.Region, &target.ProjectID, tags, data)
		},
		func(a, b *vpc.PrivateNetwork) int {
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
			diag.NewErrorDiagnostic("Listing Private Networks", "Failed to list Private Networks: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, pn := range allPNs {
			result := req.NewListResult(ctx)
			result.DisplayName = pn.Name

			pnResource := ResourcePrivateNetwork()
			resourceData := pnResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, pn.Region, pn.ID)
			if err != nil {
				result.Diagnostics.AddError("Retrieving identity data",
					"An error was encountered when retrieving the identity data: "+err.Error(),
				)

				if !push(result) {
					return
				}

				continue
			}

			tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
			if errIdentityState != nil {
				result.Diagnostics.AddError(
					"Converting identity data",
					"An error was encountered when converting the identity data: "+errIdentityState.Error(),
				)
			}

			identitySetDiags := result.Identity.Set(ctx, *tfTypeIdentity)
			result.Diagnostics.Append(identitySetDiags...)

			sdkDiags := setPrivateNetworkState(resourceData, r.meta, pn)
			if sdkDiags.HasError() {
				tflog.Error(ctx, "error from setting private network state")

				for _, d := range sdkDiags {
					result.Diagnostics.AddError(d.Summary, d.Detail)
				}

				if !push(result) {
					return
				}

				continue
			}

			tfTypeResource, errTfTypeResourceState := resourceData.TfTypeResourceState()
			if errTfTypeResourceState != nil {
				result.Diagnostics.AddError(
					"Converting resource state",
					"An error was encountered when converting the resource state: "+errTfTypeResourceState.Error(),
				)
			}

			resourceSetDiags := result.Resource.Set(ctx, *tfTypeResource)
			result.Diagnostics.Append(resourceSetDiags...)

			if !push(result) {
				return
			}
		}
	}
}
