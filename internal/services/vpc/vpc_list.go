package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	terraformSDKv2 "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                     = (*ListResource)(nil)
	_ list.ListResourceWithConfigure        = (*ListResource)(nil)
	_ list.ListResourceWithRawV6Schemas     = (*ListResource)(nil)
	_ list.ListResourceWithConfigValidators = (*ListResource)(nil)
	_ list.ListResourceWithValidateConfig   = (*ListResource)(nil)
)

type ListResource struct {
	vpcAPI *vpc.API
}

func (r *ListResource) ValidateListResourceConfig(ctx context.Context, request list.ValidateConfigRequest, response *list.ValidateConfigResponse) {
}

func (r *ListResource) ListResourceConfigValidators(ctx context.Context) []list.ConfigValidator {
	return []list.ConfigValidator{}
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
				Description: "Region of the VPC",
				Optional:    true,
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

	listRequest := &vpc.ListVPCsRequest{
		Region: scw.Region(data.Region.ValueString()),
		Name:   data.Name.ValueStringPointer(),
		// Tags:           data.Tags.Elements(),
		OrganizationID: data.OrganizationID.ValueStringPointer(),
		ProjectID:      data.ProjectID.ValueStringPointer(),
		IsDefault:      data.IsDefault.ValueBoolPointer(),
		RoutingEnabled: data.RoutingEnabled.ValueBoolPointer(),
	}

	vpcs, err := r.vpcAPI.ListVPCs(listRequest, scw.WithContext(ctx))
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, rawVPC := range vpcs.Vpcs {
			result := req.NewListResult(ctx)
			result.DisplayName = rawVPC.Name

			v := ResourceVPC()
			d := v.Data(&terraformSDKv2.InstanceState{})
			setVPCState(d, rawVPC)

			err := identity.SetRegionalIdentity(d, rawVPC.Region, rawVPC.ID)
			if err != nil {
				return
			}

			// Send the result to the stream.
			if !push(result) {
				return
			}
		}
	}
}
