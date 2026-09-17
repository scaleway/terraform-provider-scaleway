package block

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*VolumeListResource)(nil)
	_ list.ListResourceWithConfigure    = (*VolumeListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*VolumeListResource)(nil)
)

type VolumeListResource struct {
	meta     *meta.Meta
	blockAPI *blockSDK.API
}

func (r *VolumeListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.blockAPI = blockSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewVolumeListResource() list.ListResource {
	return &VolumeListResource{}
}

func (r *VolumeListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_volume"
}

func (r *VolumeListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zones": listscw.ZonesAttribute("Zones of the block volume to filter on"),
			"project_ids": listscw.ProjectIDsAttribute(
				"Project IDs of the block volume to filter on",
			),
			"name":            listscw.NameAttribute("Name of the volume to filter on"),
			"tags":            listscw.TagsAttribute("Tags of the volume to filter on"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID of the volume to filter on"),
		},
	}
}

func (r *VolumeListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	volumeResource := ResourceVolume()

	resp.ProtoV6Schema = translate.Schema(volumeResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(volumeResource.ProtoIdentitySchema(ctx)())
}

type VolumeListResourceModel struct {
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Zones          types.List   `tfsdk:"zones"`
	Name           types.String `tfsdk:"name"`
	Tags           types.List   `tfsdk:"tags"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (m *VolumeListResourceModel) GetZones() types.List {
	return m.Zones
}

func (m *VolumeListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (m *VolumeListResourceModel) GetTags() types.List {
	return m.Tags
}

type blockVolumeRow struct {
	Volume    *blockSDK.Volume
	Zone      scw.Zone
	ProjectID string
}

type volumeListTarget struct {
	Zone      scw.Zone
	ProjectID string
}

func (r *VolumeListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data VolumeListResourceModel

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

	targets := make([]volumeListTarget, 0, len(zones)*len(projects))
	for _, zone := range zones {
		for _, project := range projects {
			targets = append(targets, volumeListTarget{
				Zone:      zone,
				ProjectID: project,
			})
		}
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target volumeListTarget) ([]blockVolumeRow, error) {
			return r.fetchVolumeRows(ctx, target, data)
		},
		func(a, b blockVolumeRow) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Zone != b.Zone {
				return strings.Compare(string(a.Zone), string(b.Zone))
			}

			if a.Volume.ID != b.Volume.ID {
				return strings.Compare(a.Volume.ID, b.Volume.ID)
			}

			return 0
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing block volumes", "Failed to list block volumes: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = row.Volume.Name

			volumeResource := ResourceVolume()
			resourceData := volumeResource.Data(&terraform.InstanceState{})

			err := identity.SetZonalIdentity(resourceData, row.Zone, row.Volume.ID)
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

			setVolumeState(r.blockAPI, resourceData, row.Volume)

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

func (r *VolumeListResource) fetchVolumeRows(ctx context.Context, target volumeListTarget, data VolumeListResourceModel) ([]blockVolumeRow, error) {
	listReq := &blockSDK.ListVolumesRequest{
		Zone:      target.Zone,
		ProjectID: &target.ProjectID,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		n := strings.TrimSpace(data.Name.ValueString())
		if n != "" {
			listReq.Name = &n
		}
	}

	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tagStrings []string
		data.Tags.ElementsAs(ctx, &tagStrings, false)

		if len(tagStrings) > 0 {
			listReq.Tags = tagStrings
		}
	}

	if !data.OrganizationID.IsNull() && !data.OrganizationID.IsUnknown() {
		orgID := data.OrganizationID.ValueString()
		listReq.OrganizationID = &orgID
	}

	resp, err := r.blockAPI.ListVolumes(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	rows := make([]blockVolumeRow, 0, len(resp.Volumes))
	for _, volume := range resp.Volumes {
		if volume == nil {
			continue
		}

		rows = append(rows, blockVolumeRow{
			Zone:      target.Zone,
			ProjectID: target.ProjectID,
			Volume:    volume,
		})
	}

	return rows, nil
}
