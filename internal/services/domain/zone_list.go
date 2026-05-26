package domain

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server/translate"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*ZoneListResource)(nil)
	_ list.ListResourceWithConfigure    = (*ZoneListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*ZoneListResource)(nil)
)

type ZoneListResource struct {
	meta      *meta.Meta
	domainAPI *domainSDK.API
}

func NewZoneListResource() list.ListResource {
	return &ZoneListResource{}
}

func (r *ZoneListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.domainAPI = NewDomainAPI(m)
}

func (r *ZoneListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_zone"
}

func (r *ZoneListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_ids": listscw.ProjectIDsAttribute("Project IDs to filter DNS zones on"),
			"domains": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Domain apex names to list DNS zones for. Use [\"*\"] to list zones across all domains. Must contain at least one value.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.Any(
							stringvalidator.OneOf("*"),
							stringvalidator.LengthAtLeast(1),
						),
					),
				},
			},
			"dns_zones": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Filter by DNS zone FQDNs (for example subdomain.example.com or example.com for a root zone).",
			},
			"created_after":  optionalRFC3339Attribute("Only list DNS zones created after this date (RFC3339)."),
			"created_before": optionalRFC3339Attribute("Only list DNS zones created before this date (RFC3339)."),
			"updated_after":  optionalRFC3339Attribute("Only list DNS zones updated after this date (RFC3339)."),
			"updated_before": optionalRFC3339Attribute("Only list DNS zones updated before this date (RFC3339)."),
		},
	}
}

func optionalRFC3339Attribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Description: description,
		Optional:    true,
	}
}

func (r *ZoneListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	zoneResource := ResourceZone()

	resp.ProtoV6Schema = translate.Schema(zoneResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(zoneResource.ProtoIdentitySchema(ctx)())
}

type ZoneListResourceModel struct {
	ProjectIDs    types.List   `tfsdk:"project_ids"`
	Domains       types.List   `tfsdk:"domains"`
	DNSZones      types.List   `tfsdk:"dns_zones"`
	CreatedAfter  types.String `tfsdk:"created_after"`
	CreatedBefore types.String `tfsdk:"created_before"`
	UpdatedAfter  types.String `tfsdk:"updated_after"`
	UpdatedBefore types.String `tfsdk:"updated_before"`
}

func (m *ZoneListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

type zoneListTarget struct {
	ProjectID string
	Domain    string
}

func (r *ZoneListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data ZoneListResourceModel

	diags := req.Config.Get(ctx, &data)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	projects, err := listscw.ExtractProjects(ctx, &data, r.meta)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing projects", "Failed to list projects: "+err.Error()),
		})

		return
	}

	var domainElems []string

	diags = data.Domains.ElementsAs(ctx, &domainElems, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	if len(domainElems) == 0 {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid domains", "`domains` must contain at least one element."),
		})

		return
	}

	if slices.Contains(domainElems, "*") && len(domainElems) != 1 {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid domains", `When using "*", domains must be exactly ["*"].`),
		})

		return
	}

	timeFilters, timeDiags := parseZoneListTimeFilters(ctx, &data)
	if timeDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(timeDiags)

		return
	}

	var dnsZoneElems []string

	if !data.DNSZones.IsNull() && !data.DNSZones.IsUnknown() {
		diags = data.DNSZones.ElementsAs(ctx, &dnsZoneElems, true)
		if diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)

			return
		}
	}

	targets := buildZoneListTargets(projects, domainElems)

	allZones, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target zoneListTarget) ([]*domainSDK.DNSZone, error) {
			return r.fetchDNSZones(ctx, target, dnsZoneElems, timeFilters)
		},
		func(a, b *domainSDK.DNSZone) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.Domain != b.Domain {
				return strings.Compare(a.Domain, b.Domain)
			}

			return strings.Compare(a.Subdomain, b.Subdomain)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing domain zones", "Failed to list domain zones: "+err.Error()),
		})

		return
	}

	allZones = dedupeDNSZones(allZones)

	stream.Results = func(push func(list.ListResult) bool) {
		for _, zone := range allZones {
			result := req.NewListResult(ctx)
			result.DisplayName = BuildZoneName(zone.Subdomain, zone.Domain)

			zoneResource := ResourceZone()
			resourceData := zoneResource.Data(&terraform.InstanceState{})

			zoneName := BuildZoneName(zone.Subdomain, zone.Domain)
			if err := identity.SetGlobalIdentity(resourceData, zoneName); err != nil {
				result.Diagnostics.AddError("Retrieving identity data",
					"An error was encountered when retrieving the identity data: "+err.Error())

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

			setZoneState(resourceData, zone)

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

type zoneListTimeFilters struct {
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
}

func parseZoneListTimeFilters(_ context.Context, data *ZoneListResourceModel) (zoneListTimeFilters, diag.Diagnostics) {
	var filters zoneListTimeFilters

	var diags diag.Diagnostics

	createdAfter, d := parseOptionalRFC3339(data.CreatedAfter, "created_after")
	diags.Append(d...)

	filters.CreatedAfter = createdAfter

	createdBefore, d := parseOptionalRFC3339(data.CreatedBefore, "created_before")
	diags.Append(d...)

	filters.CreatedBefore = createdBefore

	updatedAfter, d := parseOptionalRFC3339(data.UpdatedAfter, "updated_after")
	diags.Append(d...)

	filters.UpdatedAfter = updatedAfter

	updatedBefore, d := parseOptionalRFC3339(data.UpdatedBefore, "updated_before")
	diags.Append(d...)

	filters.UpdatedBefore = updatedBefore

	return filters, diags
}

func parseOptionalRFC3339(value types.String, name string) (*time.Time, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value.ValueString())
	if err != nil {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid "+name,
				fmt.Sprintf("%q must be a valid RFC3339 timestamp: %s", name, err.Error()),
			),
		}
	}

	return &parsed, nil
}

func buildZoneListTargets(projects, domains []string) []zoneListTarget {
	if slices.Contains(domains, "*") {
		targets := make([]zoneListTarget, 0, len(projects))
		for _, projectID := range projects {
			targets = append(targets, zoneListTarget{ProjectID: projectID})
		}

		return targets
	}

	targets := make([]zoneListTarget, 0, len(projects)*len(domains))
	for _, projectID := range projects {
		for _, domainName := range domains {
			targets = append(targets, zoneListTarget{
				ProjectID: projectID,
				Domain:    strings.ToLower(domainName),
			})
		}
	}

	return targets
}

func (r *ZoneListResource) fetchDNSZones(
	ctx context.Context,
	target zoneListTarget,
	dnsZones []string,
	timeFilters zoneListTimeFilters,
) ([]*domainSDK.DNSZone, error) {
	request := &domainSDK.ListDNSZonesRequest{
		ProjectID:     &target.ProjectID,
		Domain:        target.Domain,
		DNSZones:      dnsZones,
		CreatedAfter:  timeFilters.CreatedAfter,
		CreatedBefore: timeFilters.CreatedBefore,
		UpdatedAfter:  timeFilters.UpdatedAfter,
		UpdatedBefore: timeFilters.UpdatedBefore,
	}

	response, err := r.domainAPI.ListDNSZones(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	return response.DNSZones, nil
}

func dedupeDNSZones(zones []*domainSDK.DNSZone) []*domainSDK.DNSZone {
	seen := make(map[string]struct{}, len(zones))
	result := make([]*domainSDK.DNSZone, 0, len(zones))

	for _, zone := range zones {
		key := zone.ProjectID + "/" + BuildZoneName(zone.Subdomain, zone.Domain)
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		result = append(result, zone)
	}

	return result
}

func setZoneState(d *sdkschema.ResourceData, zone *domainSDK.DNSZone) {
	_ = d.Set("subdomain", zone.Subdomain)
	_ = d.Set("domain", zone.Domain)
	_ = d.Set("ns", zone.Ns)
	_ = d.Set("ns_default", zone.NsDefault)
	_ = d.Set("ns_master", zone.NsMaster)
	_ = d.Set("status", zone.Status.String())
	_ = d.Set("message", zone.Message)
	_ = d.Set("updated_at", zone.UpdatedAt.String())
	_ = d.Set("project_id", zone.ProjectID)
}
