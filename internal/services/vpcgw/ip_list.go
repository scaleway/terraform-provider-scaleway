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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*IPListResource)(nil)
	_ list.ListResourceWithConfigure    = (*IPListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*IPListResource)(nil)
)

type IPListResource struct {
	meta     *meta.Meta
	vpcgwAPI *vpcgw.API
}

func (r *IPListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.vpcgwAPI = vpcgw.NewAPI(meta.ExtractScwClient(m))
}

func NewIPListResource() list.ListResource {
	return &IPListResource{}
}

func (r *IPListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tags":            listscw.TagsAttribute("Tags to filter for."),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for."),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter for."),
			"zones":           listscw.ZonesAttribute("Zones to filter for."),
			"reverse": schema.StringAttribute{
				Description: "Filter for IPs whose reverse DNS contains this substring",
				Optional:    true,
			},
			"is_free": schema.BoolAttribute{
				Description: "Filter based on whether the IP is attached to a gateway or not",
				Optional:    true,
			},
		},
	}
}

func (r *IPListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	ipResource := ResourceIP()

	resp.ProtoV6Schema = translate.Schema(ipResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(ipResource.ProtoIdentitySchema(ctx)())
}

type IPListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	Zones          types.List   `tfsdk:"zones"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Reverse        types.String `tfsdk:"reverse"`
	IsFree         types.Bool   `tfsdk:"is_free"`
}

func (m *IPListResourceModel) GetTags() types.List     { return m.Tags }
func (m *IPListResourceModel) GetZones() types.List    { return m.Zones }
func (m *IPListResourceModel) GetProjects() types.List { return m.ProjectIDs }

func (r *IPListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_public_gateway_ip"
}

func (r *IPListResource) FetchIPs(ctx context.Context, zone scw.Zone, project *string, tags []string, data IPListResourceModel) ([]*vpcgw.IP, error) {
	listRequest := &vpcgw.ListIPsRequest{
		Zone:           zone,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      project,
		Tags:           tags,
	}

	if !data.Reverse.IsNull() && !data.Reverse.IsUnknown() {
		if v := data.Reverse.ValueString(); v != "" {
			listRequest.Reverse = &v
		}
	}

	if !data.IsFree.IsNull() && !data.IsFree.IsUnknown() {
		listRequest.IsFree = data.IsFree.ValueBoolPointer()
	}

	response, err := r.vpcgwAPI.ListIPs(listRequest, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.IPs, nil
}

func (r *IPListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data IPListResourceModel

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

	allIPs, err := listscw.FetchConcurrently(ctx, listscw.ZonalProjectTargets(zones, projects),
		func(ctx context.Context, target listscw.ZonalFetchTarget) ([]*vpcgw.IP, error) {
			return r.FetchIPs(ctx, target.Zone, &target.ProjectID, tags, data)
		},
		func(a, b *vpcgw.IP) int {
			return listscw.CompareZonalProjectItems(a.ProjectID, b.ProjectID, a.Zone, b.Zone, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Public Gateway IPs", "Failed to list Public Gateway IPs: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, ip := range allIPs {
			result := req.NewListResult(ctx)
			result.DisplayName = ip.Address.String()

			ipResource := ResourceIP()
			resourceData := ipResource.Data(&terraform.InstanceState{})

			err = identity.SetZonalIdentity(resourceData, ip.Zone, ip.ID)
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

			sdkDiags := setIPState(resourceData, ip)
			if sdkDiags.HasError() {
				tflog.Error(ctx, "error from setting public gateway IP state")

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
