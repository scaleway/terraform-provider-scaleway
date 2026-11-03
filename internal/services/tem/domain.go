package tem

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/tem_domain_resource.md
var domainResourceDescription string

var (
	_ resource.Resource                = (*DomainResource)(nil)
	_ resource.ResourceWithConfigure   = (*DomainResource)(nil)
	_ resource.ResourceWithImportState = (*DomainResource)(nil)
	_ resource.ResourceWithIdentity    = (*DomainResource)(nil)
)

func NewDomainResource() resource.Resource {
	return &DomainResource{}
}

type DomainResource struct {
	api  *tem.API
	meta *meta.Meta
}

type domainModel struct {
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
	ID                   types.String   `tfsdk:"id"`
	Name                 types.String   `tfsdk:"name"`
	AcceptTos            types.Bool     `tfsdk:"accept_tos"`
	Autoconfig           types.Bool     `tfsdk:"autoconfig"`
	Status               types.String   `tfsdk:"status"`
	CreatedAt            types.String   `tfsdk:"created_at"`
	NextCheckAt          types.String   `tfsdk:"next_check_at"`
	LastValidAt          types.String   `tfsdk:"last_valid_at"`
	RevokedAt            types.String   `tfsdk:"revoked_at"`
	LastError            types.String   `tfsdk:"last_error"`
	SpfConfig            types.String   `tfsdk:"spf_config"`
	DkimConfig           types.String   `tfsdk:"dkim_config"`
	DmarcName            types.String   `tfsdk:"dmarc_name"`
	DmarcConfig          types.String   `tfsdk:"dmarc_config"`
	DkimName             types.String   `tfsdk:"dkim_name"`
	SpfValue             types.String   `tfsdk:"spf_value"`
	MxConfig             types.String   `tfsdk:"mx_config"`
	MxPriority           types.Int64    `tfsdk:"mx_priority"`
	SmtpHost             types.String   `tfsdk:"smtp_host"`
	SmtpPortUnsecure     types.Int64    `tfsdk:"smtp_port_unsecure"`
	SmtpPort             types.Int64    `tfsdk:"smtp_port"`
	SmtpPortAlternative  types.Int64    `tfsdk:"smtp_port_alternative"`
	SmtpsPort            types.Int64    `tfsdk:"smtps_port"`
	SmtpsAuthUser        types.String   `tfsdk:"smtps_auth_user"`
	SmtpsPortAlternative types.Int64    `tfsdk:"smtps_port_alternative"`
	MxBlackhole          types.String   `tfsdk:"mx_blackhole"`
	Reputation           types.List     `tfsdk:"reputation"`
	Region               types.String   `tfsdk:"region"`
	ProjectID            types.String   `tfsdk:"project_id"`
}

type domainResourceIdentityModel = identityfw.RegionalIdentity

func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tem_domain"
}

func (r *DomainResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: domainResourceDescription,
		Description:         domainResourceDescription,
		Attributes:          domainSchemaAttributes(ctx),
	}
}

func domainSchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "The domain name used when sending emails",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"accept_tos": schema.BoolAttribute{
			Required:            true,
			MarkdownDescription: "Accept the Scaleway Terms of Service",
			Validators: []validator.Bool{
				acceptTosValidator{},
			},
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"autoconfig": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
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
		"region": regional.SchemaAttributeComputed(),
		"project_id": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: "The project_id you want to attach the resource to",
			Validators: []validator.String{
				verify.IsStringUUID(),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
			Create: true,
			Update: true,
			Delete: true,
		}),
	}
}

type acceptTosValidator struct{}

func (v acceptTosValidator) Description(_ context.Context) string {
	return "you must accept the Scaleway Terms of Service to use this service"
}

func (v acceptTosValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v acceptTosValidator) ValidateBool(_ context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Attribute Value",
			"you must accept the Scaleway Terms of Service to use this service",
		)
	}
}

func (r *DomainResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *DomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = tem.NewAPI(r.meta.ScwClient())
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	acceptTos := plan.AcceptTos.ValueBool()

	domain, err := r.api.CreateDomain(&tem.CreateDomainRequest{
		Region:     region,
		ProjectID:  projectID,
		DomainName: plan.Name.ValueString(),
		AcceptTos:  &acceptTos, //nolint:staticcheck
		Autoconfig: plan.Autoconfig.ValueBool(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create TEM domain", err.Error())

		return
	}

	timeout, diags := plan.Timeouts.Create(ctx, DefaultDomainCreateTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// TEM Create can return the domain before Get is consistent; retry transient 404s.
	domain, err = transport.RetryOnStatusValue(ctx, transport.NotFound, func(ctx context.Context) (*tem.Domain, error) {
		return WaitForDomain(ctx, r.api, region, domain.ID, timeout)
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for TEM domain", err.Error())

		return
	}

	state := flattenDomain(ctx, domain, region, &resp.Diagnostics)
	state.Timeouts = plan.Timeouts
	state.AcceptTos = types.BoolValue(true)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(domain.Region, domain.ID))...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    domainModel
		identity domainResourceIdentityModel
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

	var (
		id     string
		region scw.Region
		err    error
	)

	if identityAvailable {
		region, id = identityfw.ResolveRegionalIdentity(identity)
	} else {
		region, id, err = regional.ParseID(state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to parse TEM domain ID", err.Error())

			return
		}
	}

	domain, err := r.api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read TEM domain", err.Error())

		return
	}

	newState := flattenDomain(ctx, domain, region, &resp.Diagnostics)
	newState.Timeouts = state.Timeouts
	newState.AcceptTos = types.BoolValue(true)

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(domain.Region, domain.ID))...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan  domainModel
		state domainModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse TEM domain ID", err.Error())

		return
	}

	if !plan.Autoconfig.Equal(state.Autoconfig) {
		autoconfig := plan.Autoconfig.ValueBool()

		_, err = r.api.UpdateDomain(&tem.UpdateDomainRequest{
			Region:     region,
			DomainID:   id,
			Autoconfig: &autoconfig,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError("Failed to update TEM domain", err.Error())

			return
		}

		timeout, diags := plan.Timeouts.Update(ctx, DefaultDomainCreateTimeout)
		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		err = WaitForDomainAutoconfig(ctx, r.api, region, id, autoconfig, timeout)
		if err != nil {
			resp.Diagnostics.AddError("Failed waiting for TEM domain autoconfig", err.Error())

			return
		}
	}

	domain, err := r.api.GetDomain(&tem.GetDomainRequest{
		Region:   region,
		DomainID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to read TEM domain after update", err.Error())

		return
	}

	newState := flattenDomain(ctx, domain, region, &resp.Diagnostics)
	newState.Timeouts = plan.Timeouts
	newState.AcceptTos = types.BoolValue(true)

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(domain.Region, domain.ID))...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse TEM domain ID", err.Error())

		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, DefaultDomainTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = WaitForDomain(ctx, r.api, region, id, timeout)
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed waiting for TEM domain before delete", err.Error())

		return
	}

	_, err = r.api.RevokeDomain(&tem.RevokeDomainRequest{
		Region:   region,
		DomainID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to revoke TEM domain", err.Error())

		return
	}

	_, err = WaitForDomain(ctx, r.api, region, id, timeout)
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed waiting for TEM domain after revoke", err.Error())
	}
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func flattenDomain(ctx context.Context, domain *tem.Domain, region scw.Region, diags *diag.Diagnostics) domainModel {
	model := domainModel{
		ID:                   types.StringValue(regional.NewIDString(region, domain.ID)),
		Name:                 types.StringValue(domain.Name),
		AcceptTos:            types.BoolValue(true),
		Autoconfig:           types.BoolValue(domain.Autoconfig),
		Status:               types.StringValue(string(domain.Status)),
		CreatedAt:            types.StringValue(scwtypes.FlattenTime(domain.CreatedAt).(string)),
		NextCheckAt:          types.StringValue(scwtypes.FlattenTime(domain.NextCheckAt).(string)),
		LastValidAt:          types.StringValue(scwtypes.FlattenTime(domain.LastValidAt).(string)),
		RevokedAt:            types.StringValue(scwtypes.FlattenTime(domain.RevokedAt).(string)),
		LastError:            types.StringValue(scwtypes.FlattenStringPtr(domain.LastError).(string)), //nolint:staticcheck
		SpfConfig:            types.StringValue(domain.SpfConfig),
		DkimConfig:           types.StringValue(domain.DkimConfig),
		DmarcName:            types.StringValue(""),
		DmarcConfig:          types.StringValue(""),
		DkimName:             types.StringValue(""),
		SpfValue:             types.StringValue(""),
		MxConfig:             types.StringValue(""),
		MxPriority:           types.Int64Value(0),
		MxBlackhole:          types.StringValue(tem.MXBlackhole),
		SmtpHost:             types.StringValue(tem.SMTPHost),
		SmtpPortUnsecure:     types.Int64Value(int64(tem.SMTPPortUnsecure)),
		SmtpPort:             types.Int64Value(int64(tem.SMTPPort)),
		SmtpPortAlternative:  types.Int64Value(int64(tem.SMTPPortAlternative)),
		SmtpsPort:            types.Int64Value(int64(tem.SMTPSPort)),
		SmtpsPortAlternative: types.Int64Value(int64(tem.SMTPSPortAlternative)),
		SmtpsAuthUser:        types.StringValue(domain.ProjectID),
		Reputation:           flattenDomainReputation(ctx, domain.Reputation, diags),
		Region:               types.StringValue(region.String()),
		ProjectID:            types.StringValue(domain.ProjectID),
	}

	if domain.Records != nil {
		if domain.Records.Dmarc != nil {
			model.DmarcName = types.StringValue(domain.Records.Dmarc.Name)
			model.DmarcConfig = types.StringValue(domain.Records.Dmarc.Value)
		}

		if domain.Records.Dkim != nil {
			model.DkimName = types.StringValue(domain.Records.Dkim.Name)
		}

		if domain.Records.Spf != nil {
			model.SpfValue = types.StringValue(domain.Records.Spf.Value)
		}

		if domain.Records.Mx != nil {
			priority, exchange := FlattenMXRecordValue(domain.Records.Mx.Value)
			model.MxConfig = types.StringValue(exchange)
			model.MxPriority = types.Int64Value(int64(priority))

			if exchange != "" {
				model.MxBlackhole = types.StringValue(exchange)
			}
		}
	}

	return model
}
