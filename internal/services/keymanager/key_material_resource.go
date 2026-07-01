package keymanager

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	key_manager "github.com/scaleway/scaleway-sdk-go/api/key_manager/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/secret"
)

var (
	_ resource.Resource              = (*KeyMaterialResource)(nil)
	_ resource.ResourceWithConfigure = (*KeyMaterialResource)(nil)
)

func NewKeyMaterialResource() resource.Resource {
	return &KeyMaterialResource{}
}

type KeyMaterialResource struct {
	keyManagerAPI *key_manager.API
	meta          *meta.Meta
}

type keyMaterialResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	KeyID                types.String `tfsdk:"key_id"`
	Region               types.String `tfsdk:"region"`
	KeyMaterial          types.String `tfsdk:"key_material"`
	KeyMaterialWo        types.String `tfsdk:"key_material_wo"`
	KeyMaterialWoVersion types.Int64  `tfsdk:"key_material_wo_version"`
	Salt                 types.String `tfsdk:"salt"`
	SaltWo               types.String `tfsdk:"salt_wo"`
	SaltWoVersion        types.Int64  `tfsdk:"salt_wo_version"`
	KeyState             types.String `tfsdk:"key_state"`
	Origin               types.String `tfsdk:"origin"`
}

func (r *KeyMaterialResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_manager_key_material"
}

//go:embed descriptions/key_material_resource.md
var keyMaterialResourceDescription string

func (r *KeyMaterialResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: keyMaterialResourceDescription,
		Description:         keyMaterialResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the key material resource (same as key_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the key to import key material into. The key's origin must be external (UUID format). Can be a plain UUID or a regional ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Region of the key. If not set, the region is derived from the key_id when possible or from the provider configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_material": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The key material to import. The key material is a random sequence of bytes used to derive a cryptographic key. Can be provided as raw bytes or a base64-encoded string (the provider will automatically normalize the input).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("key_material_wo"),
					),
				},
			},
			"key_material_wo": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "The key material to import in write-only mode. The key material is a random sequence of bytes used to derive a cryptographic key. Can be provided as raw bytes or a base64-encoded string (the provider will automatically normalize the input). The key material will not be stored in the Terraform state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("key_material"),
					),
				},
			},
			"salt": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Optional salt for key derivation. A salt is random data added to key material to ensure unique derived keys, even if the input is similar. It helps strengthen security when the key material has low randomness (low entropy). Can be provided as raw bytes or a base64-encoded string (the provider will automatically normalize the input).",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("salt_wo"),
					),
				},
			},
			"salt_wo": schema.StringAttribute{
				Optional:            true,
				WriteOnly:           true,
				MarkdownDescription: "Optional salt for key derivation in write-only mode. A salt is random data added to key material to ensure unique derived keys. Can be provided as raw bytes or a base64-encoded string (the provider will automatically normalize the input). The salt will not be stored in the Terraform state.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRoot("salt"),
					),
				},
			},
			"salt_wo_version": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Version number to track changes to the write-only salt. Increment this value to recreate the resource with new salt. Required when using 'salt_wo'.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"key_material_wo_version": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Version number to track changes to the write-only key material. Increment this value to trigger resource recreation. Required when using 'key_material_wo'.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"key_state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The current state of the key (enabled, disabled, pending_key_material).",
			},
			"origin": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The origin of the key (should be 'external').",
			},
		},
	}
}

func (r *KeyMaterialResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.keyManagerAPI = key_manager.NewAPI(r.meta.ScwClient())
}

func (r *KeyMaterialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data keyMaterialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	keyID := locality.ExpandID(data.KeyID.ValueString())

	var region scw.Region
	var err error

	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		region = scw.Region(data.Region.ValueString())
	} else {
		if derivedRegion, id, parseErr := regional.ParseID(data.KeyID.ValueString()); parseErr == nil {
			region = derivedRegion
			keyID = id
		} else {
			defaultRegion, exists := r.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Missing region",
					"The region attribute is required to import key material. Please provide it explicitly or configure a default region in the provider.",
				)
				return
			}
			region = defaultRegion
		}
	}

	hasKeyMaterial := !data.KeyMaterial.IsNull() && !data.KeyMaterial.IsUnknown()
	var keyMaterialBytes []byte
	if hasKeyMaterial {
		normalized := secret.Base64Encoded([]byte(data.KeyMaterial.ValueString()))
		keyMaterialBytes, err = base64.StdEncoding.DecodeString(normalized)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid key_material format",
				fmt.Sprintf("key_material must be either raw bytes or a valid base64-encoded string: %s", err),
			)
			return
		}
	} else {
		normalized := secret.Base64Encoded([]byte(data.KeyMaterialWo.ValueString()))
		keyMaterialBytes, err = base64.StdEncoding.DecodeString(normalized)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid key_material_wo format",
				fmt.Sprintf("key_material_wo must be either raw bytes or a valid base64-encoded string: %s", err),
			)
			return
		}
	}

	importReq := &key_manager.ImportKeyMaterialRequest{
		Region:      region,
		KeyID:       keyID,
		KeyMaterial: keyMaterialBytes,
	}

	if !data.SaltWo.IsNull() && !data.SaltWo.IsUnknown() {
		if data.SaltWoVersion.IsNull() || data.SaltWoVersion.IsUnknown() {
			resp.Diagnostics.AddError(
				"Missing salt_wo_version",
				"salt_wo_version is required when using salt_wo",
			)
			return
		}
		normalized := secret.Base64Encoded([]byte(data.SaltWo.ValueString()))
		saltBytes, err := base64.StdEncoding.DecodeString(normalized)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid salt_wo format",
				fmt.Sprintf("salt_wo must be either raw bytes or a valid base64-encoded string: %s", err),
			)
			return
		}
		importReq.Salt = &saltBytes
	} else if !data.Salt.IsNull() && !data.Salt.IsUnknown() {
		normalized := secret.Base64Encoded([]byte(data.Salt.ValueString()))
		saltBytes, err := base64.StdEncoding.DecodeString(normalized)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid salt format",
				fmt.Sprintf("salt must be either raw bytes or a valid base64-encoded string: %s", err),
			)
			return
		}
		importReq.Salt = &saltBytes
	}

	_, err = r.keyManagerAPI.ImportKeyMaterial(importReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import key material",
			fmt.Sprintf("Failed to import key material: %s", err),
		)
		return
	}

	data.ID = types.StringValue(regional.NewIDString(region, keyID))
	data.Region = types.StringValue(region.String())
	data.KeyID = types.StringValue(regional.NewIDString(region, keyID))

	state := r.readKeyState(region, keyID, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KeyMaterialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state keyMaterialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, keyID, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse ID",
			fmt.Sprintf("Failed to parse resource ID: %s", err),
		)
		return
	}

	newState := r.readKeyState(region, keyID, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *KeyMaterialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Key material cannot be updated.",
	)
}

func (r *KeyMaterialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state keyMaterialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, keyID, err := regional.ParseID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse ID",
			fmt.Sprintf("Failed to parse resource ID: %s", err),
		)
		return
	}

	deleteReq := &key_manager.DeleteKeyMaterialRequest{
		Region: region,
		KeyID:  keyID,
	}

	err = r.keyManagerAPI.DeleteKeyMaterial(deleteReq)
	if err != nil {
		if httperrors.Is404(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete key material",
			fmt.Sprintf("Failed to delete key material: %s", err),
		)
		return
	}
}

func (r *KeyMaterialResource) readKeyState(region scw.Region, keyID string, data *keyMaterialResourceModel, diags *diag.Diagnostics) keyMaterialResourceModel {
	key, err := r.keyManagerAPI.GetKey(&key_manager.GetKeyRequest{
		Region: region,
		KeyID:  keyID,
	})
	if err != nil {
		if httperrors.Is404(err) {
			diags.AddError("Key not found", "The key was not found")
			return *data
		}
		diags.AddError(
			"Failed to read key",
			fmt.Sprintf("Failed to read key: %s", err),
		)
		return *data
	}

	return keyMaterialResourceModel{
		ID:                   types.StringValue(regional.NewIDString(region, keyID)),
		KeyID:                types.StringValue(regional.NewIDString(region, keyID)),
		Region:               types.StringValue(region.String()),
		KeyMaterial:          data.KeyMaterial,
		KeyMaterialWo:        data.KeyMaterialWo,
		KeyMaterialWoVersion: data.KeyMaterialWoVersion,
		Salt:                 data.Salt,
		SaltWo:               data.SaltWo,
		SaltWoVersion:        data.SaltWoVersion,
		KeyState:             types.StringValue(key.State.String()),
		Origin:               types.StringValue(key.Origin.String()),
	}
}
