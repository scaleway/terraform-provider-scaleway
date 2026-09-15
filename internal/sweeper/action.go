package sweeper

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*SweepResourcesAction)(nil)
	_ action.ActionWithConfigure = (*SweepResourcesAction)(nil)
)

//go:embed descriptions/sweep_resources_action.md
var sweepResourcesActionDescription string

type SweepResourcesAction struct {
	meta *meta.Meta
}

func NewSweepResourcesAction() action.Action {
	return &SweepResourcesAction{}
}

func (a *SweepResourcesAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sweep_resources"
}

func (a *SweepResourcesAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.meta = m
}

// sweepResourcesActionModel is the config model for the action. It implements
// the listscw RegionalModel, ProjectModel and TagsModel interfaces so the
// shared extraction helpers can be reused.
type sweepResourcesActionModel struct {
	ResourceType   types.String `tfsdk:"resource_type"`
	Tags           types.List   `tfsdk:"tags"`
	Regions        types.List   `tfsdk:"regions"`
	ProjectIDs     types.List   `tfsdk:"project_ids"`
	OrganizationID types.String `tfsdk:"organization_id"`
	DryRun         types.Bool   `tfsdk:"dry_run"`
}

func (m *sweepResourcesActionModel) GetRegions() types.List  { return m.Regions }
func (m *sweepResourcesActionModel) GetProjects() types.List { return m.ProjectIDs }
func (m *sweepResourcesActionModel) GetTags() types.List     { return m.Tags }

func (a *SweepResourcesAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         sweepResourcesActionDescription,
		MarkdownDescription: sweepResourcesActionDescription,
		Attributes: map[string]schema.Attribute{
			"resource_type": schema.StringAttribute{
				Required:    true,
				Description: "The Terraform resource type to sweep (e.g. scaleway_iam_api_key).",
				Validators: []validator.String{
					stringvalidator.OneOf(SupportedTypes()...),
				},
			},
			"tags": schema.ListAttribute{
				Description: "List of Scaleway tags to filter on. Only resources carrying at least one of the given tags are swept. Use this for taggable types (e.g. scaleway_secret, scaleway_key_manager_key).",
				ElementType: types.StringType,
				Optional:    true,
			},
			"regions": schema.ListAttribute{
				Description: "Regions to scan. Use '*' to scan all regions. Only used for regional resource types.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(append(regional.AllRegions(), "*")...),
					),
				},
			},
			"project_ids": schema.ListAttribute{
				Description: "Project IDs to scope the sweep. Use '*' to sweep across all projects. Only used for regional resource types.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.Any(
							stringvalidator.OneOf("*"),
							verify.IsStringUUID(),
						),
					),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "Organization ID to scope the sweep. Only used for global resource types. If omitted, the provider default organization is used.",
				Optional:    true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"dry_run": schema.BoolAttribute{
				Description: "When true (the default), list the resources that would be deleted without actually deleting them. Set to false to perform the deletion.",
				Optional:    true,
			},
		},
	}
}

func (a *SweepResourcesAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data sweepResourcesActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.meta == nil {
		resp.Diagnostics.AddError(
			"Unconfigured sweeper action",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	sweeper, ok := FindSweeper(data.ResourceType.ValueString())
	if !ok {
		resp.Diagnostics.AddError(
			"Unsupported resource type",
			fmt.Sprintf("The resource type %q is not supported by the sweeper action. Supported types: %s.",
				data.ResourceType.ValueString(), strings.Join(SupportedTypes(), ", ")),
		)

		return
	}

	// dry_run defaults to true when unset, so the action never deletes by accident.
	dryRun := true
	if !data.DryRun.IsNull() && !data.DryRun.IsUnknown() {
		dryRun = data.DryRun.ValueBool()
	}

	tags, diags := listscw.ExtractTags(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(tags) > 0 && !sweeper.SupportsTagFilter() {
		resp.Diagnostics.Append(diag.NewWarningDiagnostic(
			"Tags ignored for this resource type",
			fmt.Sprintf("The resource type %q does not support tag filtering. The provided tags will be ignored and all matching resources will be swept.",
				sweeper.TypeName()),
		))
	}

	filter := Filter{
		Tags: tags,
	}

	switch sweeper.Locality() {
	case LocalityGlobal:
		if !data.Regions.IsNull() {
			resp.Diagnostics.Append(diag.NewWarningDiagnostic(
				"Regions ignored for global resource type",
				fmt.Sprintf("The resource type %q is global; the regions argument is ignored.", sweeper.TypeName()),
			))
		}

		if !data.ProjectIDs.IsNull() {
			resp.Diagnostics.Append(diag.NewWarningDiagnostic(
				"Project IDs ignored for global resource type",
				fmt.Sprintf("The resource type %q is global; the project_ids argument is ignored.", sweeper.TypeName()),
			))
		}

		filter.OrganizationID = data.OrganizationID.ValueString()
	case LocalityRegional:
		regions, err := listscw.ExtractRegions(ctx, &data, a.meta)
		if err != nil {
			resp.Diagnostics.AddError("Failed to extract regions", err.Error())

			return
		}

		filter.Regions = regions

		projects, err := listscw.ExtractProjects(ctx, &data, a.meta)
		if err != nil {
			resp.Diagnostics.AddError("Failed to extract projects", err.Error())

			return
		}

		filter.ProjectIDs = projects
	}

	resources, err := sweeper.List(ctx, a.meta, filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list resources",
			fmt.Sprintf("Failed to list %s resources: %s", sweeper.TypeName(), err),
		)

		return
	}

	prefix := "would delete"
	if !dryRun {
		prefix = "deleted"
	}

	var deletionErrors []string

	for _, resource := range resources {
		if dryRun {
			a.sendProgress(resp, fmt.Sprintf("[dry-run] %s %s %q (%s)", prefix, sweeper.TypeName(), resource.ID, resource.DisplayName))

			continue
		}

		if err := sweeper.Delete(ctx, a.meta, resource); err != nil {
			deletionErrors = append(deletionErrors, fmt.Sprintf("%s (%s): %s", resource.ID, resource.DisplayName, err))
			a.sendProgress(resp, fmt.Sprintf("failed to delete %s %q (%s): %s", sweeper.TypeName(), resource.ID, resource.DisplayName, err))
		} else {
			a.sendProgress(resp, fmt.Sprintf("%s %s %q (%s)", prefix, sweeper.TypeName(), resource.ID, resource.DisplayName))
		}
	}

	summary := fmt.Sprintf("%s: %d resource(s) %s", sweeper.TypeName(), len(resources), prefix)
	if dryRun {
		summary += " (dry-run)"
	}

	if len(deletionErrors) > 0 {
		resp.Diagnostics.Append(diag.NewWarningDiagnostic(
			"Sweep completed with errors",
			fmt.Sprintf("%s\n\nThe following deletions failed:\n  - %s", summary, strings.Join(deletionErrors, "\n  - ")),
		))
	} else {
		resp.Diagnostics.Append(diag.NewWarningDiagnostic("Sweep summary", summary))
	}
}

func (a *SweepResourcesAction) sendProgress(resp *action.InvokeResponse, message string) {
	if resp.SendProgress == nil {
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{Message: message})
}

// FindSweeper returns the sweeper registered for the given Terraform type name.
func FindSweeper(typeName string) (Sweeper, bool) {
	for _, s := range All() {
		if s.TypeName() == typeName {
			return s, true
		}
	}

	return nil, false
}
