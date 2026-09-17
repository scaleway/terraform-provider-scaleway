package vpcgw

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*PublicGatewayListResource)(nil)
	_ list.ListResourceWithConfigure    = (*PublicGatewayListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*PublicGatewayListResource)(nil)
)

type PublicGatewayListResource struct {
	meta        *meta.Meta
	publicGWAPI *vpcgw.API
}

func (r *PublicGatewayListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.publicGWAPI = vpcgw.NewAPI(meta.ExtractScwClient(m))
}

func NewPublicGatewayListResource() list.ListResource {
	return &PublicGatewayListResource{}
}

func (r *PublicGatewayListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":            listscw.NameAttribute("Name of the Public Gateway to filter for"),
			"tags":            listscw.TagsAttribute("Tags of the Public Gateway to filter for"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter for."),
			"zones":           listscw.ZonesAttribute("Zones to filter for."),
			"types": schema.ListAttribute{
				Description: "Filter for gateways of these types (e.g. VPC-GW-S, VPC-GW-M)",
				Optional:    true,
				ElementType: types.StringType,
			},
			"private_network_ids": schema.ListAttribute{
				Description: "Filter for gateways attached to these Private Networks (accepts zonal IDs like fr-par-1/uuid or raw UUIDs)",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *PublicGatewayListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	gwResource := ResourcePublicGateway()

	resp.ProtoV6Schema = translate.Schema(gwResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(gwResource.ProtoIdentitySchema(ctx)())
}

type PublicGatewayListResourceModel struct {
	Tags              types.List   `tfsdk:"tags"`
	Types             types.List   `tfsdk:"types"`
	PrivateNetworkIDs types.List   `tfsdk:"private_network_ids"`
	Zones             types.List   `tfsdk:"zones"`
	ProjectIDs        types.List   `tfsdk:"project_ids"`
	Name              types.String `tfsdk:"name"`
	OrganizationID    types.String `tfsdk:"organization_id"`
}

func (m *PublicGatewayListResourceModel) GetTags() types.List     { return m.Tags }
func (m *PublicGatewayListResourceModel) GetZones() types.List    { return m.Zones }
func (m *PublicGatewayListResourceModel) GetProjects() types.List { return m.ProjectIDs }

func (r *PublicGatewayListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_public_gateway"
}

func (r *PublicGatewayListResource) FetchPublicGateways(ctx context.Context, zone scw.Zone, project *string, tags, gwTypes, pnIDs []string, data PublicGatewayListResourceModel) ([]*vpcgw.Gateway, error) {
	listRequest := &vpcgw.ListGatewaysRequest{
		Zone:              zone,
		Name:              data.Name.ValueStringPointer(),
		Tags:              tags,
		OrganizationID:    data.OrganizationID.ValueStringPointer(),
		ProjectID:         project,
		Types:             gwTypes,
		PrivateNetworkIDs: pnIDs,
	}

	response, err := r.publicGWAPI.ListGateways(listRequest, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.Gateways, nil
}

func (r *PublicGatewayListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data PublicGatewayListResourceModel

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

	var gwTypes []string

	if !data.Types.IsNull() && !data.Types.IsUnknown() {
		diags = data.Types.ElementsAs(ctx, &gwTypes, false)
		if diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)

			return
		}
	}

	pnIDs, diags := locality.ExpandFrameworkIDs(ctx, data.PrivateNetworkIDs)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	zones, err := listscw.ExtractZones(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing zones", "An error was encountered when listing zones: "+err.Error()),
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

	allGateways, err := listscw.FetchConcurrently(ctx, listscw.ZonalProjectTargets(zones, projects),
		func(ctx context.Context, target listscw.ZonalFetchTarget) ([]*vpcgw.Gateway, error) {
			return r.FetchPublicGateways(ctx, target.Zone, &target.ProjectID, tags, gwTypes, pnIDs, data)
		},
		func(a, b *vpcgw.Gateway) int {
			return listscw.CompareZonalProjectItems(a.ProjectID, b.ProjectID, a.Zone, b.Zone, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Public Gateways", "Failed to list Public Gateways: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, gw := range allGateways {
			result := req.NewListResult(ctx)
			result.DisplayName = gw.Name

			gwResource := ResourcePublicGateway()
			resourceData := gwResource.Data(&terraform.InstanceState{})

			err := identity.SetZonalIdentity(resourceData, gw.Zone, gw.ID)
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

			sdkDiags := setPublicGatewayState(resourceData, gw)
			if sdkDiags.HasError() {
				tflog.Error(ctx, "error from setting public gateway state")

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
