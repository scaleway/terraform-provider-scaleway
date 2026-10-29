package webhosting

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*WebhostingResource)(nil)
	_ resource.ResourceWithConfigure   = (*WebhostingResource)(nil)
	_ resource.ResourceWithImportState = (*WebhostingResource)(nil)
	_ resource.ResourceWithIdentity    = (*WebhostingResource)(nil)
)

type WebhostingResource struct {
	api  *webhosting.HostingAPI
	meta *meta.Meta
}

func NewWebhostingResource() resource.Resource {
	return &WebhostingResource{}
}

type webhostingResourceModel struct {
	Options          types.List   `tfsdk:"options"`
	NameServers      types.List   `tfsdk:"name_servers"`
	Records          types.List   `tfsdk:"records"`
	CpanelURLs       types.List   `tfsdk:"cpanel_urls"`
	Tags             types.List   `tfsdk:"tags"`
	OptionIDs        types.List   `tfsdk:"option_ids"`
	Username         types.String `tfsdk:"username"`
	Email            types.String `tfsdk:"email"`
	Status           types.String `tfsdk:"status"`
	PlatformHostname types.String `tfsdk:"platform_hostname"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	OfferName        types.String `tfsdk:"offer_name"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DnsStatus        types.String `tfsdk:"dns_status"`
	Domain           types.String `tfsdk:"domain"`
	OfferID          types.String `tfsdk:"offer_id"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	ID               types.String `tfsdk:"id"`
	Region           types.String `tfsdk:"region"`
	ProjectID        types.String `tfsdk:"project_id"`
	PlatformNumber   types.Int64  `tfsdk:"platform_number"`
}

type webhostingOptionsModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type webhostingCpanelURLsModel struct {
	Dashboard types.String `tfsdk:"dashboard"`
	Webmail   types.String `tfsdk:"webmail"`
}

type webhostingRecordsModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Value    types.String `tfsdk:"value"`
	Status   types.String `tfsdk:"status"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int64  `tfsdk:"priority"`
}

type webhostingNameServersModel struct {
	Hostname  types.String `tfsdk:"hostname"`
	Status    types.String `tfsdk:"status"`
	IsDefault types.Bool   `tfsdk:"is_default"`
}

type webhostingResourceIdentityModel = identityfw.RegionalIdentity

// resolveHostingRegionID resolves the region and hosting ID from the identity
// when available, falling back to parsing the state ID.
func resolveHostingRegionID(identity webhostingResourceIdentityModel, fallbackID string) (scw.Region, string, error) {
	if !identity.ID.IsNull() && !identity.ID.IsUnknown() {
		region, id := identityfw.ResolveRegionalIdentity(identity)

		return region, id, nil
	}

	return regional.ParseID(fallbackID)
}

func (r *WebhostingResource) Metadata(
	_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_webhosting"
}

func (r *WebhostingResource) Schema(
	_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Scaleway WebHosting",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the webhosting",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"offer_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the selected offer for the hosting",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
				},
			},
			"email": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Contact email of the client for the hosting",
				Validators: []validator.String{
					verify.IsStringEmail(),
				},
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain name of the hosting",
			},
			"tags": schema.ListAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags of the hosting",
				PlanModifiers: []planmodifier.List{
					preserveInternalTag{},
				},
			},
			"option_ids": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "IDs of the selected options for the hosting",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of hosting's creation (RFC 3339 format)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of hosting's last update (RFC 3339 format)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The hosting status",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"platform_hostname": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Hostname of the host platform",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"platform_number": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of the host platform",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"offer_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the active offer",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"options": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Active options of the hosting",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the active option",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the option",
						},
					},
				},
			},
			"dns_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DNS status of the hosting",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpanel_urls": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "URL to connect to cPanel Dashboard and to Webmail interface",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dashboard": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "URL to connect to dashboard interface",
						},
						"webmail": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "URL to connect to Webmail interface",
						},
					},
				},
			},
			"username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Main hosting cPanel username",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"records": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of DNS records associated with the webhosting.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the DNS record",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Type of the DNS record",
						},
						"ttl": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Time to live in seconds of the record",
						},
						"value": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Value of the DNS record",
						},
						"priority": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Priority of DNS records associated with the webhosting.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Status of the hosting record",
						},
					},
				},
			},
			"name_servers": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of nameservers associated with the webhosting.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"hostname": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Hostname of the server",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Status of the nameserver",
						},
						"is_default": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether or not the webhosting is the default one",
						},
					},
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region of the hosting",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the hosting belongs to",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The organization ID the hosting belongs to. Deprecated: this field is deprecated and will be removed in the next major version.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *WebhostingResource) IdentitySchema(
	_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse,
) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *WebhostingResource) ImportState(
	ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *WebhostingResource) Configure(
	_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse,
) {
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
	r.api = webhosting.NewHostingAPI(r.meta.ScwClient())
}

func (r *WebhostingResource) Create(
	ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse,
) {
	var data webhostingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(data.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	_, offerID, err := regional.ParseID(data.OfferID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse offer_id", err.Error())

		return
	}

	createReq := &webhosting.HostingAPICreateHostingRequest{
		Region:       region,
		OfferID:      offerID,
		ProjectID:    projectID,
		Domain:       data.Domain.ValueString(),
		Email:        data.Email.ValueString(),
		Tags:         scwtypes.ExpandStringList(ctx, data.Tags, &resp.Diagnostics),
		OfferOptions: expandOfferOptionIDs(ctx, data.OptionIDs, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	hostingResponse, err := r.api.CreateHosting(createReq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create webhosting", err.Error())

		return
	}

	hostingResponse, err = waitForHosting(ctx, r.api, region, hostingResponse.ID, defaultHostingTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for webhosting", err.Error())

		return
	}

	state := flattenWebhosting(ctx, r.meta, hostingResponse, &resp.Diagnostics)
	state.OptionIDs = data.OptionIDs
	state.Email = data.Email
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(hostingResponse.Region, hostingResponse.ID))...)
}

func (r *WebhostingResource) Read(
	ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse,
) {
	var (
		state    webhostingResourceModel
		identity webhostingResourceIdentityModel
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

	region, id, err := resolveHostingRegionID(identity, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse webhosting ID during READ", err.Error())

		return
	}

	hostingResponse, err := waitForHosting(ctx, r.api, region, id, defaultHostingTimeout)
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read webhosting", err.Error())

		return
	}

	newState := flattenWebhosting(ctx, r.meta, hostingResponse, &resp.Diagnostics)
	newState.OptionIDs = state.OptionIDs
	newState.Email = state.Email
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(hostingResponse.Region, hostingResponse.ID))...)
}

func (r *WebhostingResource) Update(
	ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	var (
		plan     webhostingResourceModel
		state    webhostingResourceModel
		identity webhostingResourceIdentityModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := resolveHostingRegionID(identity, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse webhosting ID during UPDATE", err.Error())

		return
	}

	res, err := waitForHosting(ctx, r.api, region, id, defaultHostingTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read webhosting before update", err.Error())

		return
	}

	updateReq := &webhosting.HostingAPIUpdateHostingRequest{
		Region:    region,
		HostingID: res.ID,
	}

	hasChanged := false

	if !plan.OptionIDs.Equal(state.OptionIDs) {
		updateReq.OfferOptions = expandOfferOptionIDs(ctx, plan.OptionIDs, &resp.Diagnostics)
		hasChanged = true
	}

	if !plan.OfferID.Equal(state.OfferID) {
		_, offerID, err := regional.ParseID(plan.OfferID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse offer_id", err.Error())

			return
		}

		updateReq.OfferID = &offerID
		hasChanged = true
	}

	if !plan.Email.Equal(state.Email) {
		email := plan.Email.ValueString()
		updateReq.Email = &email
		hasChanged = true
	}

	if !plan.Tags.Equal(state.Tags) {
		tags := scwtypes.ExpandStringList(ctx, plan.Tags, &resp.Diagnostics)
		updateReq.Tags = &tags
		hasChanged = true
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if hasChanged {
		if _, err := r.api.UpdateHosting(updateReq, scw.WithContext(ctx)); err != nil {
			resp.Diagnostics.AddError("Failed to update webhosting", err.Error())

			return
		}
	}

	hostingResponse, err := waitForHosting(ctx, r.api, region, id, defaultHostingTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read webhosting after update", err.Error())

		return
	}

	newState := flattenWebhosting(ctx, r.meta, hostingResponse, &resp.Diagnostics)
	newState.OptionIDs = plan.OptionIDs
	newState.Email = plan.Email
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(hostingResponse.Region, hostingResponse.ID))...)
}

func (r *WebhostingResource) Delete(
	ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var (
		state    webhostingResourceModel
		identity webhostingResourceIdentityModel
	)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := resolveHostingRegionID(identity, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse webhosting ID during DELETE", err.Error())

		return
	}

	_, err = waitForHosting(ctx, r.api, region, id, defaultHostingTimeout)
	if err != nil {
		return
	}

	_, err = r.api.DeleteHosting(&webhosting.HostingAPIDeleteHostingRequest{
		Region:    region,
		HostingID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to delete webhosting", err.Error())

		return
	}

	_, _ = waitForHosting(ctx, r.api, region, id, defaultHostingTimeout)
}

// flattenWebhosting builds the framework state model from an API Hosting response.
// option_ids and email are intentionally not preserved from prior state here:
// callers (resource CRUD) override option_ids and email from plan/prior state
// after calling this function. The data source reads email directly from the API.
func flattenWebhosting(
	ctx context.Context, m *meta.Meta, h *webhosting.Hosting, diags *diag.Diagnostics,
) webhostingResourceModel {
	state := webhostingResourceModel{
		ID:             types.StringValue(regional.NewIDString(h.Region, h.ID)),
		OfferID:        types.StringValue(regional.NewIDString(h.Region, h.Offer.ID)),
		Region:         types.StringValue(h.Region.String()),
		ProjectID:      types.StringValue(h.ProjectID),
		OrganizationID: types.StringNull(),
		Status:         types.StringValue(h.Status.String()),
		OfferName:      types.StringValue(h.Offer.Name),
		Options:        flattenHostingOptionsList(ctx, h.Offer.Options, diags),
		OptionIDs:      types.ListNull(types.StringType),
	}

	if h.Domain != nil { //nolint:staticcheck
		state.Domain = types.StringValue(*h.Domain) //nolint:staticcheck
	}

	if h.User != nil {
		state.Username = types.StringValue(h.User.Username)
		state.Email = types.StringValue(h.User.ContactEmail)
	}

	if h.Platform != nil {
		state.PlatformHostname = types.StringValue(h.Platform.Hostname)
		state.PlatformNumber = types.Int64Value(int64(h.Platform.Number))

		if h.Platform.ControlPanel != nil && h.Platform.ControlPanel.URLs != nil {
			state.CpanelURLs = flattenCpanelURLsList(ctx, h.Platform.ControlPanel.URLs, diags)
		}
	}

	if h.DNSStatus != nil { //nolint:staticcheck
		state.DnsStatus = types.StringValue(h.DNSStatus.String()) //nolint:staticcheck
	}

	if h.CreatedAt != nil {
		state.CreatedAt = types.StringValue(h.CreatedAt.Format(time.RFC3339))
	}

	if h.UpdatedAt != nil {
		state.UpdatedAt = types.StringValue(h.UpdatedAt.Format(time.RFC3339))
	}

	tagsList, d := types.ListValueFrom(ctx, types.StringType, h.Tags)
	diags.Append(d...)

	state.Tags = tagsList

	// DNS records and nameservers come from the DNS API, not the hosting response.
	if h.Domain != nil { //nolint:staticcheck
		dnsAPI := webhosting.NewDnsAPI(m.ScwClient())

		dnsResp, err := dnsAPI.GetDomainDNSRecords(&webhosting.DNSAPIGetDomainDNSRecordsRequest{
			Region: h.Region,
			Domain: *h.Domain, //nolint:staticcheck // deprecated in SDK, kept until domain_info fully propagated
		}, scw.WithContext(ctx))
		if err != nil {
			diags.AddError("Failed to get webhosting DNS records", err.Error())
		} else {
			state.Records = flattenDNSRecordsList(ctx, dnsResp.Records, diags)
			state.NameServers = flattenNameServersList(ctx, dnsResp.NameServers, diags)
		}
	}

	return state
}

func expandOfferOptionIDs(
	ctx context.Context, list types.List, diags *diag.Diagnostics,
) []*webhosting.OfferOptionRequest {
	ids := scwtypes.ExpandStringList(ctx, list, diags)

	options := make([]*webhosting.OfferOptionRequest, 0, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}

		options = append(options, &webhosting.OfferOptionRequest{
			ID:       id,
			Quantity: 1,
		})
	}

	return options
}

func flattenHostingOptionsList(
	ctx context.Context, options []*webhosting.OfferOption, diags *diag.Diagnostics,
) types.List {
	if len(options) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: hostingOptionsAttrTypes()})
	}

	models := make([]webhostingOptionsModel, 0, len(options))
	for _, option := range options {
		models = append(models, webhostingOptionsModel{
			ID:   types.StringValue(option.ID),
			Name: types.StringValue(string(option.Name)),
		})
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: hostingOptionsAttrTypes()}, models)
	diags.Append(d...)

	return list
}

func flattenCpanelURLsList(
	ctx context.Context, urls *webhosting.PlatformControlPanelURLs, diags *diag.Diagnostics,
) types.List {
	if urls == nil {
		return types.ListNull(types.ObjectType{AttrTypes: cpanelURLsAttrTypes()})
	}

	models := []webhostingCpanelURLsModel{{
		Dashboard: types.StringValue(urls.Dashboard),
		Webmail:   types.StringValue(urls.Webmail),
	}}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: cpanelURLsAttrTypes()}, models)
	diags.Append(d...)

	return list
}

func flattenDNSRecordsList(
	ctx context.Context, records []*webhosting.DNSRecord, diags *diag.Diagnostics,
) types.List {
	if len(records) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: dnsRecordsAttrTypes()})
	}

	models := make([]webhostingRecordsModel, 0, len(records))
	for _, record := range records {
		name := record.Name
		if name == "" {
			name = extractDNSRecordName(record.RawData)
		}

		m := webhostingRecordsModel{
			Name:   types.StringValue(name),
			Type:   types.StringValue(record.Type.String()),
			TTL:    types.Int64Value(int64(record.TTL)),
			Value:  types.StringValue(record.Value),
			Status: types.StringValue(record.Status.String()),
		}

		if record.Priority != nil {
			m.Priority = types.Int64Value(int64(*record.Priority))
		} else {
			m.Priority = types.Int64Null()
		}

		models = append(models, m)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: dnsRecordsAttrTypes()}, models)
	diags.Append(d...)

	return list
}

func flattenNameServersList(
	ctx context.Context, servers []*webhosting.Nameserver, diags *diag.Diagnostics,
) types.List {
	if len(servers) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: nameServersAttrTypes()})
	}

	models := make([]webhostingNameServersModel, 0, len(servers))
	for _, server := range servers {
		models = append(models, webhostingNameServersModel{
			Hostname:  types.StringValue(server.Hostname),
			Status:    types.StringValue(server.Status.String()),
			IsDefault: types.BoolValue(server.IsDefault),
		})
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: nameServersAttrTypes()}, models)
	diags.Append(d...)

	return list
}

func hostingOptionsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	}
}

func cpanelURLsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"dashboard": types.StringType,
		"webmail":   types.StringType,
	}
}

func dnsRecordsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.StringType,
		"type":     types.StringType,
		"ttl":      types.Int64Type,
		"value":    types.StringType,
		"priority": types.Int64Type,
		"status":   types.StringType,
	}
}

func nameServersAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"hostname":   types.StringType,
		"status":     types.StringType,
		"is_default": types.BoolType,
	}
}

// preserveInternalTag is a list plan modifier that mirrors the SDK CustomizeDiff:
// if the "internal" tag was present in the prior state and is now removed, the
// entire prior tags list is restored, preventing users from dropping "internal".
type preserveInternalTag struct{}

func (preserveInternalTag) Description(_ context.Context) string {
	return "Prevents the 'internal' tag from being removed once it has been applied."
}

func (preserveInternalTag) MarkdownDescription(_ context.Context) string {
	return "Prevents the `internal` tag from being removed once it has been applied."
}

func (preserveInternalTag) PlanModifyList(
	ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse,
) {
	// Only applies when a prior state exists (i.e. updates, not creates).
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	var oldTags, newTags []string
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &oldTags, false)...)
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &newTags, false)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !slices.Contains(oldTags, "internal") || slices.Contains(newTags, "internal") {
		return
	}

	// "internal" was present and is being dropped: restore the prior tags.
	resp.PlanValue = req.StateValue
}
