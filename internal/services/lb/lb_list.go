package lb

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
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*LbListResource)(nil)
	_ list.ListResourceWithConfigure    = (*LbListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*LbListResource)(nil)
)

type LbListResource struct {
	meta  *meta.Meta
	lbAPI *lbSDK.ZonedAPI
}

func (r *LbListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.lbAPI = lbSDK.NewZonedAPI(meta.ExtractScwClient(m))
}

func NewLbListResource() list.ListResource {
	return &LbListResource{}
}

func (r *LbListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":            listscw.NameAttribute("Name of the Load Balancer to filter for"),
			"tags":            listscw.TagsAttribute("Tags of the Load Balancer to filter for"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter for."),
			"zones":           listscw.ZonesAttribute("Zones to filter for."),
		},
	}
}

func (r *LbListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	lbResource := ResourceLb()

	resp.ProtoV6Schema = translate.Schema(lbResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(lbResource.ProtoIdentitySchema(ctx)())
}

type LbListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	Zones          types.List   `tfsdk:"zones"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (m *LbListResourceModel) GetTags() types.List     { return m.Tags }
func (m *LbListResourceModel) GetZones() types.List    { return m.Zones }
func (m *LbListResourceModel) GetProjects() types.List { return m.ProjectIDs }

func (r *LbListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb"
}

func (r *LbListResource) FetchLBs(ctx context.Context, zone scw.Zone, project *string, tags []string, data LbListResourceModel) ([]*lbSDK.LB, error) {
	listRequest := &lbSDK.ZonedAPIListLBsRequest{
		Zone:           zone,
		Name:           data.Name.ValueStringPointer(),
		Tags:           tags,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      project,
	}

	response, err := r.lbAPI.ListLBs(listRequest, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.LBs, nil
}

func (r *LbListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data LbListResourceModel

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

	allLBs, err := listscw.FetchConcurrently(ctx, listscw.ZonalProjectTargets(zones, projects),
		func(ctx context.Context, target listscw.ZonalFetchTarget) ([]*lbSDK.LB, error) {
			return r.FetchLBs(ctx, target.Zone, &target.ProjectID, tags, data)
		},
		func(a, b *lbSDK.LB) int {
			return listscw.CompareZonalProjectItems(a.ProjectID, b.ProjectID, a.Zone, b.Zone, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Load Balancers", "Failed to list Load Balancers: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, loadbalancer := range allLBs {
			result := req.NewListResult(ctx)
			result.DisplayName = loadbalancer.Name

			lbResource := ResourceLb()
			resourceData := lbResource.Data(&terraform.InstanceState{})

			err = identity.SetZonalIdentity(resourceData, loadbalancer.Zone, loadbalancer.ID)
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

			sdkDiags := setLBState(ctx, resourceData, r.meta, r.lbAPI, loadbalancer, false)
			if sdkDiags.HasError() {
				tflog.Error(ctx, "error from setting load balancer state")

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
