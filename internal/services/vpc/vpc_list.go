package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                   = (*ListResource)(nil)
	_ list.ListResourceWithConfigure      = (*ListResource)(nil)
	_ list.ListResourceWithRawV6Schemas   = (*ListResource)(nil)
	_ list.ListResourceWithValidateConfig = (*ListResource)(nil)
)

type ListResource struct {
	vpcAPI *vpc.API
}

func (r *ListResource) ValidateListResourceConfig(ctx context.Context, request list.ValidateConfigRequest, response *list.ValidateConfigResponse) {
}

func (r *ListResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	m, ok := request.ProviderData.(*meta.Meta)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected List Configure Type",
			fmt.Sprintf("Expected *scw.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	client := m.ScwClient()
	r.vpcAPI = vpc.NewAPI(client)
}

func NewVPCListResource() list.ListResource {
	return &ListResource{}
}

func (r *ListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name of the vpc to list for",
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Organization ID of the VPC to list for",
				Optional:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "Project ID of the VPC to list for",
				Optional:    true,
			},
			"routing_enabled": schema.BoolAttribute{
				Description: "Whether routing is enabled for VPC",
				Optional:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with VPC",
				ElementType: types.StringType,
				Optional:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether the VPC is the default VPC",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region of the VPC. Use 'all' to list VPCs from all regions",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(append(regional.AllRegions(), "all")...),
				},
			},
		},
	}
}

func (r *ListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	resourceVPC := ResourceVPC()

	resp.ProtoV6Schema = translate.Schema(resourceVPC.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(resourceVPC.ProtoIdentitySchema(ctx)())
}

type ListResourceModel struct {
	Tags           types.List   `tfsdk:"tags"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Region         types.String `tfsdk:"region"`
	RoutingEnabled types.Bool   `tfsdk:"routing_enabled"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
}

func (r *ListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *ListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ListResourceModel

	// Read list config data into the model
	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	// Determine regions to query
	var regionsToQuery []scw.Region
	if data.Region.ValueString() == "all" {
		regionsToQuery = scw.AllRegions
	} else {
		regionsToQuery = []scw.Region{scw.Region(data.Region.ValueString())}
	}

	// Convert tags from types.List to []string
	var tags []string
	if !data.Tags.IsNull() {
		diags = data.Tags.ElementsAs(ctx, &tags, false)
		if diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, region := range regionsToQuery {
			listRequest := &vpc.ListVPCsRequest{
				Region:         region,
				Name:           data.Name.ValueStringPointer(),
				Tags:           tags,
				OrganizationID: data.OrganizationID.ValueStringPointer(),
				ProjectID:      data.ProjectID.ValueStringPointer(),
				IsDefault:      data.IsDefault.ValueBoolPointer(),
				RoutingEnabled: data.RoutingEnabled.ValueBoolPointer(),
			}

			vpcs, err := r.vpcAPI.ListVPCs(listRequest, scw.WithContext(ctx))
			if err != nil {
				diags.AddError("Error listing VPCs", fmt.Sprintf("Failed to list VPCs in region %s: %s", region, err.Error()))
				stream.Results = list.ListResultsStreamDiagnostics(diags)

				return
			}

			for _, rawVPC := range vpcs.Vpcs {
				result := req.NewListResult(ctx)
				result.DisplayName = rawVPC.Name

				vpcResource := ResourceVPC()
				resourceData := vpcResource.Data(&terraform.InstanceState{})
				err = identity.SetRegionalIdentity(resourceData, region, rawVPC.ID)
				if err != nil {
					result.Diagnostics.AddError(
						"Retrieving identity data",
						"An error was encountered when retrieving the identity data: "+err.Error(),
					)

					return
				}

				// Convert and set the identity and resource state into the result
				tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
				if errIdentityState != nil {
					result.Diagnostics.AddError(
						"Converting identity data",
						"An error was encountered when converting the identity data: "+err.Error(),
					)
				}

				errtfTypeIdentity := result.Identity.Set(ctx, *tfTypeIdentity)
				if errtfTypeIdentity != nil {
					result.Diagnostics.AddError(
						"Setting identity data",
						"An error was encountered when setting the identity data: "+err.Error(),
					)
				}

				diagsState := setVPCState(resourceData, rawVPC)
				if diagsState.HasError() {
					tflog.Error(ctx, "error from setting setVPCState")
					return
				}

				// Convert and set the resource state into the result
				tfTypeResource, errTfTypeResourceState := resourceData.TfTypeResourceState()
				if errTfTypeResourceState != nil {
					result.Diagnostics.AddError(
						"Converting resource state",
						"An error was encountered when converting the resource state: "+err.Error(),
					)
				}

				errtfTypeResource := result.Resource.Set(ctx, *tfTypeResource)
				if errtfTypeResource != nil {
					result.Diagnostics.AddError(
						"Setting resource state",
						"An error was encountered when setting the resource state: "+err.Error(),
					)
				}

				// Send the result to the stream.
				if !push(result) {
					return
				}
			}
		}
	}
}
