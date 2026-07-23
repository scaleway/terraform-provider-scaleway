package applesilicon

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	applesilicon "github.com/scaleway/scaleway-sdk-go/api/applesilicon/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*ServerListResource)(nil)
	_ list.ListResourceWithConfigure    = (*ServerListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*ServerListResource)(nil)
)

type ServerListResource struct {
	meta  *meta.Meta
	asAPI *applesilicon.API
}

func (r *ServerListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.asAPI = applesilicon.NewAPI(meta.ExtractScwClient(m))
}

func NewServerListResource() list.ListResource {
	return &ServerListResource{}
}

func (r *ServerListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apple_silicon_server"
}

func (r *ServerListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zones":           listscw.ZonesAttribute("Zones of the apple silicon server to filter on"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs of the apple silicon server to filter on"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID of the apple silicon server to filter on"),
		},
	}
}

func (r *ServerListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	serverResource := ResourceServer()

	resp.ProtoV6Schema = translate.Schema(serverResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(serverResource.ProtoIdentitySchema(ctx)())
}

type ServerListResourceModel struct {
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Zones          types.List   `tfsdk:"zones"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (m *ServerListResourceModel) GetZones() types.List {
	return m.Zones
}

func (m *ServerListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

type appleSiliconServerRow struct {
	Server    *applesilicon.Server
	Zone      scw.Zone
	ProjectID string
}

type appleSiliconServerListTarget struct {
	Zone      scw.Zone
	ProjectID string
}

func (r *ServerListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ServerListResourceModel

	diags := req.Config.Get(ctx, &data)
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

	targets := make([]appleSiliconServerListTarget, 0, len(zones)*len(projects))
	for _, zone := range zones {
		for _, project := range projects {
			targets = append(targets, appleSiliconServerListTarget{
				Zone:      zone,
				ProjectID: project,
			})
		}
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target appleSiliconServerListTarget) ([]appleSiliconServerRow, error) {
			return r.fetchServerRows(ctx, target, data)
		},
		func(a, b appleSiliconServerRow) int {
			return listscw.CompareZonalProjectItems(a.ProjectID, b.ProjectID, a.Zone, b.Zone, a.Server.ID, b.Server.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing apple silicon servers", "Failed to list apple silicon servers: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = row.Server.Name

			serverResource := ResourceServer()
			resourceData := serverResource.Data(&terraform.InstanceState{})

			err := identity.SetZonalIdentity(resourceData, row.Zone, row.Server.ID)
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

			setServerState(resourceData, row.Server)

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

func (r *ServerListResource) fetchServerRows(ctx context.Context, target appleSiliconServerListTarget, data ServerListResourceModel) ([]appleSiliconServerRow, error) {
	listReq := &applesilicon.ListServersRequest{
		Zone:      target.Zone,
		ProjectID: &target.ProjectID,
	}

	if !data.OrganizationID.IsNull() && !data.OrganizationID.IsUnknown() {
		orgID := data.OrganizationID.ValueString()
		listReq.OrganizationID = &orgID
	}

	resp, err := r.asAPI.ListServers(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	rows := make([]appleSiliconServerRow, 0, len(resp.Servers))
	for _, server := range resp.Servers {
		if server == nil {
			continue
		}

		rows = append(rows, appleSiliconServerRow{
			Zone:      target.Zone,
			ProjectID: target.ProjectID,
			Server:    server,
		})
	}

	return rows, nil
}

func setServerState(resourceData *sdkschema.ResourceData, server *applesilicon.Server) {
	if server == nil {
		return
	}

	_ = resourceData.Set("name", server.Name)
	_ = resourceData.Set("type", server.Type)
	_ = resourceData.Set("state", server.Status.String())
	_ = resourceData.Set("created_at", server.CreatedAt.Format(time.RFC3339))
	_ = resourceData.Set("updated_at", server.UpdatedAt.Format(time.RFC3339))
	_ = resourceData.Set("deletable_at", server.DeletableAt.Format(time.RFC3339))
	_ = resourceData.Set("ip", server.IP.String())
	_ = resourceData.Set("vnc_url", server.VncURL)
	_ = resourceData.Set("vpc_status", server.VpcStatus)
	_ = resourceData.Set("zone", server.Zone.String())
	_ = resourceData.Set("organization_id", server.OrganizationID)
	_ = resourceData.Set("project_id", server.ProjectID)
	_ = resourceData.Set("password", server.SudoPassword)
	_ = resourceData.Set("username", server.SSHUsername)
	_ = resourceData.Set("public_bandwidth", int(server.PublicBandwidthBps))
	_ = resourceData.Set("runner_ids", server.AppliedRunnerConfigurationIDs)

	switch server.VpcStatus {
	case applesilicon.ServerPrivateNetworkStatusVpcDisabled:
		_ = resourceData.Set("enable_vpc", false)
	case applesilicon.ServerPrivateNetworkStatusVpcEnabled:
		_ = resourceData.Set("enable_vpc", true)
	}

	if server.Commitment != nil {
		switch server.Commitment.Type {
		case applesilicon.CommitmentTypeNone, applesilicon.CommitmentTypeDuration24h:
			_ = resourceData.Set("commitment", applesilicon.CommitmentTypeDuration24h.String())
		case applesilicon.CommitmentTypeRenewedMonthly:
			_ = resourceData.Set("commitment", applesilicon.CommitmentTypeRenewedMonthly.String())
		}
	}
}
