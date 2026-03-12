package keymanager

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ ephemeral.EphemeralResource              = (*GenerateDataKeyEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*GenerateDataKeyEphemeralResource)(nil)
)

type GenerateDataKeyEphemeralResource struct {
	keyManagerAPI *key_manager.API
	meta          *meta.Meta
}

func NewGenerateDataKeyEphemeralResource() ephemeral.EphemeralResource {
	return &GenerateDataKeyEphemeralResource{}
}

func (r *GenerateDataKeyEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (r *GenerateDataKeyEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_manager_generate_data_key"
}

type GenerateDataKeyEphemeralResourceModel struct {
	KeyID            types.String `tfsdk:"key_id"`
	Region           types.String `tfsdk:"region"`
	Algorithm        types.String `tfsdk:"algorithm"`
	Ciphertext       types.String `tfsdk:"ciphertext"`
	Plaintext        types.String `tfsdk:"plaintext"`
	CreatedAt        types.String `tfsdk:"created_at"`
	WithoutPlaintext types.Bool   `tfsdk:"without_plaintext"`
}

//go:embed descriptions/generate_data_key_ephemeral_resource.md
var generateDataKeyEphemeralResourceDescription string

func (r *GenerateDataKeyEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         generateDataKeyEphemeralResourceDescription,
		MarkdownDescription: generateDataKeyEphemeralResourceDescription,
		Attributes: map[string]schema.Attribute{
			"region": regional.SchemaAttribute("Region of the key. If not set, the region is derived from the key_id when possible or from the provider configuration."),
			"key_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the key. Can be a plain UUID or a regional ID.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
			"without_plaintext": schema.BoolAttribute{
				Optional:    true,
				Description: "Defines whether to return the data encryption key's plaintext in the response object. Default value is false, meaning that the plaintext is returned. Set it to true if you do not wish the plaintext to be returned in the response object.",
			},
			"algorithm": schema.StringAttribute{
				Computed:    true, // Only aes_256_gcm is supported by Key Manager API for now.
				Description: "Algorithm with which the data encryption key will be used to encrypt and decrypt arbitrary payloads (AES-256-GCM).",
			},
			"ciphertext": schema.StringAttribute{
				Computed:    true,
				Description: "Data encryption key ciphertext. Your data encryption key's ciphertext can be stored safely. It can only be decrypted through the keys you create in Key Manager, using the relevant key ID.",
			},
			"plaintext": schema.StringAttribute{
				Computed:    true,
				Description: "Data encryption key plaintext. Your data encryption key's plaintext allows you to use the key immediately upon creation. It must neither be stored or shared.",
				Sensitive:   true,
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Data encryption key creation date. (RFC 3339 format)",
			},
		},
	}
}

func (r *GenerateDataKeyEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data GenerateDataKeyEphemeralResourceModel
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

	generateDataKeyReq := &key_manager.GenerateDataKeyRequest{
		Region:           region,
		KeyID:            keyID,
		Algorithm:        key_manager.DataKeyAlgorithmSymmetricEncryptionAes256Gcm,
		WithoutPlaintext: data.WithoutPlaintext.ValueBool(),
	}

	generateDataKeyResp, err := r.keyManagerAPI.GenerateDataKey(generateDataKeyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Key Manager Generate Data Key",
			fmt.Sprintf("%s", err),
		)

		return
	}

	data.Algorithm = types.StringValue(string(generateDataKeyResp.Algorithm))

	var plaintext string

	if generateDataKeyResp.Plaintext != nil {
		plaintext = string(*generateDataKeyResp.Plaintext)
	}

	data.Plaintext = types.StringValue(plaintext)
	data.Ciphertext = types.StringValue(string(generateDataKeyResp.Ciphertext))
	data.CreatedAt = types.StringValue(generateDataKeyResp.CreatedAt.Format(time.RFC3339))

	resp.Result.Set(ctx, &data)
}
