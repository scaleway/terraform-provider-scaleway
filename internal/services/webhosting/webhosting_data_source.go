package webhosting

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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
	WebhostingID types.String `tfsdk:"webhosting_id"`
	webhostingResourceModel
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

	//nolint:modernize // cannot specify promoted field ID and enclosing embedded field webhostingResourceModel
	state := webhostingDataSourceModel{
		webhostingResourceModel: flattenWebhosting(ctx, d.meta, hostingResponse, &resp.Diagnostics),
		WebhostingID:            types.StringValue(regionalID),
	}

	state.ID = types.StringValue(regionalID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
