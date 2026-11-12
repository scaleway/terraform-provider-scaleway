package domain

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	domainsdk "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ resource.Resource                = (*ExternalDomainValidatedResource)(nil)
	_ resource.ResourceWithConfigure   = (*ExternalDomainValidatedResource)(nil)
	_ resource.ResourceWithImportState = (*ExternalDomainValidatedResource)(nil)
	_ resource.ResourceWithIdentity    = (*ExternalDomainValidatedResource)(nil)
)

func NewExternalDomainValidatedResource() resource.Resource {
	return &ExternalDomainValidatedResource{}
}

type ExternalDomainValidatedResource struct {
	api  *domainsdk.RegistrarAPI
	meta *meta.Meta
}

type externalDomainValidatedModel struct {
	ID             types.String   `tfsdk:"id"`
	Domain         types.String   `tfsdk:"domain"`
	ProjectID      types.String   `tfsdk:"project_id"`
	OrganizationID types.String   `tfsdk:"organization_id"`
	NSServers      types.List     `tfsdk:"ns_servers"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
	Validated      types.Bool     `tfsdk:"validated"`
}

type externalDomainValidatedIdentityModel = identityfw.GlobalIdentity

//go:embed descriptions/domain_external_validated_resource.md
var externalDomainValidatedResourceDescription string

func (r *ExternalDomainValidatedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_external_validated"
}

func (r *ExternalDomainValidatedResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: externalDomainValidatedResourceDescription,
		Description:         externalDomainValidatedResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The domain name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain name to be validated",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The project ID the domain belongs to",
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
				MarkdownDescription: "The organization ID the domain is associated with",
			},
			"ns_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of NS servers for the domain",
			},
			"validated": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Indicates if the domain has been validated by Scaleway",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *ExternalDomainValidatedResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultGlobal()
}

func (r *ExternalDomainValidatedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = domainsdk.NewRegistrarAPI(m.ScwClient())
}

func (r *ExternalDomainValidatedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan externalDomainValidatedModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultDomainRegistrationTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainName := plan.Domain.ValueString()

	_, err := waitForExternalDomainValidation(ctx, r.api, domainName, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for external domain validation", err.Error())

		return
	}

	state, diags := r.readExternalDomainValidated(ctx, domainName, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state == nil {
		resp.Diagnostics.AddError(
			"Failed to read external domain after validation",
			"Validation waiter succeeded but GetDomain returned not found",
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetGlobalIdentity(domainName))...)
}

func (r *ExternalDomainValidatedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    externalDomainValidatedModel
		identity externalDomainValidatedIdentityModel
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

	domainName := state.ID.ValueString()
	if identityAvailable {
		domainName = identity.ID.ValueString()
	}

	if domainName == "" {
		domainName = state.Domain.ValueString()
	}

	if domainName == "" {
		resp.Diagnostics.AddError("External domain ID not set", "Cannot refresh validated domain without an ID")

		return
	}

	readState, diags := r.readExternalDomainValidated(ctx, domainName, state.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if readState == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, readState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetGlobalIdentity(readState.ID.ValueString()))...)
}

func (r *ExternalDomainValidatedResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"External domain validation cannot be updated. Changes to domain or project_id require resource replacement.",
	)
}

func (r *ExternalDomainValidatedResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Wait-only resource: destroy is a no-op on the API.
}

func (r *ExternalDomainValidatedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *ExternalDomainValidatedResource) readExternalDomainValidated(
	ctx context.Context,
	domainName string,
	timeoutsValue timeouts.Value,
) (*externalDomainValidatedModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	resp, err := r.api.GetDomain(&domainsdk.RegistrarAPIGetDomainRequest{Domain: domainName}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) || (httperrors.Is403(err) && strings.Contains(err.Error(), "not found")) {
			log.Printf("[WARN] domain external domain (%s) not found, removing from state", domainName)

			return nil, diags
		}

		diags.AddError("Failed to get external domain", err.Error())

		return nil, diags
	}

	nsServers := []string{}
	if len(resp.DNSZones) > 0 {
		nsServers = resp.DNSZones[0].NsDefault
	}

	nsList, listDiags := types.ListValueFrom(ctx, types.StringType, nsServers)
	diags.Append(listDiags...)

	return &externalDomainValidatedModel{
		ID:             types.StringValue(resp.Domain),
		Domain:         types.StringValue(resp.Domain),
		ProjectID:      types.StringValue(resp.ProjectID),
		OrganizationID: types.StringValue(resp.OrganizationID),
		NSServers:      nsList,
		Timeouts:       timeoutsValue,
		Validated:      types.BoolValue(resp.Status == domainsdk.DomainStatusActive),
	}, diags
}
