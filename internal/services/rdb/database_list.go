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
	_ list.ListResource                 = (*DatabaseListResource)(nil)
	_ list.ListResourceWithConfigure    = (*DatabaseListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*DatabaseListResource)(nil)
)

type DatabaseListResource struct {
	meta   *meta.Meta
	rdbAPI *rdbSDK.API
}

func (r *DatabaseListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func NewDatabaseListResource() list.ListResource {
	return &DatabaseListResource{}
}

func (r *DatabaseListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_database"
}

func (r *DatabaseListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions":     listscw.RegionsAttribute("Regions of the RDB Database Instance to filter on"),
			"project_ids": listscw.ProjectIDsAttribute("Project IDs of the RDB Database Instance to filter on"),
			"instance_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Database Instance IDs to list databases from. Use [\"*\"] only to include all instances in each selected region and project. Otherwise each value must be a regional ID (`region/uuid`) or a bare instance UUID.",
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
			"name":            listscw.NameAttribute("Name of the database to filter on"),
			"managed":         databaseListManagedAttribute(),
			"owner":           databaseListOwnerAttribute(),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID of the RDB Database Instance to filter on"),
		},
	}
}

func databaseListManagedAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		Description: "Whether to only list managed databases",
		Optional:    true,
	}
}

func databaseListOwnerAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "Owner user name to filter on",
		Optional:    true,
	}
}

func (r *DatabaseListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	dbResource := ResourceDatabase()

	resp.ProtoV6Schema = translate.Schema(dbResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(dbResource.ProtoIdentitySchema(ctx)())
}

type DatabaseListResourceModel struct {
	InstanceIDs    types.List   `tfsdk:"instance_ids"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	Regions        types.List   `tfsdk:"regions"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Owner          types.String `tfsdk:"owner"`
	Managed        types.Bool   `tfsdk:"managed"`
}

func (m *DatabaseListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *DatabaseListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

type rdbDatabaseRow struct {
	Database   *rdbSDK.Database
	Region     scw.Region
	ProjectID  string
	InstanceID string
}

type databaseListTarget struct {
	Region     scw.Region
	ProjectID  string
	InstanceID string
}

func (r *DatabaseListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data DatabaseListResourceModel

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

	targets, targetDiags := r.buildDatabaseListTargets(ctx, instanceIDElems, regions, projects)
	if targetDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(targetDiags)

		return
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target databaseListTarget) ([]rdbDatabaseRow, error) {
			return r.fetchDatabaseRows(ctx, target, data)
		},
		func(a, b rdbDatabaseRow) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Region != b.Region {
				return strings.Compare(string(a.Region), string(b.Region))
			}

			if a.InstanceID != b.InstanceID {
				return strings.Compare(a.InstanceID, b.InstanceID)
			}

			return strings.Compare(a.Database.Name, b.Database.Name)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing RDB databases", "Failed to list RDB databases: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = row.Database.Name

			dbResource := ResourceDatabase()
			resourceData := dbResource.Data(&terraform.InstanceState{})

			err := identity.SetMultiPartIdentity(resourceData, map[string]string{
				"region":        row.Region.String(),
				"instance_id":   row.InstanceID,
				"database_name": row.Database.Name,
			}, "region", "instance_id", "database_name")
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

			setDatabaseState(resourceData, row.Region, row.InstanceID, row.Database)

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

func (r *DatabaseListResource) buildDatabaseListTargets(
	ctx context.Context,
	instanceIDElems []string,
	regions []scw.Region,
	projects []string,
) ([]databaseListTarget, diag.Diagnostics) {
	var diags diag.Diagnostics

	if instanceIDElems[0] == "*" {
		targets := make([]databaseListTarget, 0, len(regions)*len(projects))
		for _, region := range regions {
			for _, project := range projects {
				targets = append(targets, databaseListTarget{
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

	targets := make([]databaseListTarget, 0, len(instanceIDElems))

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
			diags.AddError("Listing RDB databases", fmt.Sprintf("Could not load instance %q: %v", rawID, err))

			continue
		}

		if _, ok := projectSet[inst.ProjectID]; !ok {
			diags.AddError(
				"Invalid instance_ids",
				fmt.Sprintf("Instance %q belongs to project %q which is not included in the configured project_ids for this list.", rawID, inst.ProjectID),
			)

			continue
		}

		targets = append(targets, databaseListTarget{
			Region:     region,
			ProjectID:  inst.ProjectID,
			InstanceID: instanceUUID,
		})
	}

	return targets, diags
}

func (r *DatabaseListResource) fetchDatabaseRows(ctx context.Context, target databaseListTarget, data DatabaseListResourceModel) ([]rdbDatabaseRow, error) {
	if target.InstanceID == "" {
		return r.fetchDatabaseRowsForAllInstances(ctx, target, data)
	}

	return r.fetchDatabaseRowsForInstance(ctx, target.Region, target.ProjectID, target.InstanceID, data)
}

func (r *DatabaseListResource) fetchDatabaseRowsForAllInstances(ctx context.Context, target databaseListTarget, data DatabaseListResourceModel) ([]rdbDatabaseRow, error) {
	listReq := &rdbSDK.ListInstancesRequest{
		Region:    target.Region,
		ProjectID: &target.ProjectID,
	}

	if !data.OrganizationID.IsNull() && !data.OrganizationID.IsUnknown() {
		org := data.OrganizationID.ValueString()
		listReq.OrganizationID = &org
	}

	instances, err := r.rdbAPI.ListInstances(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	var rows []rdbDatabaseRow

	for _, inst := range instances.Instances {
		if inst == nil {
			continue
		}

		if inst.ProjectID != target.ProjectID {
			continue
		}

		part, err := r.fetchDatabaseRowsForInstance(ctx, target.Region, target.ProjectID, inst.ID, data)
		if err != nil {
			return nil, err
		}

		rows = append(rows, part...)
	}

	return rows, nil
}

func (r *DatabaseListResource) fetchDatabaseRowsForInstance(
	ctx context.Context,
	region scw.Region,
	projectID string,
	instanceID string,
	data DatabaseListResourceModel,
) ([]rdbDatabaseRow, error) {
	listReq := &rdbSDK.ListDatabasesRequest{
		Region:     region,
		InstanceID: instanceID,
		OrderBy:    rdbSDK.ListDatabasesRequestOrderByNameAsc,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		n := strings.TrimSpace(data.Name.ValueString())
		if n != "" {
			listReq.Name = &n
		}
	}

	if !data.Managed.IsNull() && !data.Managed.IsUnknown() {
		m := data.Managed.ValueBool()
		listReq.Managed = &m
	}

	if !data.Owner.IsNull() && !data.Owner.IsUnknown() {
		o := strings.TrimSpace(data.Owner.ValueString())
		if o != "" {
			listReq.Owner = &o
		}
	}

	resp, err := r.rdbAPI.ListDatabases(listReq, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	rows := make([]rdbDatabaseRow, 0, len(resp.Databases))
	for _, db := range resp.Databases {
		if db == nil {
			continue
		}

		rows = append(rows, rdbDatabaseRow{
			Region:     region,
			ProjectID:  projectID,
			InstanceID: instanceID,
			Database:   db,
		})
	}

	return rows, nil
}
