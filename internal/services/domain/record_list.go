package domain

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
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	domainSDK "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	listscw "github.com/scaleway/terraform-provider-scaleway/v2/internal/list"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ list.ListResource                 = (*RecordListResource)(nil)
	_ list.ListResourceWithConfigure    = (*RecordListResource)(nil)
	_ list.ListResourceWithRawV6Schemas = (*RecordListResource)(nil)
)

type RecordListResource struct {
	meta      *meta.Meta
	domainAPI *domainSDK.API
}

func NewRecordListResource() list.ListResource {
	return &RecordListResource{}
}

func (r *RecordListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	m := listscw.ConfigureMeta(request, response)
	if m == nil {
		return
	}

	r.meta = m
	r.domainAPI = NewDomainAPI(m)
}

func (r *RecordListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_record"
}

func (r *RecordListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_ids": listscw.ProjectIDsAttribute("Project IDs to filter DNS zone records on"),
			"dns_zones": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "DNS zone FQDNs to list records from. Use [\"*\"] to list records across all zones in each selected project. Must contain at least one value.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.Any(
							stringvalidator.OneOf("*", "all"),
							stringvalidator.LengthAtLeast(1),
						),
					),
				},
			},
			"name": listscw.NameAttribute("Name of the DNS zone record to filter on"),
			"type": schema.StringAttribute{
				Description: "Type of the DNS zone record to filter on",
				Optional:    true,
			},
		},
	}
}

func (r *RecordListResource) RawV6Schemas(ctx context.Context, _ list.RawV6SchemaRequest, resp *list.RawV6SchemaResponse) {
	recordResource := ResourceRecord()

	resp.ProtoV6Schema = translate.Schema(recordResource.ProtoSchema(ctx)())
	resp.ProtoV6IdentitySchema = translate.ResourceIdentitySchema(recordResource.ProtoIdentitySchema(ctx)())
}

type RecordListResourceModel struct {
	ProjectIDs types.List   `tfsdk:"project_ids"`
	DNSZones   types.List   `tfsdk:"dns_zones"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
}

func (m *RecordListResourceModel) GetProjects() types.List {
	return m.ProjectIDs
}

type recordListTarget struct {
	ProjectID string
	DNSZone   string
}

type recordListRow struct {
	Record    *domainSDK.Record
	DNSZone   string
	ProjectID string
	RootZone  bool
}

func (r *RecordListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var data RecordListResourceModel

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

	var dnsZoneElems []string

	diags = data.DNSZones.ElementsAs(ctx, &dnsZoneElems, true)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)

		return
	}

	if len(dnsZoneElems) == 0 {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid dns_zones", "`dns_zones` must contain at least one element."),
		})

		return
	}

	if recordListDNSZonesWildcard(dnsZoneElems) && len(dnsZoneElems) != 1 {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid dns_zones", `When using "*" or "all", dns_zones must be exactly ["*"] or ["all"].`),
		})

		return
	}

	targets, targetDiags := r.buildRecordListTargets(ctx, projects, dnsZoneElems)
	if targetDiags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(targetDiags)

		return
	}

	allRows, err := listscw.FetchConcurrently(ctx, targets,
		func(ctx context.Context, target recordListTarget) ([]recordListRow, error) {
			return r.fetchRecordRows(ctx, target, data)
		},
		func(a, b recordListRow) int {
			if a.ProjectID != b.ProjectID {
				return strings.Compare(a.ProjectID, b.ProjectID)
			}

			if a.DNSZone != b.DNSZone {
				return strings.Compare(a.DNSZone, b.DNSZone)
			}

			if a.Record.Type != b.Record.Type {
				return strings.Compare(a.Record.Type.String(), b.Record.Type.String())
			}

			return strings.Compare(a.Record.Name, b.Record.Name)
		},
	)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Listing domain records", "Failed to list domain records: "+err.Error()),
		})

		return
	}

	allRows = dedupeRecordListRows(allRows)

	stream.Results = func(push func(list.ListResult) bool) {
		for _, row := range allRows {
			result := req.NewListResult(ctx)
			result.DisplayName = recordListDisplayName(row)

			recordResource := ResourceRecord()
			resourceData := recordResource.Data(&terraform.InstanceState{})

			if err := identity.SetMultiPartIdentity(resourceData, map[string]string{
				"dns_zone": row.DNSZone,
				"id":       row.Record.ID,
			}, "dns_zone", "id"); err != nil {
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

			setRecordListState(resourceData, row)

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

func recordListDNSZonesWildcard(dnsZones []string) bool {
	return slices.Contains(dnsZones, "*") || slices.Contains(dnsZones, "all")
}

func (r *RecordListResource) buildRecordListTargets(ctx context.Context, projects, dnsZones []string) ([]recordListTarget, diag.Diagnostics) {
	if recordListDNSZonesWildcard(dnsZones) {
		targets := make([]recordListTarget, 0)

		for _, projectID := range projects {
			zones, err := r.listProjectDNSZoneNames(ctx, projectID)
			if err != nil {
				return nil, diag.Diagnostics{
					diag.NewErrorDiagnostic("Listing DNS zones", "Failed to list DNS zones: "+err.Error()),
				}
			}

			for _, zoneName := range zones {
				targets = append(targets, recordListTarget{
					ProjectID: projectID,
					DNSZone:   zoneName,
				})
			}
		}

		return targets, nil
	}

	targets := make([]recordListTarget, 0, len(projects)*len(dnsZones))
	for _, projectID := range projects {
		for _, dnsZone := range dnsZones {
			targets = append(targets, recordListTarget{
				ProjectID: projectID,
				DNSZone:   strings.TrimSuffix(strings.ToLower(strings.TrimSpace(dnsZone)), "."),
			})
		}
	}

	return targets, nil
}

func (r *RecordListResource) listProjectDNSZoneNames(ctx context.Context, projectID string) ([]string, error) {
	response, err := r.domainAPI.ListDNSZones(&domainSDK.ListDNSZonesRequest{
		ProjectID: &projectID,
	}, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	zones := make([]string, 0, len(response.DNSZones))
	for _, zone := range response.DNSZones {
		zones = append(zones, BuildZoneName(zone.Subdomain, zone.Domain))
	}

	return zones, nil
}

func (r *RecordListResource) fetchRecordRows(ctx context.Context, target recordListTarget, data RecordListResourceModel) ([]recordListRow, error) {
	request := &domainSDK.ListDNSZoneRecordsRequest{
		DNSZone:   target.DNSZone,
		ProjectID: &target.ProjectID,
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		request.Name = normalizeRecordName(data.Name.ValueString(), target.DNSZone)
	}

	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		recordType := domainSDK.RecordType(data.Type.ValueString())
		if recordType != domainSDK.RecordTypeUnknown {
			request.Type = recordType
		}
	}

	response, err := r.domainAPI.ListDNSZoneRecords(request, scw.WithContext(ctx), scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	rootZone, err := r.isRootDNSZone(ctx, target.ProjectID, target.DNSZone)
	if err != nil {
		return nil, err
	}

	rows := make([]recordListRow, 0, len(response.Records))
	for _, record := range response.Records {
		if record == nil {
			continue
		}

		rows = append(rows, recordListRow{
			Record:    record,
			DNSZone:   target.DNSZone,
			ProjectID: target.ProjectID,
			RootZone:  rootZone,
		})
	}

	return rows, nil
}

func (r *RecordListResource) isRootDNSZone(ctx context.Context, projectID, dnsZone string) (bool, error) {
	response, err := r.domainAPI.ListDNSZones(&domainSDK.ListDNSZonesRequest{
		ProjectID: &projectID,
		DNSZones:  []string{dnsZone},
	}, scw.WithContext(ctx))
	if err != nil {
		return false, err
	}

	if len(response.DNSZones) == 0 {
		return false, nil
	}

	return response.DNSZones[0].Subdomain == "", nil
}

func dedupeRecordListRows(rows []recordListRow) []recordListRow {
	seen := make(map[string]struct{}, len(rows))
	result := make([]recordListRow, 0, len(rows))

	for _, row := range rows {
		key := row.ProjectID + "/" + row.DNSZone + "/" + row.Record.ID
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		result = append(result, row)
	}

	return result
}

func recordListDisplayName(row recordListRow) string {
	if row.Record.Name == "" || row.Record.Name == "@" {
		return fmt.Sprintf("%s (%s)", row.DNSZone, row.Record.Type.String())
	}

	return fmt.Sprintf("%s.%s (%s)", row.Record.Name, row.DNSZone, row.Record.Type.String())
}

func setRecordListState(d *sdkschema.ResourceData, row recordListRow) {
	record := row.Record
	dnsZone := row.DNSZone

	_ = d.Set("root_zone", row.RootZone)
	_ = d.Set("dns_zone", dnsZone)
	_ = d.Set("name", record.Name)
	_ = d.Set("type", record.Type.String())
	_ = d.Set("data", FlattenDomainData(record.Data, record.Type, dnsZone).(string))
	_ = d.Set("ttl", int(record.TTL))
	_ = d.Set("priority", int(record.Priority))
	_ = d.Set("geo_ip", flattenDomainGeoIP(record.GeoIPConfig))
	_ = d.Set("http_service", flattenDomainHTTPService(record.HTTPServiceConfig))
	_ = d.Set("weighted", flattenDomainWeighted(record.WeightedConfig))
	_ = d.Set("view", flattenDomainView(record.ViewConfig))
	_ = d.Set("project_id", row.ProjectID)

	if record.Name == "" || record.Name == "@" {
		_ = d.Set("fqdn", dnsZone)
	} else {
		_ = d.Set("fqdn", fmt.Sprintf("%s.%s", record.Name, dnsZone))
	}
}
