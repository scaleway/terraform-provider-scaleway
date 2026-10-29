package webhosting

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	scwds "github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*WebhostingDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*WebhostingDataSource)(nil)
)

func NewWebhostingDataSource() datasource.DataSource {
	return &WebhostingDataSource{}
}

type WebhostingDataSource struct {
	api  *webhosting.HostingAPI
	meta *meta.Meta
}

type webhostingDataSourceModel struct {
	CpanelURLs       types.List   `tfsdk:"cpanel_urls"`
	Options          types.List   `tfsdk:"options"`
	NameServers      types.List   `tfsdk:"name_servers"`
	Records          types.List   `tfsdk:"records"`
	Tags             types.List   `tfsdk:"tags"`
	OptionIDs        types.List   `tfsdk:"option_ids"`
	Username         types.String `tfsdk:"username"`
	ID               types.String `tfsdk:"id"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	OfferName        types.String `tfsdk:"offer_name"`
	Domain           types.String `tfsdk:"domain"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	Status           types.String `tfsdk:"status"`
	CreatedAt        types.String `tfsdk:"created_at"`
	DnsStatus        types.String `tfsdk:"dns_status"`
	WebhostingID     types.String `tfsdk:"webhosting_id"`
	PlatformHostname types.String `tfsdk:"platform_hostname"`
	Email            types.String `tfsdk:"email"`
	OfferID          types.String `tfsdk:"offer_id"`
	Region           types.String `tfsdk:"region"`
	ProjectID        types.String `tfsdk:"project_id"`
	PlatformNumber   types.Int64  `tfsdk:"platform_number"`
}

func (d *WebhostingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhosting"
}

func (d *WebhostingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		MarkdownDescription: "Finds a Scaleway WebHosting by ID or domain.",
		Attributes: map[string]fwschema.Attribute{
			"webhosting_id": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The ID of the webhosting",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
					stringvalidator.ConflictsWith(path.MatchRoot("domain")),
				},
			},
			"domain": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The domain name of the hosting",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("webhosting_id")),
				},
			},
			"offer_id": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the selected offer for the hosting",
			},
			"email": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Contact email of the client for the hosting",
			},
			"tags": fwschema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tags of the hosting",
			},
			"option_ids": fwschema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "IDs of the selected options for the hosting",
			},
			"id": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the webhosting",
			},
			"created_at": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of hosting's creation (RFC 3339 format)",
			},
			"updated_at": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of hosting's last update (RFC 3339 format)",
			},
			"status": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The hosting status",
			},
			"platform_hostname": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Hostname of the host platform",
			},
			"platform_number": fwschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of the host platform",
			},
			"offer_name": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the active offer",
			},
			"options": fwschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Active options of the hosting",
				NestedObject: fwschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"id": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the active option",
						},
						"name": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the option",
						},
					},
				},
			},
			"dns_status": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DNS status of the hosting",
			},
			"cpanel_urls": fwschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "URL to connect to cPanel Dashboard and to Webmail interface",
				NestedObject: fwschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"dashboard": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "URL to connect to dashboard interface",
						},
						"webmail": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "URL to connect to Webmail interface",
						},
					},
				},
			},
			"username": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Main hosting cPanel username",
			},
			"records": fwschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of DNS records associated with the webhosting.",
				NestedObject: fwschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"name": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the DNS record",
						},
						"type": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Type of the DNS record",
						},
						"ttl": fwschema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Time to live in seconds of the record",
						},
						"value": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Value of the DNS record",
						},
						"priority": fwschema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Priority of DNS records associated with the webhosting.",
						},
						"status": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Status of the hosting record",
						},
					},
				},
			},
			"name_servers": fwschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of nameservers associated with the webhosting.",
				NestedObject: fwschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"hostname": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Hostname of the server",
						},
						"status": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Status of the nameserver",
						},
						"is_default": fwschema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether or not the webhosting is the default one",
						},
					},
				},
			},
			"region": fwschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region of the hosting",
			},
			"project_id": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The project ID the resource is associated to",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"organization_id": fwschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The organization ID the resource is associated to",
			},
		},
	}
}

func (d *WebhostingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.meta = m
	d.api = webhosting.NewHostingAPI(d.meta.ScwClient())
}

func (d *WebhostingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config webhostingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	webhostingID := config.WebhostingID.ValueString()

	if webhostingID == "" {
		hostingDomain := config.Domain.ValueString()

		listReq := &webhosting.HostingAPIListHostingsRequest{
			Region: region,
			Domain: &hostingDomain,
		}
		if v := config.ProjectID.ValueString(); v != "" {
			listReq.ProjectID = &v
		}

		if v := config.OrganizationID.ValueString(); v != "" {
			listReq.OrganizationID = &v
		}

		listResp, err := d.api.ListHostings(listReq, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to list webhostings", err.Error())

			return
		}

		found, err := scwds.FindExact(
			listResp.Hostings,
			func(s *webhosting.HostingSummary) bool { return s.Domain != nil && *s.Domain == hostingDomain }, //nolint:staticcheck // deprecated in SDK, used until domain_info available here
			hostingDomain,
		)
		if err != nil {
			resp.Diagnostics.AddError("Webhosting not found", err.Error())

			return
		}

		webhostingID = found.ID
	}

	// webhosting_id may be a regional ID; parse to get the raw UUID and region.
	parsedRegion, parsedID, parseErr := regional.ParseID(webhostingID)
	if parseErr == nil {
		region = parsedRegion
		webhostingID = parsedID
	}

	regionalID := scwds.NewRegionalID(webhostingID, region)

	hostingResponse, err := waitForHosting(ctx, d.api, region, webhostingID, defaultHostingTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read webhosting", err.Error())

		return
	}

	state := flattenWebhostingDataSource(ctx, d.meta, hostingResponse, &resp.Diagnostics)
	state.ID = types.StringValue(regionalID)
	state.WebhostingID = types.StringValue(regionalID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// flattenWebhostingDataSource mirrors the resource's flattenWebhosting, exposed as a
// free function so the data source can reuse it without a *WebhostingResource.
func flattenWebhostingDataSource(
	ctx context.Context, m *meta.Meta, h *webhosting.Hosting, diags *diag.Diagnostics,
) webhostingDataSourceModel {
	state := webhostingDataSourceModel{
		ID:             types.StringValue(regional.NewIDString(h.Region, h.ID)),
		OfferID:        types.StringValue(regional.NewIDString(h.Region, h.Offer.ID)),
		Region:         types.StringValue(h.Region.String()),
		ProjectID:      types.StringValue(h.ProjectID),
		OrganizationID: types.StringNull(),
		Status:         types.StringValue(h.Status.String()),
		OfferName:      types.StringValue(h.Offer.Name),
		Options:        flattenHostingOptionsList(ctx, h.Offer.Options, diags),
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

	// option_ids is not returned by the API; leave null on the data source.
	state.OptionIDs = types.ListNull(types.StringType)

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
