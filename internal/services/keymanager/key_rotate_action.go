package keymanager

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
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
	_ action.Action              = (*RotateKeyAction)(nil)
	_ action.ActionWithConfigure = (*RotateKeyAction)(nil)
)

type RotateKeyAction struct {
	keyManagerAPI *key_manager.API
	meta          *meta.Meta
}

func (a *RotateKeyAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *scw.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client := m.ScwClient()
	a.keyManagerAPI = key_manager.NewAPI(client)
	a.meta = m
}

func (a *RotateKeyAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_manager_rotate_key"
}

type RotateKeyActionModel struct {
	Region types.String `tfsdk:"region"`
	KeyID  types.String `tfsdk:"key_id"`
}

func NewRotateKeyAction() action.Action {
	return &RotateKeyAction{}
}

//go:embed descriptions/rotate_key_action.md
var rotateKeyActionDescription string

func (a *RotateKeyAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: rotateKeyActionDescription,
		Description:         rotateKeyActionDescription,
		Attributes: map[string]schema.Attribute{
			"region": regional.SchemaAttribute("Region of the key. If not set, the region is derived from the key_id when possible or from the provider configuration."),
			"key_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the key to rotate. Can be a plain UUID or a regional ID.",
				Validators: []validator.String{
					verify.IsStringUUIDOrUUIDWithLocality(),
				},
			},
		},
	}
}

func (a *RotateKeyAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data RotateKeyActionModel
	// Read action config data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.keyManagerAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured keymanagerAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	keyID := locality.ExpandID(data.KeyID.ValueString())

	var (
		region scw.Region
		err    error
	)

	if !data.Region.IsNull() && data.Region.ValueString() != "" {
		region = scw.Region(data.Region.ValueString())
	} else {
		// Try to derive region from the job_definition_id if it is a regional ID.
		if derivedRegion, id, parseErr := regional.ParseID(data.KeyID.ValueString()); parseErr == nil {
			region = derivedRegion
			keyID = id
		} else {
			// Use default region from provider configuration
			defaultRegion, exists := a.meta.ScwClient().GetDefaultRegion()
			if !exists {
				resp.Diagnostics.AddError(
					"Missing region",
					"The region attribute is required to rotate a key. Please provide it explicitly or configure a default region in the provider.",
				)

				return
			}

			region = defaultRegion
		}
	}

	rotateReq := &key_manager.RotateKeyRequest{
		Region: region,
		KeyID:  keyID,
	}

	_, err = a.keyManagerAPI.RotateKey(rotateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Key Manager RotateKey action",
			fmt.Sprintf("%s", err))
	}
}
