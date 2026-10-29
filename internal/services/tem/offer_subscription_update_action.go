package tem

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

var (
	_ action.Action              = (*OfferSubscriptionUpdateAction)(nil)
	_ action.ActionWithConfigure = (*OfferSubscriptionUpdateAction)(nil)
)

type OfferSubscriptionUpdateAction struct {
	temAPI *tem.API
	meta   *meta.Meta
}

func (a *OfferSubscriptionUpdateAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
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

	a.meta = m
	a.temAPI = tem.NewAPI(meta.ExtractScwClient(m))
}

func (a *OfferSubscriptionUpdateAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tem_update_offer_subscription"
}

type OfferSubscriptionUpdateActionModel struct {
	OfferName types.String `tfsdk:"offer_name"`
	ProjectID types.String `tfsdk:"project_id"`
	Region    types.String `tfsdk:"region"`
}

func NewOfferSubscriptionUpdateAction() action.Action {
	return &OfferSubscriptionUpdateAction{}
}

//go:embed descriptions/offer_subscription_update_action.md
var offerSubscriptionUpdateActionDescription string

func (a *OfferSubscriptionUpdateAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: offerSubscriptionUpdateActionDescription,
		Description:         offerSubscriptionUpdateActionDescription,
		Attributes: map[string]schema.Attribute{
			"offer_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the offer to subscribe to (`essential` or `scale`)",
				Validators: []validator.String{
					verify.FrameworkValidateEnum[tem.OfferName](),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the project. Defaults to the provider project_id.",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"region": regional.SchemaAttribute(),
		},
	}
}

func (a *OfferSubscriptionUpdateAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data OfferSubscriptionUpdateActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.temAPI == nil || a.meta == nil {
		resp.Diagnostics.AddError(
			"Unconfigured temAPI",
			"The action was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	client := meta.ExtractScwClient(a.meta)

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Missing project_id",
			"Could not determine project_id. Please set the project_id attribute or configure a default project_id in the provider.",
		)

		return
	}

	region, err := meta.ExtractFrameworkRegion(data.Region, client)
	if err != nil {
		resp.Diagnostics.AddError(
			"Missing region",
			"Could not determine region. Please set the region attribute or configure a default region in the provider.",
		)

		return
	}

	offerName := tem.OfferName(data.OfferName.ValueString())

	_, err = a.temAPI.UpdateOfferSubscription(&tem.UpdateOfferSubscriptionRequest{
		Region:    region,
		ProjectID: projectID,
		Name:      offerName,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error executing TEM UpdateOfferSubscription action",
			fmt.Sprintf("Failed to update offer subscription for project %s to %q: %s", projectID, offerName, err),
		)

		return
	}
}
