package iam

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ ephemeral.EphemeralResource              = (*ApiKeyEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*ApiKeyEphemeralResource)(nil)
)

type ApiKeyEphemeralResource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

func NewApiKeyEphemeralResource() ephemeral.EphemeralResource {
	return &ApiKeyEphemeralResource{}
}

func (r *ApiKeyEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client := m.ScwClient()
	r.iamAPI = iam.NewAPI(client)
	r.meta = m
}

func (r *ApiKeyEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_api_key"
}

type ApiKeyEphemeralResourceModel struct {
	Description   types.String `tfsdk:"description"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
	ApplicationID types.String `tfsdk:"application_id"`
	UserID        types.String `tfsdk:"user_id"`
	// Output
	AccessKey        types.String `tfsdk:"access_key"`
	SecretKey        types.String `tfsdk:"secret_key"`
	CreationIP       types.String `tfsdk:"creation_ip"`
	DefaultProjectID types.String `tfsdk:"default_project_id"`
}

//go:embed descriptions/api_key_ephemeral_resource.md
var apiKeyEphemeralResourceDescription string

func (r *ApiKeyEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         apiKeyEphemeralResourceDescription,
		MarkdownDescription: apiKeyEphemeralResourceDescription,
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the iam api key",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the creation of the iam api key",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the last update of the iam api key",
			},
			"expires_at": schema.StringAttribute{
				Description: "The date and time (UTC) of the expiration of the iam api key. Cannot be changed afterwards",
				Optional:    true,
			},
			"access_key": schema.StringAttribute{
				Computed:    true,
				Description: "The access key of the iam api key",
			},
			"secret_key": schema.StringAttribute{
				Computed:    true,
				Description: "The secret Key of the iam api key",
				Sensitive:   true,
			},
			"application_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the application attached to the api key",
				Validators: []validator.String{
					verify.IsStringUUID(),
					stringvalidator.ConflictsWith(path.MatchRoot("user_id")),
				},
			},
			"user_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the user attached to the api key",
				Validators: []validator.String{
					verify.IsStringUUID(),
					stringvalidator.ConflictsWith(path.MatchRoot("application_id")),
				},
			},
			"creation_ip": schema.StringAttribute{
				Computed:    true,
				Description: "The IPv4 Address of the device which created the API key",
			},
			"default_project_id": schema.StringAttribute{
				Description: "Default Project ID to use with Object Storage.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
		},
	}
}

func (r *ApiKeyEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ApiKeyEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if r.iamAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured iamAPI",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	createApiKeyreq := iam.CreateAPIKeyRequest{
		ApplicationID:    data.ApplicationID.ValueStringPointer(),
		UserID:           data.UserID.ValueStringPointer(),
		DefaultProjectID: data.DefaultProjectID.ValueStringPointer(),
		Description:      data.Description.String(),
	}

	var err error

	if !data.ExpiresAt.IsNull() && !data.ExpiresAt.IsUnknown() && data.ExpiresAt.ValueString() != "" {
		parsedExpiresAt, err := time.Parse(time.RFC3339, data.ExpiresAt.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid expires_at value",
				fmt.Sprintf("The start_date attribute must be a valid RFC3339 timestamp. Got %q: %s", data.ExpiresAt.ValueString(), err),
			)

			return
		}

		createApiKeyreq.ExpiresAt = &parsedExpiresAt
	}

	res, err := r.iamAPI.CreateAPIKey(&createApiKeyreq, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing IAM Api Key Create",
			fmt.Sprintf("%s", err),
		)

		return
	}

	data.CreatedAt = types.StringValue(res.CreatedAt.Format(time.RFC3339))
	data.UpdatedAt = types.StringValue(res.UpdatedAt.Format(time.RFC3339))
	data.AccessKey = types.StringValue(res.AccessKey)
	data.SecretKey = types.StringValue(*res.SecretKey)
	data.ExpiresAt = types.StringValue(res.ExpiresAt.Format(time.RFC3339))
	data.CreationIP = types.StringValue(res.CreationIP)
	data.DefaultProjectID = types.StringValue(res.DefaultProjectID)

	resp.Result.Set(ctx, &data)
}
