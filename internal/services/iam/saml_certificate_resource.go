package iam

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ resource.Resource                = (*SamlCertificateResource)(nil)
	_ resource.ResourceWithConfigure   = (*SamlCertificateResource)(nil)
	_ resource.ResourceWithImportState = (*SamlCertificateResource)(nil)
)

func NewSamlCertificateResource() resource.Resource {
	return &SamlCertificateResource{}
}

type SamlCertificateResource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

type samlCertificateResourceModel struct {
	SamlID         types.String `tfsdk:"saml_id"`
	Content        types.String `tfsdk:"content"`
	OrganizationID types.String `tfsdk:"organization_id"`
	// Output
	ID        types.String `tfsdk:"id"`
	Type      types.String `tfsdk:"type"`
	Origin    types.String `tfsdk:"origin"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}

func (r *SamlCertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_saml_certificate"
}

//go:embed descriptions/saml_certificate_resource.md
var samlCertificateResourceDescription string

func (r *SamlCertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: samlCertificateResourceDescription,
		Attributes: map[string]schema.Attribute{
			"saml_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the SAML configuration",
				Optional:            true,
				Computed:            true, // We allow retrieving the saml_id from the organization_id for easier resource import
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The content of the SAML certificate",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SAML certificate",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the SAML certificate",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The origin of the SAML certificate",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The expiration date and time of the SAML certificate",
				Computed:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *SamlCertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.iamAPI = iam.NewAPI(r.meta.ScwClient())
}

func (r *SamlCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data samlCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgID := data.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			orgID = defaultOrgID
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("organization_id"),
				"Organization ID is required",
				"Either set organization_id or configure a default organization",
			)

			return
		}
	}

	samlID := data.SamlID.ValueString()
	if samlID == "" {
		saml, err := r.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
			OrganizationID: orgID,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to retrieve SAML config",
				err.Error(),
			)
		}

		samlID = saml.ID
	}

	res, err := r.iamAPI.AddSamlCertificate(&iam.AddSamlCertificateRequest{
		SamlID:  samlID,
		Type:    iam.SamlCertificateType(data.Type.ValueString()),
		Content: data.Content.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to add SAML certificate",
			err.Error(),
		)

		return
	}

	state := r.convertToState(res, orgID, samlID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SamlCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state samlCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	orgID := state.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := r.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			orgID = defaultOrgID
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("organization_id"),
				"Organization ID is required",
				"Either set organization_id or configure a default organization",
			)

			return
		}
	}

	samlID := state.SamlID.ValueString()
	if samlID == "" {
		saml, err := r.iamAPI.GetOrganizationSaml(&iam.GetOrganizationSamlRequest{
			OrganizationID: orgID,
		}, scw.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to retrieve SAML config",
				err.Error(),
			)
		}

		samlID = saml.ID
	}

	res, err := r.iamAPI.ListSamlCertificates(&iam.ListSamlCertificatesRequest{
		SamlID: samlID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list SAML certificates",
			err.Error(),
		)

		return
	}

	var foundCert *iam.SamlCertificate

	for _, cert := range res.Certificates {
		if cert.ID == state.ID.ValueString() {
			foundCert = cert

			break
		}
	}

	if foundCert == nil {
		resp.State.RemoveResource(ctx)

		return
	}

	state = r.convertToState(foundCert, orgID, samlID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SamlCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"SAML certificates cannot be updated.",
	)
}

func (r *SamlCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state samlCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.iamAPI.DeleteSamlCertificate(&iam.DeleteSamlCertificateRequest{
		CertificateID: state.ID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete SAML certificate",
			err.Error(),
		)

		return
	}
}

func (r *SamlCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SamlCertificateResource) convertToState(cert *iam.SamlCertificate, orgID string, samlID string) samlCertificateResourceModel {
	state := samlCertificateResourceModel{
		SamlID:         types.StringValue(samlID),
		Content:        types.StringValue(cert.Content),
		ID:             types.StringValue(cert.ID),
		Type:           types.StringValue(string(cert.Type)),
		Origin:         types.StringValue(string(cert.Origin)),
		OrganizationID: types.StringValue(orgID),
	}

	if cert.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(cert.ExpiresAt.String())
	}

	return state
}
