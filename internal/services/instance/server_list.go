package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	instanceV2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
)

var (
	_ list.ListResource                 = (*ServerListResource)(nil)
	_ list.ListResourceWithConfigure    = (*ServerListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*ServerListResource)(nil)
)

type ServerListResource struct {
	meta        *meta.Meta
	instanceAPI *instancehelpers.BlockAndInstanceAPI
}

func (r *ServerListResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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
	r.instanceAPI = instancehelpers.NewBlockAndInstanceAPI(meta.ExtractScwClient(m))
}

func NewServerListResource() list.ListResource {
	return &ServerListResource{}
}

func (r *ServerListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zones":       listscw.ZonesAttribute("Zones of the Server to list for"),
			"project_ids": listscw.ProjectIDsAttribute("Project IDs of the Server to list for"),
			"name":        listscw.NameAttribute("Name of the Server to list for"),
			"server_type": schema.StringAttribute{
				Description: "The type of server to list for",
				Optional:    true,
			},
			"tags": listscw.TagsAttribute("Tags of the Server to list for"),
			"security_group_ids": schema.ListAttribute{
				Description: "Filter server list by security group IDs",
				Optional:    true,
				ElementType: types.StringType,
			},
			"placement_group_ids": schema.ListAttribute{
				Description: "Filter server list by placement group IDs",
				Optional:    true,
				ElementType: types.StringType,
			},
			"private_network_ids": schema.ListAttribute{
				Description: "Filter server list by private network IDs",
				Optional:    true,
				ElementType: types.StringType,
			},
			"mac_addresses": schema.ListAttribute{
				Description: "Filter server list by MAC addresses",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *ServerListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	resourceServer := ResourceServer()

	resp.ProtoV6Schema = translate.Schema(resourceServer.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(resourceServer.ProtoIdentitySchema(ctx)())
}

type ListResourceModel struct {
	Zones             types.List   `tfsdk:"zones"`
	ProjectIDs        types.List   `tfsdk:"project_ids"`
	Name              types.String `tfsdk:"name"`
	ServerType        types.String `tfsdk:"server_type"`
	Tags              types.List   `tfsdk:"tags"`
	SecurityGroupIDs  types.List   `tfsdk:"security_group_ids"`
	PlacementGroupIDs types.List   `tfsdk:"placement_group_ids"`
	PrivateNetworkIDs types.List   `tfsdk:"private_network_ids"`
	MacAddresses      types.List   `tfsdk:"mac_addresses"`
}

func (m *ListResourceModel) GetTags() types.List {
	return m.Tags
}

func (m *ListResourceModel) GetZones() types.List {
	return m.Zones
}

func (m *ListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (r *ServerListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_server"
}

func (r *ServerListResource) FetchServers(ctx context.Context, zone scw.Zone, project string, tags []string, data ListResourceModel) ([]*instanceV2.Server, error) {
	var diags diag.Diagnostics

	listRequest := &instanceV2.ListServersRequest{
		Zone:       zone,
		ProjectID:  project,
		Name:       data.Name.ValueStringPointer(),
		ServerType: data.ServerType.ValueStringPointer(),
		Tags:       tags,
	}

	if !data.SecurityGroupIDs.IsNull() {
		var sgIDs []string

		diags = data.SecurityGroupIDs.ElementsAs(ctx, sgIDs, false)

		if !diags.HasError() {
			listRequest.SecurityGroupIDs = sgIDs
		}
	}

	if !data.PlacementGroupIDs.IsNull() {
		var pgIDs []string

		diags = append(diags, data.PlacementGroupIDs.ElementsAs(ctx, pgIDs, false)...)

		if !diags.HasError() {
			listRequest.PlacementGroupIDs = pgIDs
		}
	}

	if !data.PrivateNetworkIDs.IsNull() {
		var pnIDs []string

		diags = append(diags, data.PrivateNetworkIDs.ElementsAs(ctx, pnIDs, false)...)

		if !diags.HasError() {
			listRequest.PrivateNetworkIDs = pnIDs
		}
	}

	if !data.MacAddresses.IsNull() {
		var macAddrs []string

		diags = append(diags, data.MacAddresses.ElementsAs(ctx, macAddrs, false)...)

		if !diags.HasError() {
			listRequest.MacAddresses = macAddrs
		}
	}

	if diags.HasError() {
		return nil, fmt.Errorf("%v", diags) // TODO: probably not good
	}

	response, err := r.instanceAPI.InstanceV2API.ListServers(listRequest, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	servers := make([]*instanceV2.Server, 0, response.TotalCount)

	for _, serverSummary := range response.Servers {
		serverComplete, err := r.instanceAPI.InstanceV2API.GetServer(&instanceV2.GetServerRequest{
			Zone:     zone,
			ServerID: serverSummary.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to get server: %w", err)
		}

		servers = append(servers, serverComplete)
	}

	return servers, nil
}

func (r *ServerListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
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

	var targets []listscw.ZonalFetchTarget

	for _, r := range zones {
		for _, p := range projects {
			targets = append(targets, listscw.ZonalFetchTarget{Zone: r, ProjectID: p})
		}
	}

	allServers, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target listscw.ZonalFetchTarget) ([]*instanceV2.Server, error) {
			return r.FetchServers(ctx, target.Zone, target.ProjectID, tags, data)
		},
		func(a, b *instanceV2.Server) int {
			return listscw.CompareZonalProjectItems(a.ProjectID, b.ProjectID, a.Zone, b.Zone, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing Servers", "Failed to list Servers: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, rawServer := range allServers {
			result := req.NewListResult(ctx)
			result.DisplayName = rawServer.Name

			serverResource := ResourceServer()
			resourceData := serverResource.Data(&terraform.InstanceState{})

			err := identity.SetZonalIdentity(resourceData, rawServer.Zone, rawServer.ID)
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

			diagsState := setServerState(ctx, resourceData, rawServer, r.instanceAPI, rawServer.Zone, rawServer.ID)
			if diagsState.HasError() {
				tflog.Error(ctx, "error from setting server state")

				for _, d := range diagsState {
					result.Diagnostics.AddError(d.Summary, d.Detail)
				}

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
