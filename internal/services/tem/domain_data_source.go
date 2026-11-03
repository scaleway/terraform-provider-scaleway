package tem

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	scwds "github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/tem_domain_data_source.md
var domainDataSourceDescription string

var (
	_ datasource.DataSource              = (*DomainDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*DomainDataSource)(nil)
)

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

type DomainDataSource struct {
	api  *tem.API
	meta *meta.Meta
}

type domainDataSourceModel struct {
	Reputation           types.List   `tfsdk:"reputation"`
	DmarcConfig          types.String `tfsdk:"dmarc_config"`
	DkimConfig           types.String `tfsdk:"dkim_config"`
	ProjectID            types.String `tfsdk:"project_id"`
	Region               types.String `tfsdk:"region"`
	Status               types.String `tfsdk:"status"`
	CreatedAt            types.String `tfsdk:"created_at"`
	NextCheckAt          types.String `tfsdk:"next_check_at"`
	LastValidAt          types.String `tfsdk:"last_valid_at"`
	RevokedAt            types.String `tfsdk:"revoked_at"`
	LastError            types.String `tfsdk:"last_error"`
	SpfConfig            types.String `tfsdk:"spf_config"`
	SmtpsAuthUser        types.String `tfsdk:"smtps_auth_user"`
	DmarcName            types.String `tfsdk:"dmarc_name"`
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	MxConfig             types.String `tfsdk:"mx_config"`
	DkimName             types.String `tfsdk:"dkim_name"`
	DomainID             types.String `tfsdk:"domain_id"`
	SmtpHost             types.String `tfsdk:"smtp_host"`
	SpfValue             types.String `tfsdk:"spf_value"`
	MxBlackhole          types.String `tfsdk:"mx_blackhole"`
	SmtpPortUnsecure     types.Int64  `tfsdk:"smtp_port_unsecure"`
	SmtpsPort            types.Int64  `tfsdk:"smtps_port"`
	SmtpPortAlternative  types.Int64  `tfsdk:"smtp_port_alternative"`
	SmtpsPortAlternative types.Int64  `tfsdk:"smtps_port_alternative"`
	SmtpPort             types.Int64  `tfsdk:"smtp_port"`
	MxPriority           types.Int64  `tfsdk:"mx_priority"`
	Autoconfig           types.Bool   `tfsdk:"autoconfig"`
	AcceptTos            types.Bool   `tfsdk:"accept_tos"`
}

func (d *DomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tem_domain"
}

func (d *DomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: domainDataSourceDescription,
		Description:         domainDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"domain_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The ID of the tem domain",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithRegion(),
					stringvalidator.ConflictsWith(path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The domain name used when sending emails",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("domain_id")),
				},
			},
			"accept_tos": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Accept the Scaleway Terms of Service",
			},
			"autoconfig": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable automatic configuration options for the domain",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the domain",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of domain's creation (RFC 3339 format)",
			},
			"next_check_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of the next scheduled check (RFC 3339 format)",
			},
			"last_valid_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time the domain was last found to be valid (RFC 3339 format)",
			},
			"revoked_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of the revocation of the domain (RFC 3339 format)",
			},
			"last_error": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Error message if the last check failed",
				DeprecationMessage:  "last_error is deprecated",
			},
			"spf_config": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Snippet of the SPF record that should be registered in the DNS zone",
			},
			"dkim_config": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DKIM public key, as should be recorded in the DNS zone",
			},
			"dmarc_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DMARC name for the domain, as should be recorded in the DNS zone",
			},
			"dmarc_config": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DMARC record for the domain, as should be recorded in the DNS zone",
			},
			"dkim_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DKIM name for the domain, as should be recorded in the DNS zone",
			},
			"spf_value": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Complete SPF record value for the domain, as should be recorded in the DNS zone",
			},
			"mx_config": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "MX exchange hostname for the domain blackhole (without priority), suitable for scaleway_domain_record.data",
			},
			"mx_priority": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "MX priority for the domain blackhole, suitable for scaleway_domain_record.priority",
			},
			"smtp_host": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SMTP host to use to send emails",
			},
			"smtp_port_unsecure": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: fmt.Sprintf("SMTP port to use to send emails. (Port %d)", tem.SMTPPortUnsecure),
			},
			"smtp_port": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: fmt.Sprintf("SMTP port to use to send emails over TLS. (Port %d)", tem.SMTPPort),
			},
			"smtp_port_alternative": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: fmt.Sprintf("SMTP port to use to send emails over TLS. (Port %d)", tem.SMTPPortAlternative),
			},
			"smtps_port": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: fmt.Sprintf("SMTPS port to use to send emails over TLS Wrapper. (Port %d)", tem.SMTPSPort),
			},
			"smtps_auth_user": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SMTPS auth user refers to the identifier for a user authorized to send emails via SMTPS, ensuring secure email transmission",
			},
			"smtps_port_alternative": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: fmt.Sprintf("SMTPS port to use to send emails over TLS Wrapper. (Port %d)", tem.SMTPSPortAlternative),
			},
			"mx_blackhole": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Scaleway's blackhole MX server to use",
			},
			"reputation": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The domain's reputation",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Status of the domain's reputation",
						},
						"score": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "A range from 0 to 100 that determines your domain's reputation score",
						},
						"scored_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Time and date the score was calculated",
						},
						"previous_score": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The previously-calculated domain's reputation score",
						},
						"previous_scored_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Time and date the previous reputation score was calculated",
						},
					},
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region you want to attach the resource to",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(regional.AllRegions()),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project_id you want to attach the resource to",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
		},
	}
}

func (d *DomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = tem.NewAPI(d.meta.ScwClient())
}

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config domainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	var domainID string

	if domainID = config.DomainID.ValueString(); domainID == "" {
		listReq := &tem.ListDomainsRequest{
			Region: region,
		}

		if name := config.Name.ValueString(); name != "" {
			listReq.Name = scwtypes.ExpandStringPtr(name)
		}

		if projectID := config.ProjectID.ValueString(); projectID != "" {
			listReq.ProjectID = scwtypes.ExpandStringPtr(projectID)
		}

		res, listErr := d.api.ListDomains(listReq, scw.WithContext(ctx))
		if listErr != nil {
			resp.Diagnostics.AddError("Failed to list TEM domains", listErr.Error())

			return
		}

		name := config.Name.ValueString()

		found, findErr := scwds.FindExact(res.Domains, func(domain *tem.Domain) bool {
			return domain.Status != tem.DomainStatusRevoked && domain.Name == name
		}, name)
		if findErr != nil {
			resp.Diagnostics.AddError("TEM domain lookup failed", findErr.Error())

			return
		}

		domainID = found.ID
	}

	region, id, err := regional.ResolveID(domainID, config.Region.ValueString(), d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve TEM domain ID", err.Error())

		return
	}

	regionalID := regional.NewIDString(region, id)

	domain, err := d.api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.Diagnostics.AddError("TEM domain not found", fmt.Sprintf("tem domain (%s) not found", regionalID))

			return
		}

		resp.Diagnostics.AddError("Failed to read TEM domain", err.Error())

		return
	}

	state := domainToDataSourceModel(flattenDomain(ctx, domain, region, &resp.Diagnostics), regionalID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func domainToDataSourceModel(m domainModel, domainID string) domainDataSourceModel {
	return domainDataSourceModel{
		ID:                   m.ID,
		DomainID:             types.StringValue(domainID),
		Name:                 m.Name,
		AcceptTos:            m.AcceptTos,
		Autoconfig:           m.Autoconfig,
		Status:               m.Status,
		CreatedAt:            m.CreatedAt,
		NextCheckAt:          m.NextCheckAt,
		LastValidAt:          m.LastValidAt,
		RevokedAt:            m.RevokedAt,
		LastError:            m.LastError,
		SpfConfig:            m.SpfConfig,
		DkimConfig:           m.DkimConfig,
		DmarcName:            m.DmarcName,
		DmarcConfig:          m.DmarcConfig,
		DkimName:             m.DkimName,
		SpfValue:             m.SpfValue,
		MxConfig:             m.MxConfig,
		MxPriority:           m.MxPriority,
		SmtpHost:             m.SmtpHost,
		SmtpPortUnsecure:     m.SmtpPortUnsecure,
		SmtpPort:             m.SmtpPort,
		SmtpPortAlternative:  m.SmtpPortAlternative,
		SmtpsPort:            m.SmtpsPort,
		SmtpsAuthUser:        m.SmtpsAuthUser,
		SmtpsPortAlternative: m.SmtpsPortAlternative,
		MxBlackhole:          m.MxBlackhole,
		Reputation:           m.Reputation,
		Region:               m.Region,
		ProjectID:            m.ProjectID,
	}
}
