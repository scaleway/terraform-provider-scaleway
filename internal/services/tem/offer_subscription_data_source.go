package tem

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/tem_offer_subscription_data_source.md
var offerSubscriptionDataSourceDescription string

var (
	_ datasource.DataSource              = (*OfferSubscriptionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*OfferSubscriptionDataSource)(nil)
)

func NewOfferSubscriptionDataSource() datasource.DataSource {
	return &OfferSubscriptionDataSource{}
}

type OfferSubscriptionDataSource struct {
	api  *tem.API
	meta *meta.Meta
}

type offerSubscriptionDataSourceModel struct {
	SLA                          types.Float64 `tfsdk:"sla"`
	CancellationAvailableAt      types.String  `tfsdk:"cancellation_available_at"`
	ProjectID                    types.String  `tfsdk:"project_id"`
	OrganizationID               types.String  `tfsdk:"organization_id"`
	OfferName                    types.String  `tfsdk:"offer_name"`
	SubscribedAt                 types.String  `tfsdk:"subscribed_at"`
	ID                           types.String  `tfsdk:"id"`
	Region                       types.String  `tfsdk:"region"`
	MaxDomains                   types.Int64   `tfsdk:"max_domains"`
	MaxDedicatedIPs              types.Int64   `tfsdk:"max_dedicated_ips"`
	MaxWebhooksPerDomain         types.Int64   `tfsdk:"max_webhooks_per_domain"`
	MaxCustomBlocklistsPerDomain types.Int64   `tfsdk:"max_custom_blocklists_per_domain"`
	IncludedMonthlyEmails        types.Int64   `tfsdk:"included_monthly_emails"`
}

func (d *OfferSubscriptionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tem_offer_subscription"
}

func (d *OfferSubscriptionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	projectID := account.ProjectIDSchemaAttributeOptional()
	projectID.Validators = append(projectID.Validators, stringvalidator.ConflictsWith(path.MatchRoot("organization_id")))

	organizationID := account.OrganizationIDSchemaAttributeOptional("The organization_id the offer subscription is associated to")
	organizationID.Validators = append(organizationID.Validators, stringvalidator.ConflictsWith(path.MatchRoot("project_id")))

	resp.Schema = schema.Schema{
		MarkdownDescription: offerSubscriptionDataSourceDescription,
		Description:         offerSubscriptionDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"region":          regional.SchemaAttributeComputed(),
			"project_id":      projectID,
			"organization_id": organizationID,
			"offer_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the offer",
			},
			"subscribed_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of the subscription",
			},
			"cancellation_available_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Date and time of the end of the offer-subscription commitment",
			},
			"sla": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Service Level Agreement percentage of the offer-subscription",
			},
			"max_domains": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Max number of domains that can be associated with the offer-subscription",
			},
			"max_dedicated_ips": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Max number of dedicated IPs that can be associated with the offer-subscription",
			},
			"max_webhooks_per_domain": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Max number of webhooks that can be associated with the offer-subscription",
			},
			"max_custom_blocklists_per_domain": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Max number of custom blocklists that can be associated with the offer-subscription",
			},
			"included_monthly_emails": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of emails included in the offer-subscription per month",
			},
		},
	}
}

func (d *OfferSubscriptionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = tem.NewAPI(d.meta.ScwClient())
}

func (d *OfferSubscriptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config offerSubscriptionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	listReq := &tem.ListOfferSubscriptionsRequest{
		Region: region,
	}

	if organizationID := config.OrganizationID.ValueString(); organizationID != "" {
		listReq.OrganizationID = scwtypes.ExpandStringPtr(organizationID)
	} else {
		projectID, projectErr := meta.ExtractFrameworkProjectID(config.ProjectID, d.meta.ScwClient())
		if projectErr != nil {
			resp.Diagnostics.AddError("Failed to resolve project ID", projectErr.Error())

			return
		}

		listReq.ProjectID = scwtypes.ExpandStringPtr(projectID)
	}

	offer, err := d.api.ListOfferSubscriptions(listReq, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			resp.Diagnostics.AddError("TEM offer subscription not found", err.Error())

			return
		}

		resp.Diagnostics.AddError("Failed to list TEM offer subscriptions", err.Error())

		return
	}

	if len(offer.OfferSubscriptions) == 0 {
		resp.Diagnostics.AddError(
			"TEM offer subscription not found",
			"no offer subscription found for the given project/organization",
		)

		return
	}

	offerSubscription := offer.OfferSubscriptions[0]

	state := offerSubscriptionDataSourceModel{
		ID:                           types.StringValue(regional.NewIDString(region, offerSubscription.ProjectID)),
		ProjectID:                    types.StringValue(offerSubscription.ProjectID),
		OrganizationID:               types.StringValue(offerSubscription.OrganizationID),
		Region:                       types.StringValue(region.String()),
		OfferName:                    types.StringValue(string(offerSubscription.OfferName)),
		SubscribedAt:                 types.StringValue(scwtypes.FlattenTime(offerSubscription.SubscribedAt).(string)),
		CancellationAvailableAt:      types.StringValue(scwtypes.FlattenTime(offerSubscription.CancellationAvailableAt).(string)),
		SLA:                          types.Float64Value(float64(offerSubscription.SLA)),
		MaxDomains:                   types.Int64Value(int64(offerSubscription.MaxDomains)),
		MaxDedicatedIPs:              types.Int64Value(int64(offerSubscription.MaxDedicatedIPs)),
		MaxWebhooksPerDomain:         types.Int64Value(int64(offerSubscription.MaxWebhooksPerDomain)),
		MaxCustomBlocklistsPerDomain: types.Int64Value(int64(offerSubscription.MaxCustomBlocklistsPerDomain)),
		IncludedMonthlyEmails:        types.Int64Value(int64(offerSubscription.IncludedMonthlyEmails)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
