package keymanager

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ ephemeral.EphemeralResource              = (*DecryptEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*DecryptEphemeralResource)(nil)
)

type DecryptEphemeralResource struct {
	keyManagerAPI *key_manager.API
	meta          *meta.Meta
}

func NewDecryptEphemeralResource() ephemeral.EphemeralResource {
	return &DecryptEphemeralResource{}
}

func (r *DecryptEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (r *DecryptEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_manager_decrypt"
}

type DecryptEphemeralResourceModel struct {
	Region         types.String `tfsdk:"region"`
	KeyID          types.String `tfsdk:"key_id"`
	Plaintext      types.String `tfsdk:"plaintext"`
	AssociatedData types.Object `tfsdk:"associated_data"`
	// Output
	Ciphertext types.String `tfsdk:"ciphertext"`
}

//go:embed descriptions/decrypt_ephemeral_resource.md
var decryptEphemeralResourceDescription string

func (r *DecryptEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         decryptEphemeralResourceDescription,
		MarkdownDescription: decryptEphemeralResourceDescription,
		Attributes: map[string]schema.Attribute{
			"region": regional.SchemaAttribute("Region of the key. If not set, the region is derived from the key_id when possible or from the provider configuration."),
			"key_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the key to use for decryption. Can be a plain UUID or a regional ID.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
			"ciphertext": schema.StringAttribute{
				Required:    true,
				Description: "Ciphertext data to decrypt. Data size must be between 1 and 131071 bytes.",
				Sensitive:   true,
			},
			"associated_data": schema.ObjectAttribute{
				Optional:    true,
				Description: "Must match the associated_data value passed in the encryption request. Only supported by keys with a usage set to `symmetric_encryption`.",
				AttributeTypes: map[string]attr.Type{
					"value": types.StringType,
				},
			},
			"plaintext": schema.StringAttribute{
				Computed:    true,
				Description: "Key's decrypted data.",
				Sensitive:   true,
			},
		},
	}
}

func (r *DecryptEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data DecryptEphemeralResourceModel
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
	ciphertext := data.Ciphertext.ValueString()

	var region scw.Region
	var err error

	if !data.Region.IsNull() && data.Region.ValueString() != "" {
		region = scw.Region(data.Region.ValueString())
	} else {
		// Try to derive region from the key_id if it is a regional ID
		if derivedRegion, id, parseErr := regional.ParseID(keyID); parseErr == nil {
			region = derivedRegion
			keyID = id
		} else {
			// Use default region from provider configuration
			defaultRegion, exists := r.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Missing region",
					"The region attribute is required to decrypt with a key. Please provide it explicitly or configure a default region in the provider.",
				)
				return
			}
			region = defaultRegion
		}
	}

	var associatedData []byte

	if !data.AssociatedData.IsNull() && !data.AssociatedData.IsUnknown() {
		var assocDataModel AssociatedDataModel
		diags := data.AssociatedData.As(ctx, &assocDataModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		associatedData = []byte(assocDataModel.Value.ValueString())
	}

	decryptReq := &key_manager.DecryptRequest{
		Region:         region,
		KeyID:          keyID,
		Ciphertext:     []byte(ciphertext),
		AssociatedData: &associatedData,
	}

	decryptResp, err := r.keyManagerAPI.Decrypt(decryptReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Key Manager decrypt action",
			fmt.Sprintf("%s", err),
		)
		return
	}

	data.Plaintext = types.StringValue(string(decryptResp.Plaintext))

	resp.Result.Set(ctx, &data)
}
