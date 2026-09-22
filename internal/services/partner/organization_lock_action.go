package partner

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	partner "github.com/scaleway/scaleway-sdk-go/api/partner/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*PartnerOrganizationLockAction)(nil)
	_ action.ActionWithConfigure = (*PartnerOrganizationLockAction)(nil)
)

type PartnerOrganizationLockAction struct {
	partnerAPI *partner.API
}

func (a *PartnerOrganizationLockAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	m, ok := req.ProviderData.(*meta.Meta)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *meta.Meta, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.partnerAPI = partner.NewAPI(meta.ExtractScwClient(m))
}

func (a *PartnerOrganizationLockAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_organization_lock"
}

type PartnerOrganizationLockActionModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
}

func NewPartnerOrganizationLockAction() action.Action {
	return &PartnerOrganizationLockAction{}
}

//go:embed descriptions/organization_lock_action.md
var partnerOrganizationLockActionDescription string

func (a *PartnerOrganizationLockAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: partnerOrganizationLockActionDescription,
		Description:         partnerOrganizationLockActionDescription,
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the organization to lock.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
		},
	}
}

func (a *PartnerOrganizationLockAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data PartnerOrganizationLockActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.partnerAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured partnerAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	if data.OrganizationID.IsNull() || data.OrganizationID.IsUnknown() || data.OrganizationID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing organization_id",
			"The organization_id attribute is required to lock a partner organization.",
		)

		return
	}

	orgID := locality.ExpandID(data.OrganizationID.ValueString())

	_, err := a.partnerAPI.LockOrganization(&partner.LockOrganizationRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing Partner LockOrganization action",
			fmt.Sprintf("Failed to lock organization %s: %s", orgID, err),
		)

		return
	}
}
