package keymanager

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ ephemeral.EphemeralResource              = (*SignEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*SignEphemeralResource)(nil)
)

type SignEphemeralResource struct {
	keyManagerAPI *key_manager.API
	meta          *meta.Meta
}

func NewSignEphemeralResource() ephemeral.EphemeralResource {
	return &SignEphemeralResource{}
}

func (r *SignEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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
	r.keyManagerAPI = key_manager.NewAPI(client)
	r.meta = m
}

func (r *SignEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_manager_sign"
}

type SignEphemeralResourceModel struct {
	Region types.String `tfsdk:"region"`
	KeyID  types.String `tfsdk:"key_id"`
	Digest types.String `tfsdk:"digest"`
	// Output
	Signature types.String `tfsdk:"signature"`
}

//go:embed descriptions/sign_ephemeral_resource.md
var signEphemeralResourceDescription string

func (r *SignEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         signEphemeralResourceDescription,
		MarkdownDescription: signEphemeralResourceDescription,
		Attributes: map[string]schema.Attribute{
			"region": regional.SchemaAttribute("Region of the key. If not set, the region is derived from the key_id when possible or from the provider configuration."),
			"key_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the key to use for signing. Can be a plain UUID or a regional ID.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
			"digest": schema.StringAttribute{
				Required:    true,
				Description: "Digest of the message to sign. The digest must be generated using the same algorithm defined in the key’s algorithm settings, and encoded as a base64 string.",
				Sensitive:   true,
			},
			"signature": schema.StringAttribute{
				Computed:    true,
				Description: "The message signature, returned as a base64-encoded string.",
			},
		},
	}
}

func (r *SignEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data SignEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if r.keyManagerAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured keymanagerAPI",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	keyID := locality.ExpandID(data.KeyID.ValueString())

	var region scw.Region

	var err error

	if !data.Region.IsNull() && data.Region.ValueString() != "" {
		region = scw.Region(data.Region.ValueString())
	} else {
		if derivedRegion, id, parseErr := regional.ParseID(keyID); parseErr == nil {
			region = derivedRegion
			keyID = id
		} else {
			defaultRegion, exists := r.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Missing region",
					"The region attribute is required to encrypt with a key. Please provide it explicitly or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	digestBytes, err := base64.StdEncoding.DecodeString(data.Digest.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid digest format",
			fmt.Sprintf("Digest must be valid Base64: %v", err),
		)

		return
	}

	signReq := &key_manager.SignRequest{
		Region: region,
		KeyID:  keyID,
		Digest: digestBytes,
	}

	signResp, err := r.keyManagerAPI.Sign(signReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Key Manager Sign",
			fmt.Sprintf("%s", err),
		)

		return
	}

	data.Signature = types.StringValue(secret.Base64Encoded(signResp.Signature))

	resp.Result.Set(ctx, &data)
}
