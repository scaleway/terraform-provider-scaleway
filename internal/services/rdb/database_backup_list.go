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
	_ list.ListResource                 = (*DatabaseBackupListResource)(nil)
	_ list.ListResourceWithConfigure    = (*DatabaseBackupListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*DatabaseBackupListResource)(nil)
)

type DatabaseBackupListResource struct {
	meta   *meta.Meta
	rdbAPI *rdbSDK.API
}

func (r *DatabaseBackupListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func NewDatabaseBackupListResource() list.ListResource {
	return &DatabaseBackupListResource{}
}

func (r *DatabaseBackupListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_database_backup"
}

func (r *DatabaseBackupListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions": listscw.RegionsAttribute("Regions of the RDB Database Instance to filter backups on"),
			"project_ids": listscw.ProjectIDsAttribute(
				"Project IDs of the RDB Database Instance to filter backups on",
			),
			"instance_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Database Instance IDs to list backups from. Use [\"*\"] only to include all instances in each selected region and project. Otherwise each value must be a regional ID (`region/uuid`) or a bare instance UUID.",
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
			"name": listscw.NameAttribute("Name of the database backup to filter on"),
		},
	}
}

func (r *DatabaseBackupListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	backupResource := ResourceDatabaseBackup()

	resp.ProtoV6Schema = translate.Schema(backupResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(backupResource.ProtoIdentitySchema(ctx)())
}

type DatabaseBackupListResourceModel struct {
	InstanceIDs types.List   `tfsdk:"instance_ids"`
	ProjectIDs  types.List   `tfsdk:"project_ids"`
	Regions     types.List   `tfsdk:"regions"`
	Name        types.String `tfsdk:"name"`
}

func (m *DatabaseBackupListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *DatabaseBackupListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

type rdbBackupRow struct {
	Backup    *rdbSDK.DatabaseBackup
	ProjectID string
}

type backupListTarget struct {
	Region     scw.Region
	ProjectID  string
	InstanceID string
}

func (r *DatabaseBackupListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data DatabaseBackupListResourceModel

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

	targets, targetDiags := r.buildBackupListTargets(ctx, instanceIDElems, regions, projects)
	if targetDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(targetDiags)

		return
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target backupListTarget) ([]rdbBackupRow, error) {
			return r.fetchBackupRows(ctx, target, data)
		},
		func(a, b rdbBackupRow) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Backup.Region != b.Backup.Region {
				return strings.Compare(string(a.Backup.Region), string(b.Backup.Region))
			}

			if a.Backup.InstanceID != b.Backup.InstanceID {
				return strings.Compare(a.Backup.InstanceID, b.Backup.InstanceID)
			}

			return strings.Compare(a.Backup.ID, b.Backup.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing RDB database backups", "Failed to list RDB database backups: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = row.Backup.Name

			backupResource := ResourceDatabaseBackup()
			resourceData := backupResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, row.Backup.Region, row.Backup.ID)
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

			setDatabaseBackupState(resourceData, row.Backup.Region, row.Backup)

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

func (r *DatabaseBackupListResource) buildBackupListTargets(
	ctx context.Context,
	instanceIDElems []string,
	regions []scw.Region,
	projects []string,
) ([]backupListTarget, diag.Diagnostics) {
	var diags diag.Diagnostics

	if instanceIDElems[0] == "*" {
		targets := make([]backupListTarget, 0, len(regions)*len(projects))
		for _, region := range regions {
			for _, project := range projects {
				targets = append(targets, backupListTarget{
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

	targets := make([]backupListTarget, 0, len(instanceIDElems))

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
			diags.AddError("Listing RDB database backups", fmt.Sprintf("Could not load instance %q: %v", rawID, err))

			continue
		}

		if _, ok := projectSet[inst.ProjectID]; !ok {
			diags.AddError(
				"Invalid instance_ids",
				fmt.Sprintf("Instance %q belongs to project %q which is not included in the configured project_ids for this list.", rawID, inst.ProjectID),
			)

			continue
		}

		targets = append(targets, backupListTarget{
			Region:     region,
			ProjectID:  inst.ProjectID,
			InstanceID: instanceUUID,
		})
	}

	return targets, diags
}

func (r *DatabaseBackupListResource) fetchBackupRows(ctx context.Context, target backupListTarget, data DatabaseBackupListResourceModel) ([]rdbBackupRow, error) {
	if target.InstanceID == "" {
		return r.fetchBackupRowsForAllInstances(ctx, target, data)
	}

	return r.fetchBackupRowsForInstance(ctx, target.Region, target.ProjectID, target.InstanceID, data)
}

func (r *DatabaseBackupListResource) fetchBackupRowsForAllInstances(ctx context.Context, target backupListTarget, data DatabaseBackupListResourceModel) ([]rdbBackupRow, error) {
	listReq := &rdbSDK.ListInstancesRequest{
		Region:    target.Region,
		ProjectID: &target.ProjectID,
	}

	instances, err := r.rdbAPI.ListInstances(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	var rows []rdbBackupRow

	for _, inst := range instances.Instances {
		if inst == nil || inst.ProjectID != target.ProjectID {
			continue
		}

		part, err := r.fetchBackupRowsForInstance(ctx, target.Region, target.ProjectID, inst.ID, data)
		if err != nil {
			return nil, err
		}

		rows = append(rows, part...)
	}

	return rows, nil
}

func (r *DatabaseBackupListResource) fetchBackupRowsForInstance(
	ctx context.Context,
	region scw.Region,
	projectID string,
	instanceID string,
	data DatabaseBackupListResourceModel,
) ([]rdbBackupRow, error) {
	listReq := &rdbSDK.ListDatabaseBackupsRequest{
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

	resp, err := r.rdbAPI.ListDatabaseBackups(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	rows := make([]rdbBackupRow, 0, len(resp.DatabaseBackups))
	for _, backup := range resp.DatabaseBackups {
		if backup == nil {
			continue
		}

		rows = append(rows, rdbBackupRow{
			ProjectID: projectID,
			Backup:    backup,
		})
	}

	return rows, nil
}
