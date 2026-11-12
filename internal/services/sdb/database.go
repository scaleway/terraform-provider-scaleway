package sdb

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdbSDK "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/sdb_sql_database_resource.md
var databaseResourceDescription string

var (
	_ resource.Resource                = (*DatabaseResource)(nil)
	_ resource.ResourceWithConfigure   = (*DatabaseResource)(nil)
	_ resource.ResourceWithImportState = (*DatabaseResource)(nil)
	_ resource.ResourceWithIdentity    = (*DatabaseResource)(nil)
)

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type DatabaseResource struct {
	api  *sdbSDK.API
	meta *meta.Meta
}

type databaseResourceModel struct {
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Endpoint           types.String   `tfsdk:"endpoint"`
	Version            types.String   `tfsdk:"version"`
	VersionEndOfLifeAt types.String   `tfsdk:"version_end_of_life_at"`
	SRN                types.String   `tfsdk:"srn"`
	Region             types.String   `tfsdk:"region"`
	ProjectID          types.String   `tfsdk:"project_id"`
	MaxCPU             types.Int64    `tfsdk:"max_cpu"`
	MinCPU             types.Int64    `tfsdk:"min_cpu"`
}

type databaseResourceIdentityModel = identityfw.RegionalIdentity

func (r *DatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sdb_sql_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: databaseResourceDescription,
		Description:         databaseResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the database",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The database name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_cpu": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(15),
				MarkdownDescription: "The maximum number of CPU units for your Serverless SQL Database",
			},
			"min_cpu": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "The minimum number of CPU units for your Serverless SQL Database",
			},
			"endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "endpoint of the database",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The major version of the PostgreSQL engine (e.g. `16`). If not specified, the default version is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version_end_of_life_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The end of life date of the database version (RFC3339).",
			},
			"srn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway Resource Name (SRN) of the database.",
			},
			"region": regional.SchemaAttributeComputed(),
			"project_id": account.ProjectIDSchemaFw(
				"The project_id you want to attach the resource to",
			),
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *DatabaseResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *DatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.meta = m
	r.api = sdbSDK.NewAPI(r.meta.ScwClient())
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	database, err := r.api.CreateDatabase(&sdbSDK.CreateDatabaseRequest{
		Region:       region,
		ProjectID:    projectID,
		Name:         plan.Name.ValueString(),
		CPUMin:       uint32(plan.MinCPU.ValueInt64()),
		CPUMax:       uint32(plan.MaxCPU.ValueInt64()),
		FromBackupID: nil,
		Version:      plan.Version.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Serverless SQL Database", err.Error())

		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	database, err = waitForDatabase(ctx, r.api, region, database.ID, timeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for Serverless SQL Database", err.Error())

		return
	}

	// Extra wait mirrors SDKv2 Create→Read (keeps VCR GET count aligned until cassettes are re-recorded).
	readTimeout, readDiags := plan.Timeouts.Read(ctx, defaultTimeout)
	resp.Diagnostics.Append(readDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	database, err = waitForDatabase(ctx, r.api, region, database.ID, readTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed reading Serverless SQL Database after create", err.Error())

		return
	}

	state := flattenDatabase(database)
	state.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(database.Region, database.ID))...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    databaseResourceModel
		identity databaseResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		id     string
		region scw.Region
		err    error
	)

	if identityAvailable {
		region, id = identityfw.ResolveRegionalIdentity(identity)
	} else {
		region, id, err = regional.ParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse Serverless SQL Database ID", err.Error())

			return
		}
	}

	timeout, diags := state.Timeouts.Read(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	database, err := waitForDatabase(ctx, r.api, region, id, timeout)
	if err != nil {
		// SDB returns 403 (not 404) for a database that no longer exists.
		if httperrors.Is404(err) || httperrors.Is403(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read Serverless SQL Database", err.Error())

		return
	}

	newState := flattenDatabase(database)
	newState.Timeouts = state.Timeouts

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(database.Region, database.ID))...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  databaseResourceModel
		state databaseResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Serverless SQL Database ID", err.Error())

		return
	}

	timeout, diags := plan.Timeouts.Update(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	database, err := waitForDatabase(ctx, r.api, region, id, timeout)
	if err != nil {
		// SDB returns 403 (not 404) for a database that no longer exists.
		if httperrors.Is404(err) || httperrors.Is403(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed waiting for Serverless SQL Database before update", err.Error())

		return
	}

	updateReq := &sdbSDK.UpdateDatabaseRequest{
		Region:     region,
		DatabaseID: database.ID,
	}
	hasChanges := false

	if !plan.MaxCPU.Equal(state.MaxCPU) {
		cpuMax := uint32(plan.MaxCPU.ValueInt64())
		updateReq.CPUMax = &cpuMax
		hasChanges = true
	}

	if !plan.MinCPU.Equal(state.MinCPU) {
		cpuMin := uint32(plan.MinCPU.ValueInt64())
		updateReq.CPUMin = &cpuMin
		hasChanges = true
	}

	if hasChanges {
		_, err = r.api.UpdateDatabase(updateReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update Serverless SQL Database", err.Error())

			return
		}
	}

	readTimeout, readDiags := plan.Timeouts.Read(ctx, defaultTimeout)
	resp.Diagnostics.Append(readDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	database, err = waitForDatabase(ctx, r.api, region, id, readTimeout)
	if err != nil {
		// SDB returns 403 (not 404) for a database that no longer exists.
		if httperrors.Is404(err) || httperrors.Is403(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed waiting for Serverless SQL Database after update", err.Error())

		return
	}

	newState := flattenDatabase(database)
	newState.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(database.Region, database.ID))...)
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var (
		state    databaseResourceModel
		identity databaseResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
	identityAvailable := !resp.Diagnostics.HasError() && !identity.ID.IsNull() && !identity.ID.IsUnknown()

	if !identityAvailable && resp.Diagnostics.HasError() {
		resp.Diagnostics = nil
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		id     string
		region scw.Region
		err    error
	)

	if identityAvailable {
		region, id = identityfw.ResolveRegionalIdentity(identity)
	} else {
		region, id, err = regional.ParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse Serverless SQL Database ID", err.Error())

			return
		}
	}

	timeout, diags := state.Timeouts.Delete(ctx, defaultTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = waitForDatabase(ctx, r.api, region, id, timeout)
	if err != nil {
		if httperrors.Is404(err) || httperrors.Is403(err) {
			return
		}

		resp.Diagnostics.AddError("Failed waiting for Serverless SQL Database before delete", err.Error())

		return
	}

	_, err = r.api.DeleteDatabase(&sdbSDK.DeleteDatabaseRequest{
		Region:     region,
		DatabaseID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) || httperrors.Is403(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete Serverless SQL Database", err.Error())

		return
	}

	_, err = waitForDatabase(ctx, r.api, region, id, timeout)
	if err != nil && !httperrors.Is403(err) && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed waiting for Serverless SQL Database deletion", err.Error())
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func flattenDatabase(database *sdbSDK.Database) databaseResourceModel {
	model := databaseResourceModel{
		ID:        types.StringValue(regional.NewIDString(database.Region, database.ID)),
		Name:      types.StringValue(database.Name),
		MaxCPU:    types.Int64Value(int64(database.CPUMax)),
		MinCPU:    types.Int64Value(int64(database.CPUMin)),
		Endpoint:  types.StringValue(database.Endpoint),
		Region:    types.StringValue(database.Region.String()),
		ProjectID: types.StringValue(database.ProjectID),
		SRN:       types.StringValue(database.Srn),
		Timeouts: timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"delete": types.StringType,
			"read":   types.StringType,
			"update": types.StringType,
		})},
	}

	if database.Version != nil {
		model.Version = types.StringValue(database.Version.Name)
		model.VersionEndOfLifeAt = types.StringValue(scwtypes.FlattenTime(database.Version.EndOfLifeAt).(string))
	} else {
		model.Version = types.StringNull()
		model.VersionEndOfLifeAt = types.StringValue("")
	}

	return model
}
