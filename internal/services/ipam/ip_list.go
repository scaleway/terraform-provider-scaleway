package ipam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ipamSDK "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ list.ListResource                 = (*IPListResource)(nil)
	_ list.ListResourceWithConfigure    = (*IPListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*IPListResource)(nil)
)

type IPListResource struct {
	meta    *meta.Meta
	ipamAPI *ipamSDK.API
}

func (r *IPListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.ipamAPI = ipamSDK.NewAPI(meta.ExtractScwClient(m))
}

func NewIPListResource() list.ListResource {
	return &IPListResource{}
}

func (r *IPListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attached": schema.BoolAttribute{
				Description: "Filter for IPs which are attached to a resource",
				Optional:    true,
			},
			"is_ipv6": schema.BoolAttribute{
				Description: "Filter only for IPv6 addresses",
				Optional:    true,
			},
			"vpc_id": schema.StringAttribute{
				Description: "VPC ID to filter for. Only IPs owned by resources in this VPC will be returned",
				Optional:    true,
			},
			"private_network_id": schema.StringAttribute{
				Description: "Private Network ID to filter for. Only IPs in this Private Network will be returned. Mutually exclusive with zonal, subnet_id, source_vpc_id",
				Optional:    true,
				Validators:  verify.MutuallyExclusiveStringConflicts("private_network_id", sourceFilterGroup...),
			},
			"subnet_id": schema.StringAttribute{
				Description: "Subnet ID to filter for. Only IPs inside this exact subnet will be returned. Mutually exclusive with zonal, private_network_id, source_vpc_id",
				Optional:    true,
				Validators:  verify.MutuallyExclusiveStringConflicts("subnet_id", sourceFilterGroup...),
			},
			"zonal": schema.StringAttribute{
				Description: "Zone to filter for. Only zonal IPs in this zone will be returned. Mutually exclusive with private_network_id, subnet_id, source_vpc_id",
				Optional:    true,
				Validators:  verify.MutuallyExclusiveStringConflicts("zonal", sourceFilterGroup...),
			},
			"source_vpc_id": schema.StringAttribute{
				Description: "Source VPC ID to filter for. Mutually exclusive with zonal, private_network_id, subnet_id",
				Optional:    true,
				Validators:  verify.MutuallyExclusiveStringConflicts("source_vpc_id", sourceFilterGroup...),
			},
			"resource_name": schema.StringAttribute{
				Description: "Attached resource name to filter for",
				Optional:    true,
			},
			"resource_type": schema.StringAttribute{
				Description: "Resource type to filter for (e.g. instance_server, lb_server, etc.)",
				Optional:    true,
			},
			"resource_types": schema.ListAttribute{
				Description: "Resource types to filter for",
				Optional:    true,
				ElementType: types.StringType,
			},
			"resource_ids": schema.ListAttribute{
				Description: "Resource IDs to filter for",
				Optional:    true,
				ElementType: types.StringType,
			},
			"mac_address": schema.StringAttribute{
				Description: "MAC address to filter for",
				Optional:    true,
			},
			"name":            listscw.NameAttribute("Name filter (unused for IPAM but kept for consistency)"),
			"tags":            listscw.TagsAttribute("Tags to filter for"),
			"organization_id": listscw.OrganizationIDAttribute("Organization ID to filter for"),
			"project_ids":     listscw.ProjectIDsAttribute("Project IDs to filter for."),
			"regions":         listscw.RegionsAttribute("Regions to filter for."),
		},
	}
}

func (r *IPListResource) RawV6Schemas(ctx context.Context, req list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	ipResource := ResourceIP()

	resp.ProtoV6Schema = translate.Schema(ipResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(ipResource.ProtoIdentitySchema(ctx)())
}

type IPListResourceModel struct {
	Tags             types.List   `tfsdk:"tags"`
	ProjectIDs       types.List   `tfsdk:"project_ids"`
	Regions          types.List   `tfsdk:"regions"`
	ResourceTypes    types.List   `tfsdk:"resource_types"`
	ResourceIDs      types.List   `tfsdk:"resource_ids"`
	Name             types.String `tfsdk:"name"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	VpcID            types.String `tfsdk:"vpc_id"`
	PrivateNetworkID types.String `tfsdk:"private_network_id"`
	SubnetID         types.String `tfsdk:"subnet_id"`
	Zonal            types.String `tfsdk:"zonal"`
	SourceVpcID      types.String `tfsdk:"source_vpc_id"`
	ResourceName     types.String `tfsdk:"resource_name"`
	ResourceType     types.String `tfsdk:"resource_type"`
	MacAddress       types.String `tfsdk:"mac_address"`
	Attached         types.Bool   `tfsdk:"attached"`
	IsIPv6           types.Bool   `tfsdk:"is_ipv6"`
}

func (m *IPListResourceModel) GetTags() types.List    { return m.Tags }
func (m *IPListResourceModel) GetRegions() types.List { return m.Regions }
func (m *IPListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

func (r *IPListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipam_ip"
}

var sourceFilterGroup = []string{"zonal", "private_network_id", "subnet_id", "source_vpc_id"}

func (r *IPListResource) FetchIPs(ctx context.Context, region scw.Region, project *string, tags []string, data IPListResourceModel) ([]*ipamSDK.IP, error) {
	req := &ipamSDK.ListIPsRequest{
		Region:           region,
		Tags:             tags,
		OrganizationID:   data.OrganizationID.ValueStringPointer(),
		ProjectID:        project,
		Attached:         data.Attached.ValueBoolPointer(),
		IsIPv6:           data.IsIPv6.ValueBoolPointer(),
		VpcID:            locality.ExpandFrameworkID(data.VpcID),
		PrivateNetworkID: locality.ExpandFrameworkID(data.PrivateNetworkID),
		SubnetID:         locality.ExpandFrameworkID(data.SubnetID),
		SourceVpcID:      locality.ExpandFrameworkID(data.SourceVpcID),
		Zonal:            data.Zonal.ValueStringPointer(),
		ResourceName:     data.ResourceName.ValueStringPointer(),
		MacAddress:       data.MacAddress.ValueStringPointer(),
	}

	if !data.ResourceType.IsNull() {
		req.ResourceType = ipamSDK.ResourceType(data.ResourceType.ValueString())
	}

	if !data.ResourceTypes.IsNull() {
		var resourceTypeStrings []string

		diags := data.ResourceTypes.ElementsAs(ctx, &resourceTypeStrings, false)
		if diags.HasError() {
			return nil, fmt.Errorf("converting resource_types: %s", diags.Errors()[0].Detail())
		}

		resourceTypes := make([]ipamSDK.ResourceType, len(resourceTypeStrings))
		for i, rt := range resourceTypeStrings {
			resourceTypes[i] = ipamSDK.ResourceType(rt)
		}

		req.ResourceTypes = resourceTypes
	}

	if !data.ResourceIDs.IsNull() {
		resourceIDs, diags := locality.ExpandFrameworkIDs(ctx, data.ResourceIDs)
		if diags.HasError() {
			return nil, fmt.Errorf("converting resource_ids: %s", diags.Errors()[0].Detail())
		}

		req.ResourceIDs = resourceIDs
	}

	response, err := r.ipamAPI.ListIPs(req, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.IPs, nil
}

func (r *IPListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data IPListResourceModel

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

	allIPs, err := listscw.FetchConcurrently(ctx, listscw.RegionalProjectTargets(regions, projects),
		func(ctx context.Context, target listscw.RegionalFetchTarget) ([]*ipamSDK.IP, error) {
			return r.FetchIPs(ctx, target.Region, &target.ProjectID, tags, data)
		},
		func(a, b *ipamSDK.IP) int {
			return listscw.CompareRegionalProjectItems(a.ProjectID, b.ProjectID, a.Region, b.Region, a.ID, b.ID)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing IPAM IPs", "Failed to list IPAM IPs: "+err.Error()),
		})

		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, ip := range allIPs {
			result := req.NewListResult(ctx)

			addressCidr, flattenErr := scwtypes.FlattenIPNet(ip.Address)
			if flattenErr != nil {
				result.DisplayName = ip.ID
			} else {
				result.DisplayName = addressCidr
			}

			ipResource := ResourceIP()
			resourceData := ipResource.Data(&terraform.InstanceState{})

			err = identity.SetRegionalIdentity(resourceData, ip.Region, ip.ID)
			if err != nil {
				result.Diagnostics.AddError("Retrieving identity data", "An error was encountered when retrieving the identity data: "+err.Error())

				if !push(result) {
					return
				}

				continue
			}

			tfTypeIdentity, errIdentityState := resourceData.TfTypeIdentityState()
			if errIdentityState != nil {
				result.Diagnostics.AddError("Converting identity data", "An error was encountered when converting the identity data: "+errIdentityState.Error())
			}

			identitySetDiags := result.Identity.Set(ctx, *tfTypeIdentity)
			result.Diagnostics.Append(identitySetDiags...)

			privateNetworkID := ""
			if ip.Source != nil && ip.Source.PrivateNetworkID != nil {
				privateNetworkID = regional.NewIDString(ip.Region, *ip.Source.PrivateNetworkID)
			}

			sdkDiags := setIPAMIPState(resourceData, ip, privateNetworkID)
			if sdkDiags.HasError() {
				tflog.Error(ctx, "error from setting IPAM IP state")

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
				result.Diagnostics.AddError("Converting resource state", "An error was encountered when converting the resource state: "+errTfTypeResourceState.Error())
			}

			resourceSetDiags := result.Resource.Set(ctx, *tfTypeResource)
			result.Diagnostics.Append(resourceSetDiags...)

			if !push(result) {
				return
			}
		}
	}
}
