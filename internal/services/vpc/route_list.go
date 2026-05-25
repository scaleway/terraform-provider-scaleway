package vpc

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	internaltypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ list.ListResource                 = (*RouteListResource)(nil)
	_ list.ListResourceWithConfigure    = (*RouteListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*RouteListResource)(nil)
)

type RouteListResource struct {
	meta      *meta.Meta
	routesAPI *vpc.RoutesWithNexthopAPI
}

func (r *RouteListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.routesAPI = vpc.NewRoutesWithNexthopAPI(meta.ExtractScwClient(m))
}

func NewRouteListResource() list.ListResource {
	return &RouteListResource{}
}

func (r *RouteListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_route"
}

func (r *RouteListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"regions": listscw.RegionsAttribute("Regions to filter for."),
			"tags":    listscw.TagsAttribute("Tags of the VPC route to filter for"),
			"vpc_id": schema.StringAttribute{
				Description: "Filter for routes belonging to this VPC (regional ID or plain UUID).",
				Optional:    true,
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
			"nexthop_resource_id": schema.StringAttribute{
				Description: "Filter for routes with this nexthop resource ID (regional ID or plain UUID).",
				Optional:    true,
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
			"nexthop_private_network_id": schema.StringAttribute{
				Description: "Filter for routes with this nexthop private network ID (regional ID or plain UUID).",
				Optional:    true,
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
			"nexthop_vpc_connector_id": schema.StringAttribute{
				Description: "Filter for routes with this nexthop VPC connector ID (regional ID or plain UUID).",
				Optional:    true,
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
			"nexthop_resource_type": schema.StringAttribute{
				Description: "Filter for routes with this nexthop resource type.",
				Optional:    true,
				Validators: []validator.String{
					verify.ValidateEnumFramework[vpc.RouteWithNexthopResourceType](),
				},
			},
			"is_ipv6": schema.BoolAttribute{
				Description: "Filter for routes with an IPv6 destination.",
				Optional:    true,
			},
			"contains": schema.StringAttribute{
				Description: "Filter for routes whose destination is contained in this subnet (CIDR notation).",
				Optional:    true,
			},
		},
	}
}

func (r *RouteListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	routeResource := ResourceRoute()

	resp.ProtoV6Schema = translate.Schema(routeResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(routeResource.ProtoIdentitySchema(ctx)())
}

type RouteListResourceModel struct {
	Tags                    types.List   `tfsdk:"tags"`
	Regions                 types.List   `tfsdk:"regions"`
	VpcID                   types.String `tfsdk:"vpc_id"`
	NexthopResourceID       types.String `tfsdk:"nexthop_resource_id"`
	NexthopPrivateNetworkID types.String `tfsdk:"nexthop_private_network_id"`
	NexthopVpcConnectorID   types.String `tfsdk:"nexthop_vpc_connector_id"`
	NexthopResourceType     types.String `tfsdk:"nexthop_resource_type"`
	Contains                types.String `tfsdk:"contains"`
	IsIPv6                  types.Bool   `tfsdk:"is_ipv6"`
}

func (m *RouteListResourceModel) GetTags() types.List    { return m.Tags }
func (m *RouteListResourceModel) GetRegions() types.List { return m.Regions }

func (r *RouteListResource) FetchRoutes(ctx context.Context, region scw.Region, tags []string, data RouteListResourceModel) ([]*vpc.Route, error) {
	req := &vpc.RoutesWithNexthopAPIListRoutesWithNexthopRequest{
		Region:                  region,
		Tags:                    tags,
		VpcID:                   locality.ExpandFrameworkID(data.VpcID),
		NexthopResourceID:       locality.ExpandFrameworkID(data.NexthopResourceID),
		NexthopPrivateNetworkID: locality.ExpandFrameworkID(data.NexthopPrivateNetworkID),
		NexthopVpcConnectorID:   locality.ExpandFrameworkID(data.NexthopVpcConnectorID),
	}

	if !data.NexthopResourceType.IsNull() && !data.NexthopResourceType.IsUnknown() {
		req.NexthopResourceType = vpc.RouteWithNexthopResourceType(data.NexthopResourceType.ValueString())
	}

	if !data.IsIPv6.IsNull() && !data.IsIPv6.IsUnknown() {
		req.IsIPv6 = data.IsIPv6.ValueBoolPointer()
	}

	if !data.Contains.IsNull() && !data.Contains.IsUnknown() {
		ipNet, err := internaltypes.ExpandIPNet(data.Contains.ValueString())
		if err != nil {
			return nil, err
		}

		req.Contains = &ipNet
	}

	response, err := r.routesAPI.ListRoutesWithNexthop(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	var routes []*vpc.Route

	for _, rn := range response.Routes {
		if rn.Route != nil && rn.Route.ID != "" {
			routes = append(routes, rn.Route)
		}
	}

	return routes, nil
}

func (r *RouteListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data RouteListResourceModel

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

	allRoutes, err := listscw.FetchConcurrently(ctx, regions,
		func(ctx context.Context, region scw.Region) ([]*vpc.Route, error) {
			return r.FetchRoutes(ctx, region, tags, data)
		},
		func(a, b *vpc.Route) int {
			if a.Region != b.Region {
				return strings.Compare(string(a.Region), string(b.Region))
			}

			return strings.Compare(a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing VPC Routes", "Failed to list VPC Routes: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, route := range allRoutes {
			result := req.NewListResult(ctx)
			result.DisplayName = route.ID

			routeResource := ResourceRoute()
			resourceData := routeResource.Data(&terraform.InstanceState{})

			err = identity.SetRegionalIdentity(resourceData, route.Region, route.ID)
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

			sdkDiags := setRouteState(resourceData, route)
			if sdkDiags.HasError() {
				tflog.Error(ctx, "error from setting VPC route state")

				for _, d := range sdkDiags {
					result.Diagnostics.AddError(d.Summary, d.Detail)
				}

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
