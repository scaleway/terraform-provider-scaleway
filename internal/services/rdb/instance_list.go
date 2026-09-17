package rdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*InstanceListResource)(nil)
	_ list.ListResourceWithConfigure    = (*InstanceListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*InstanceListResource)(nil)
)

type InstanceListResource struct {
	meta   *meta.Meta
	rdbAPI *rdbSDK.API
}

func (r *InstanceListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func NewInstanceListResource() list.ListResource {
	return &InstanceListResource{}
}

func (r *InstanceListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rdb_instance"
}

func (r *InstanceListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":             listscw.NameAttribute("Name of the RDB instance to filter on"),
			"tags":             listscw.TagsAttribute("Tags of the RDB instance to filter on"),
			"organization_id":  listscw.OrganizationIDAttribute("Organization ID of the RDB instance to filter on"),
			"project_ids":      listscw.ProjectIDsAttribute("Project IDs of the RDB instance to filter on"),
			"regions":          listscw.RegionsAttribute("Regions of the RDB instance to filter on"),
			"has_maintenances": schema.BoolAttribute{Optional: true, Description: "Whether to only list instances with scheduled maintenances"},
		},
	}
}

func (r *InstanceListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	resourceInstance := ResourceInstance()

	resp.ProtoV6Schema = translate.Schema(resourceInstance.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(resourceInstance.ProtoIdentitySchema(ctx)())
}

type InstanceListResourceModel struct {
	Tags            types.List   `tfsdk:"tags"`
	Name            types.String `tfsdk:"name"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	ProjectIDs      types.List   `tfsdk:"project_ids"`
	Regions         types.List   `tfsdk:"regions"`
	HasMaintenances types.Bool   `tfsdk:"has_maintenances"`
}

func (m *InstanceListResourceModel) GetTags() types.List {
	return m.Tags
}

func (m *InstanceListResourceModel) GetRegions() types.List {
	return m.Regions
}

func (m *InstanceListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (r *InstanceListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data InstanceListResourceModel

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

	var targets []listscw.RegionalFetchTarget

	for _, region := range regions {
		for _, project := range projects {
			targets = append(targets, listscw.RegionalFetchTarget{Region: region, ProjectID: project})
		}
	}

	allInstances, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*rdbSDK.Instance, error) {
			return r.fetchInstances(ctx, target, tags, data)
		},
		func(a, b *rdbSDK.Instance) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Region != b.Region {
				return strings.Compare(string(a.Region), string(b.Region))
			}

			return strings.Compare(a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing RDB instances", "Failed to list RDB instances: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, instance := range allInstances {
			result := req.NewListResult(ctx)
			result.DisplayName = instance.Name

			instanceResource := ResourceInstance()
			resourceData := instanceResource.Data(&terraform.InstanceState{})

			err := identity.SetRegionalIdentity(resourceData, instance.Region, instance.ID)
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

			diagsState := setInstanceState(ctx, resourceData, r.meta, r.rdbAPI, instance.Region, instance)
			if diagsState.HasError() {
				tflog.Error(ctx, "error from setting setInstanceState")

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

func (r *InstanceListResource) fetchInstances(ctx context.Context, target listscw.RegionalFetchTarget, tags []string, data InstanceListResourceModel) ([]*rdbSDK.Instance, error) {
	req := &rdbSDK.ListInstancesRequest{
		Region:         target.Region,
		Name:           data.Name.ValueStringPointer(),
		Tags:           tags,
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      &target.ProjectID,
	}

	if !data.HasMaintenances.IsNull() && !data.HasMaintenances.IsUnknown() {
		hasMaintenances := data.HasMaintenances.ValueBool()
		req.HasMaintenances = &hasMaintenances
	}

	response, err := r.rdbAPI.ListInstances(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	filtered := make([]*rdbSDK.Instance, 0, len(response.Instances))
	for _, instance := range response.Instances {
		if instance.ProjectID == target.ProjectID {
			filtered = append(filtered, instance)
		}
	}

	return filtered, nil
}
