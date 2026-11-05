package domain

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"strings"

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
	_ resource.Resource                = (*ExternalDomainResource)(nil)
	_ resource.ResourceWithConfigure   = (*ExternalDomainResource)(nil)
	_ resource.ResourceWithImportState = (*ExternalDomainResource)(nil)
	_ resource.ResourceWithIdentity    = (*ExternalDomainResource)(nil)
)

func NewExternalDomainResource() resource.Resource {
	return &ExternalDomainResource{}
}

type ExternalDomainResource struct {
	api  *domainsdk.RegistrarAPI
	meta *meta.Meta
}

type externalDomainModel struct {
	ID              types.String `tfsdk:"id"`
	Domain          types.String `tfsdk:"domain"`
	ProjectID       types.String `tfsdk:"project_id"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	NSServers       types.List   `tfsdk:"ns_servers"`
	Status          types.String `tfsdk:"status"`
	ValidationToken types.String `tfsdk:"validation_token"`
}

type externalDomainIdentityModel = identityfw.GlobalIdentity

//go:embed descriptions/domain_external_resource.md
var externalDomainResourceDescription string

func (r *ExternalDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_external"
}

func (r *ExternalDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: externalDomainResourceDescription,
		Description:         externalDomainResourceDescription,
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
				MarkdownDescription: "The domain name to be managed",
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
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the domain",
			},
			"validation_token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The validation token for the domain",
			},
		},
	}
}

func (r *ExternalDomainResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultGlobal()
}

func (r *ExternalDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExternalDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan externalDomainModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Missing project ID", err.Error())

		return
	}

	domainName := plan.Domain.ValueString()

	registerResp, err := r.api.RegisterExternalDomain(&domainsdk.RegistrarAPIRegisterExternalDomainRequest{
		Domain:    domainName,
		ProjectID: projectID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to register external domain", err.Error())

		return
	}

	state, diags := r.readExternalDomain(ctx, registerResp.Domain)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state == nil {
		resp.Diagnostics.AddError(
			"Failed to read external domain after create",
			"Register succeeded but GetDomain returned not found",
		)

		return
	}

	// Seed token from create response when Get still has an empty one.
	if state.ValidationToken.ValueString() == "" {
		state.ValidationToken = types.StringValue(registerResp.ValidationToken)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetGlobalIdentity(registerResp.Domain))...)
}

func (r *ExternalDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		state    externalDomainModel
		identity externalDomainIdentityModel
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
		resp.Diagnostics.AddError("External domain ID not set", "Cannot refresh domain without an ID")

		return
	}

	readState, diags := r.readExternalDomain(ctx, domainName)
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

func (r *ExternalDomainResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"External domains cannot be updated. Changes to domain or project_id require resource replacement.",
	)
}

func (r *ExternalDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state externalDomainModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	domainName := state.ID.ValueString()
	if domainName == "" {
		domainName = state.Domain.ValueString()
	}

	_, err := r.api.DeleteExternalDomain(&domainsdk.RegistrarAPIDeleteExternalDomainRequest{
		Domain: domainName,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			log.Printf("[WARN] domain external domain (%s) not found, removing from state", domainName)

			return
		}

		// DELETE can return 403 "domain not found" both when the domain is gone and
		// when it is still checking (API refuses delete). Confirm with Get.
		if httperrors.Is403(err) && strings.Contains(err.Error(), "not found") {
			_, getErr := r.api.GetDomain(&domainsdk.RegistrarAPIGetDomainRequest{
				Domain: domainName,
			}, scw.WithContext(ctx))
			if getErr != nil && (httperrors.Is404(getErr) || (httperrors.Is403(getErr) && strings.Contains(getErr.Error(), "not found"))) {
				log.Printf("[WARN] domain external domain (%s) not found, removing from state", domainName)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to delete external domain", err.Error())
	}
}

func (r *ExternalDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

func (r *ExternalDomainResource) readExternalDomain(ctx context.Context, domainName string) (*externalDomainModel, diag.Diagnostics) {
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

	validationToken := ""
	if resp.ExternalDomainRegistrationStatus != nil {
		validationToken = resp.ExternalDomainRegistrationStatus.ValidationToken
	}

	nsServers := []string{}
	if len(resp.DNSZones) > 0 {
		nsServers = resp.DNSZones[0].NsDefault
	}

	state, flattenDiags := flattenExternalDomain(ctx, resp.Domain, resp.OrganizationID, resp.ProjectID, resp.Status.String(), validationToken, nsServers)
	diags.Append(flattenDiags...)

	return &state, diags
}

func flattenExternalDomain(
	ctx context.Context,
	domainName, organizationID, projectID, status, validationToken string,
	nsServers []string,
) (externalDomainModel, diag.Diagnostics) {
	if nsServers == nil {
		nsServers = []string{}
	}

	nsList, diags := types.ListValueFrom(ctx, types.StringType, nsServers)

	return externalDomainModel{
		ID:              types.StringValue(domainName),
		Domain:          types.StringValue(domainName),
		OrganizationID:  types.StringValue(organizationID),
		ProjectID:       types.StringValue(projectID),
		Status:          types.StringValue(status),
		ValidationToken: types.StringValue(validationToken),
		NSServers:       nsList,
	}, diags
}
