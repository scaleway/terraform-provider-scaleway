package iam

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ ephemeral.EphemeralResource              = (*ApiKeyEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*ApiKeyEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithClose     = (*ApiKeyEphemeralResource)(nil)
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
	Description      types.String `tfsdk:"description"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	ExpiresAt        types.String `tfsdk:"expires_at"`
	ApplicationID    types.String `tfsdk:"application_id"`
	UserID           types.String `tfsdk:"user_id"`
	AccessKey        types.String `tfsdk:"access_key"`
	SecretKey        types.String `tfsdk:"secret_key"`
	CreationIP       types.String `tfsdk:"creation_ip"`
	DefaultProjectID types.String `tfsdk:"default_project_id"`
	DeleteOnClose    types.Bool   `tfsdk:"delete_on_close"`
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
				Computed:    true,
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
			"delete_on_close": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the API key should be deleted when the ephemeral resource is closed. Defaults to true. When set to false, the key is not deleted on close and must be managed manually (or let to expire via `expires_at`).",
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
		Description:      data.Description.ValueString(),
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
	if data.ExpiresAt.IsNull() && res.ExpiresAt != nil {
		data.ExpiresAt = types.StringValue(res.ExpiresAt.UTC().Format(time.RFC3339))
	}

	data.CreationIP = types.StringValue(res.CreationIP)
	if data.DefaultProjectID.IsNull() {
		data.DefaultProjectID = types.StringValue(res.DefaultProjectID)
	}

	accessKeyBytes, err := json.Marshal(res.AccessKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error encoding private state",
			fmt.Sprintf("Unable to encode access key in private state: %s", err),
		)

		return
	}

	resp.Private.SetKey(ctx, "access_key", accessKeyBytes)

	deleteOnClose := true
	if !data.DeleteOnClose.IsNull() {
		deleteOnClose = data.DeleteOnClose.ValueBool()
	}

	data.DeleteOnClose = types.BoolValue(deleteOnClose)
	resp.Private.SetKey(ctx, "delete_on_close", []byte(strconv.FormatBool(deleteOnClose)))

	resp.Result.Set(ctx, &data)
}

func (r *ApiKeyEphemeralResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if r.iamAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured iamAPI",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	deleteOnClose, diags := req.Private.GetKey(ctx, "delete_on_close")
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if string(deleteOnClose) == "false" {
		return
	}

	accessKey, diags := req.Private.GetKey(ctx, "access_key")
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(accessKey) == 0 {
		resp.Diagnostics.AddError(
			"Missing access key in private state",
			"The access key was not found in the private state. This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	var accessKeyStr string
	if err := json.Unmarshal(accessKey, &accessKeyStr); err != nil {
		resp.Diagnostics.AddError(
			"Error decoding private state",
			fmt.Sprintf("Unable to decode access key from private state: %s", err),
		)

		return
	}

	err := r.iamAPI.DeleteAPIKey(&iam.DeleteAPIKeyRequest{
		AccessKey: accessKeyStr,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError(
			"Error executing IAM Api Key Delete",
			fmt.Sprintf("%s", err),
		)
	}
}
