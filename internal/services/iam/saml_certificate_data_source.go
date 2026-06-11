package iam

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ datasource.DataSource              = (*SamlCertificateDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*SamlCertificateDataSource)(nil)
)

func NewSamlCertificateDataSource() datasource.DataSource {
	return &SamlCertificateDataSource{}
}

type SamlCertificateDataSource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

type samlCertificateDatasourceModel struct {
	CertificateID types.String `tfsdk:"certificate_id"`
	// Output
	Content   types.String `tfsdk:"content"`
	Type      types.String `tfsdk:"type"`
	Origin    types.String `tfsdk:"origin"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}

func (d *SamlCertificateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_saml_certificate"
}

//go:embed descriptions/saml_certificate_data_source.md
var samlCertificateDataSourceDescription string

func (d *SamlCertificateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: samlCertificateDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"certificate_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the SAML certificate",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the SAML certificate. Possible values are: `signing`, `encryption`.",
				Computed:            true,
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The origin of the SAML certificate. Possible values are: `scaleway`, `identity_provider`.",
				Computed:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The content of the SAML certificate",
				Computed:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The expiration date and time of the SAML certificate",
				Computed:            true,
			},
		},
	}
}

func (d *SamlCertificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.iamAPI = iam.NewAPI(d.meta.ScwClient())
}

func (d *SamlCertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state samlCertificateDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.iamAPI.GetSamlCertificate(&iam.GetSamlCertificateRequest{
		CertificateID: state.CertificateID.ValueString(),
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to list SAML certificates",
			err.Error(),
		)

		return
	}

	state = d.convertToState(res)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *SamlCertificateDataSource) convertToState(cert *iam.SamlCertificate) samlCertificateDatasourceModel {
	state := samlCertificateDatasourceModel{
		CertificateID: types.StringValue(cert.ID),
		Content:       types.StringValue(cert.Content),
		Type:          types.StringValue(string(cert.Type)),
		Origin:        types.StringValue(string(cert.Origin)),
	}

	if cert.ExpiresAt != nil {
		state.ExpiresAt = types.StringValue(cert.ExpiresAt.String())
	}

	return state
}
