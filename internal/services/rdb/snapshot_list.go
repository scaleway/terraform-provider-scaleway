package rdb

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ list.ListResource                 = (*SnapshotListResource)(nil)
	_ list.ListResourceWithConfigure    = (*SnapshotListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*SnapshotListResource)(nil)
)

type SnapshotListResource struct {
	meta   *meta.Meta
	rdbAPI *rdbSDK.API
}

func (r *SnapshotListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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
	r.rdbAPI = rdbSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewSnapshotListResource() list.ListResource {
	return &SnapshotListResource{}
}

func (r *SnapshotListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_snapshot"
}

func (r *SnapshotListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions": listscw.RegionsAttribute("Regions of the RDB Database Instance to filter snapshots on"),
			"project_ids": listscw.ProjectIDsAttribute(
				"Project IDs of the RDB Database Instance to filter snapshots on",
			),
			"instance_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Database Instance IDs to list snapshots from. Use [\"*\"] only to include all instances in each selected region and project. Otherwise each value must be a regional ID (`region/uuid`) or a bare instance UUID.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.Any(
							stringvalidator.OneOf("*"),
							verify.IsStringUUIDOrUUIDWithLocality(),
						),
					),
				},
			},
			"name": listscw.NameAttribute("Name of the snapshot to filter on"),
		},
	}
}

func (r *SnapshotListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	snapshotResource := ResourceSnapshot()

	resp.ProtoV6Schema = translate.Schema(snapshotResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(snapshotResource.ProtoIdentitySchema(ctx)())
}

type SnapshotListResourceModel struct {
	InstanceIDs types.List   `tfsdk:"instance_ids"`
	ProjectIDs  types.List   `tfsdk:"project_ids"`
	Regions     types.List   `tfsdk:"regions"`
	Name        types.String `tfsdk:"name"`
}

func (m *SnapshotListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *SnapshotListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

type rdbSnapshotRow struct {
	Snapshot  *rdbSDK.Snapshot
	Region    scw.Region
	ProjectID string
}

type snapshotListTarget struct {
	Region     scw.Region
	ProjectID  string
	InstanceID string
}

func (r *SnapshotListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data SnapshotListResourceModel

	diags := req.Config.Get(ctx, &data)
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

	var instanceIDElems []string

	diags = data.InstanceIDs.ElementsAs(ctx, &instanceIDElems, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	if len(instanceIDElems) == 0 {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid instance_ids", "`instance_ids` must contain at least one element."),
		})

		return
	}

	if slices.Contains(instanceIDElems, "*") && len(instanceIDElems) != 1 {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid instance_ids", `When using "*", instance_ids must be exactly ["*"].`),
		})

		return
	}

	targets, targetDiags := r.buildSnapshotListTargets(ctx, instanceIDElems, regions, projects)
	if targetDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(targetDiags)

		return
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target snapshotListTarget) ([]rdbSnapshotRow, error) {
			return r.fetchSnapshotRows(ctx, target, data)
		},
		func(a, b rdbSnapshotRow) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Region != b.Region {
				return strings.Compare(string(a.Region), string(b.Region))
			}

			if a.Snapshot.InstanceID != b.Snapshot.InstanceID {
				return strings.Compare(a.Snapshot.InstanceID, b.Snapshot.InstanceID)
			}

			return strings.Compare(a.Snapshot.ID, b.Snapshot.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing RDB snapshots", "Failed to list RDB snapshots: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = row.Snapshot.Name

			snapshotResource := ResourceSnapshot()
			resourceData := snapshotResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, row.Region, row.Snapshot.ID)
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

			setSnapshotState(resourceData, row.Snapshot)

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

func (r *SnapshotListResource) buildSnapshotListTargets(
	ctx context.Context,
	instanceIDElems []string,
	regions []scw.Region,
	projects []string,
) ([]snapshotListTarget, diag.Diagnostics) {
	var diags diag.Diagnostics

	if instanceIDElems[0] == "*" {
		targets := make([]snapshotListTarget, 0, len(regions)*len(projects))
		for _, region := range regions {
			for _, project := range projects {
				targets = append(targets, snapshotListTarget{
					Region:     region,
					ProjectID:  project,
					InstanceID: "",
				})
			}
		}

		return targets, diags
	}

	projectSet := make(map[string]struct{}, len(projects))
	for _, p := range projects {
		projectSet[p] = struct{}{}
	}

	targets := make([]snapshotListTarget, 0, len(instanceIDElems))

	for _, rawID := range instanceIDElems {
		region, instanceUUID, err := regional.ParseID(rawID)
		if err != nil {
			diags.AddError("Invalid instance_ids", fmt.Sprintf("Could not parse instance id %q: %v", rawID, err))

			continue
		}

		if !slices.Contains(regions, region) {
			diags.AddError(
				"Invalid instance_ids",
				fmt.Sprintf("Instance %q is in region %q which is not included in the configured regions for this list.", rawID, region),
			)

			continue
		}

		inst, err := r.rdbAPI.GetInstance(&rdbSDK.GetInstanceRequest{
			Region:     region,
			InstanceID: instanceUUID,
		}, scw.WithContext(ctx))
		if err != nil {
			diags.AddError("Listing RDB snapshots", fmt.Sprintf("Could not load instance %q: %v", rawID, err))

			continue
		}

		if _, ok := projectSet[inst.ProjectID]; !ok {
			diags.AddError(
				"Invalid instance_ids",
				fmt.Sprintf("Instance %q belongs to project %q which is not included in the configured project_ids for this list.", rawID, inst.ProjectID),
			)

			continue
		}

		targets = append(targets, snapshotListTarget{
			Region:     region,
			ProjectID:  inst.ProjectID,
			InstanceID: instanceUUID,
		})
	}

	return targets, diags
}

func (r *SnapshotListResource) fetchSnapshotRows(ctx context.Context, target snapshotListTarget, data SnapshotListResourceModel) ([]rdbSnapshotRow, error) {
	if target.InstanceID == "" {
		return r.fetchSnapshotRowsForAllInstances(ctx, target, data)
	}

	return r.fetchSnapshotRowsForInstance(ctx, target.Region, target.ProjectID, target.InstanceID, data)
}

func (r *SnapshotListResource) fetchSnapshotRowsForAllInstances(ctx context.Context, target snapshotListTarget, data SnapshotListResourceModel) ([]rdbSnapshotRow, error) {
	listReq := &rdbSDK.ListInstancesRequest{
		Region:    target.Region,
		ProjectID: &target.ProjectID,
	}

	instances, err := r.rdbAPI.ListInstances(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	var rows []rdbSnapshotRow

	for _, inst := range instances.Instances {
		if inst == nil || inst.ProjectID != target.ProjectID {
			continue
		}

		part, err := r.fetchSnapshotRowsForInstance(ctx, target.Region, target.ProjectID, inst.ID, data)
		if err != nil {
			return nil, err
		}

		rows = append(rows, part...)
	}

	return rows, nil
}

func (r *SnapshotListResource) fetchSnapshotRowsForInstance(
	ctx context.Context,
	region scw.Region,
	projectID string,
	instanceID string,
	data SnapshotListResourceModel,
) ([]rdbSnapshotRow, error) {
	listReq := &rdbSDK.ListSnapshotsRequest{
		Region:     region,
		InstanceID: &instanceID,
		ProjectID:  &projectID,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		n := strings.TrimSpace(data.Name.ValueString())
		if n != "" {
			listReq.Name = &n
		}
	}

	resp, err := r.rdbAPI.ListSnapshots(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	rows := make([]rdbSnapshotRow, 0, len(resp.Snapshots))
	for _, snapshot := range resp.Snapshots {
		if snapshot == nil {
			continue
		}

		rows = append(rows, rdbSnapshotRow{
			Region:    region,
			ProjectID: projectID,
			Snapshot:  snapshot,
		})
	}

	return rows, nil
}
