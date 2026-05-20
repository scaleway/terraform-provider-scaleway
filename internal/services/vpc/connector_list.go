package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ list.ListResource                 = (*ConnectorListResource)(nil)
	_ list.ListResourceWithConfigure    = (*ConnectorListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*ConnectorListResource)(nil)
)

type ConnectorListResource struct {
	meta   *meta.Meta
	vpcAPI *vpc.API
}

func (r *ConnectorListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.vpcAPI = vpc.NewAPI(meta.ExtractScwClient(m))
}

func NewConnectorListResource() list.ListResource {
	return &ConnectorListResource{}
}

func (r *ConnectorListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":            listscw.NameAttribute("Name of the VPC connector to filter for"),
			"tags":            listscw.TagsAttribute("Tags of the VPC connector to filter for"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter for."),
			"regions":         listscw.RegionsAttribute("Regions to filter for."),
			"vpc_id": schema.StringAttribute{
				Description: "Filter for connectors attached to this source VPC (regional ID or plain UUID).",
				Optional:    true,
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
			"target_vpc_id": schema.StringAttribute{
				Description: "Filter for connectors attached to this target VPC (regional ID or plain UUID).",
				Optional:    true,
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
		},
	}
}

func (r *ConnectorListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	connectorResource := ResourceConnector()

	resp.ProtoV6Schema = translate.Schema(connectorResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(connectorResource.ProtoIdentitySchema(ctx)())
}

type ConnectorListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Regions        types.List   `tfsdk:"regions"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	VpcID          types.String `tfsdk:"vpc_id"`
	TargetVpcID    types.String `tfsdk:"target_vpc_id"`
}

func (m *ConnectorListResourceModel) GetTags() types.List     { return m.Tags }
func (m *ConnectorListResourceModel) GetRegions() types.List  { return m.Regions }
func (m *ConnectorListResourceModel) GetProjects() types.List { return m.ProjectIDs }

func (r *ConnectorListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_connector"
}

func (r *ConnectorListResource) FetchConnectors(ctx context.Context, region scw.Region, project *string, tags []string, data ConnectorListResourceModel) ([]*vpc.VPCConnector, error) {
	listRequest := &vpc.ListVPCConnectorsRequest{
		Region:         region,
		Name:           data.Name.ValueStringPointer(),
		Tags:           tags,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      project,
		VpcID:          locality.ExpandFrameworkID(data.VpcID),
		TargetVpcID:    locality.ExpandFrameworkID(data.TargetVpcID),
	}

	response, err := r.vpcAPI.ListVPCConnectors(listRequest, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.VpcConnectors, nil
}

func (r *ConnectorListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ConnectorListResourceModel

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

	allConnectors, err := listscw.FetchConcurrently(ctx, listscw.RegionalProjectTargets(regions, projects),
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*vpc.VPCConnector, error) {
			return r.FetchConnectors(ctx, target.Region, &target.ProjectID, tags, data)
		},
		func(a, b *vpc.VPCConnector) int {
			return listscw.CompareRegionalProjectItems(a.ProjectID, b.ProjectID, a.Region, b.Region, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing VPC Connectors", "Failed to list VPC Connectors: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, connector := range allConnectors {
			result := req.NewListResult(ctx)
			result.DisplayName = connector.Name

			connectorResource := ResourceConnector()
			resourceData := connectorResource.Data(&terraform.InstanceState{})

			err = identity.SetRegionalIdentity(resourceData, connector.Region, connector.ID)
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

			tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
			if errIdentityState != nil {
				result.Diagnostics.AddError(
					"Converting identity data",
					"An error was encountered when converting the identity data: "+errIdentityState.Error(),
				)
			}

			identitySetDiags := result.Identity.Set(ctx, *tfTypeIdentity)
			result.Diagnostics.Append(identitySetDiags...)

			sdkDiags := setConnectorState(resourceData, connector)
			if sdkDiags.HasError() {
				tflog.Error(ctx, "error from setting VPC connector state")

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
