package mailbox

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mailboxsdk "github.com/scaleway/scaleway-sdk-go/api/mailbox/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

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
	api  *mailboxsdk.API
	meta *meta.Meta
}

type domainResourceModel struct {
	DNSRecords        types.List   `tfsdk:"dns_records"`
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ProjectID         types.String `tfsdk:"project_id"`
	Status            types.String `tfsdk:"status"`
	WebmailURL        types.String `tfsdk:"webmail_url"`
	ImapURL           types.String `tfsdk:"imap_url"`
	Pop3URL           types.String `tfsdk:"pop3_url"`
	SMTPURL           types.String `tfsdk:"smtp_url"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	MailboxTotalCount types.Int64  `tfsdk:"mailbox_total_count"`
}

type domainResourceIdentityModel = framework.GlobalIdentity

func (r *DomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mailbox_domain"
}

//go:embed descriptions/domain_resource.md
var domainResourceDescription string

func (r *DomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: domainResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the mailbox domain",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Fully qualified domain name (e.g. mail.example.com)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "ID of the project the domain belongs to",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Domain status: creating, waiting_validation, validating, validation_failed, provisioning, ready, deleting",
			},
			"mailbox_total_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of mailboxes provisioned on this domain",
			},
			"webmail_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL of the webmail interface",
			},
			"imap_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "IMAP server URL for email clients",
			},
			"pop3_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "POP3 server URL for email clients",
			},
			"smtp_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SMTP server URL for email clients",
			},
			"dns_records": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "DNS records that must be configured in your DNS zone to validate the domain",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dns_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "DNS record type (e.g. TXT, MX, CNAME)",
						},
						"dns_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Fully qualified DNS name for this record",
						},
						"dns_value": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "DNS record value to set",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Validation status of this record",
						},
						"level": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Requirement level (required, recommended, optional)",
						},
						"error": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Error detail when the record is invalid",
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of domain creation (RFC 3339 format)",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of last update (RFC 3339 format)",
			},
		},
	}
}

func (r *DomainResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = framework.DefaultGlobal()
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
	r.api = mailboxsdk.NewAPI(r.meta.ScwClient())
}

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data domainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Missing project ID", err.Error())

		return
	}

	domain, err := r.api.CreateDomain(&mailboxsdk.CreateDomainRequest{
		ProjectID: projectID,
		Name:      data.Name.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create mailbox domain", err.Error())

		return
	}

	domain, err = waitForDomain(ctx, r.api, domain.ID, defaultDomainTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for mailbox domain", err.Error())

		return
	}

	state := convertDomainToState(ctx, r.api, domain, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(domain.ID))...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    domainResourceModel
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

	domainID := state.ID.ValueString()
	if identityAvailable {
		domainID = identity.ID.ValueString()
	}

	if domainID == "" {
		resp.Diagnostics.AddError("Mailbox domain ID not set", "Cannot refresh domain without an ID")

		return
	}

	domain, err := r.api.GetDomain(&mailboxsdk.GetDomainRequest{DomainID: domainID}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to get mailbox domain", err.Error())

		return
	}

	state = convertDomainToState(ctx, r.api, domain, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, framework.SetGlobalIdentity(domain.ID))...)
}

func (r *DomainResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Mailbox domains cannot be updated. Changes to name or project_id require resource replacement.",
	)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.api.DeleteDomain(&mailboxsdk.DeleteDomainRequest{DomainID: state.ID.ValueString()}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			return
		}

		resp.Diagnostics.AddError("Failed to delete mailbox domain", err.Error())

		return
	}

	_, err = waitForDomain(ctx, r.api, state.ID.ValueString(), defaultDomainTimeout)
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed waiting for mailbox domain deletion", err.Error())
	}
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func convertDomainToState(ctx context.Context, api *mailboxsdk.API, domain *mailboxsdk.Domain, diags *diag.Diagnostics) domainResourceModel {
	state := domainResourceModel{
		ID:                types.StringValue(domain.ID),
		Name:              types.StringValue(domain.Name),
		ProjectID:         types.StringValue(domain.ProjectID),
		Status:            types.StringValue(domain.Status.String()),
		MailboxTotalCount: types.Int64Value(int64(domain.MailboxTotalCount)),
		WebmailURL:        types.StringValue(domain.WebmailURL),
		ImapURL:           types.StringValue(domain.ImapURL),
		Pop3URL:           types.StringValue(domain.Pop3URL),
		SMTPURL:           types.StringValue(domain.SMTPURL),
		CreatedAt:         flattenTime(domain.CreatedAt),
		UpdatedAt:         flattenTime(domain.UpdatedAt),
		DNSRecords:        types.ListNull(types.ObjectType{AttrTypes: dnsRecordAttrTypes()}),
	}

	records, err := api.GetDomainRecords(&mailboxsdk.GetDomainRecordsRequest{DomainID: domain.ID}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			diags.AddWarning(
				"Mailbox domain DNS records unavailable",
				"GetDomainRecords returned 404; dns_records will be empty until records are published.",
			)
		} else {
			diags.AddError("Failed to get domain DNS records", err.Error())
		}

		return state
	}

	state.DNSRecords = flattenDNSRecords(records, diags)

	return state
}
