package webhosting

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	webhosting "github.com/scaleway/scaleway-sdk-go/api/webhosting/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	scwds "github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ datasource.DataSource              = (*OfferDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*OfferDataSource)(nil)
)

func NewOfferDataSource() datasource.DataSource {
	return &OfferDataSource{}
}

type OfferDataSource struct {
	api  *webhosting.OfferAPI
	meta *meta.Meta
}

type offerDataSourceModel struct {
	Name                 types.String `tfsdk:"name"`
	ControlPanel         types.String `tfsdk:"control_panel"`
	OfferID              types.String `tfsdk:"offer_id"`
	ID                   types.String `tfsdk:"id"`
	BillingOperationPath types.String `tfsdk:"billing_operation_path"`
	Product              types.List   `tfsdk:"product"`
	Offer                types.List   `tfsdk:"offer"`
	Price                types.String `tfsdk:"price"`
	Region               types.String `tfsdk:"region"`
}

type offerDataSourceOfferModel struct {
	Options              types.List   `tfsdk:"options"`
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	BillingOperationPath types.String `tfsdk:"billing_operation_path"`
	ControlPanelName     types.String `tfsdk:"control_panel_name"`
	QuotaWarning         types.String `tfsdk:"quota_warning"`
	Price                types.String `tfsdk:"price"`
	Available            types.Bool   `tfsdk:"available"`
	EndOfLife            types.Bool   `tfsdk:"end_of_life"`
}

type offerDataSourceOptionModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	BillingOperationPath types.String `tfsdk:"billing_operation_path"`
	QuotaWarning         types.String `tfsdk:"quota_warning"`
	Price                types.String `tfsdk:"price"`
	MinValue             types.Int64  `tfsdk:"min_value"`
	CurrentValue         types.Int64  `tfsdk:"current_value"`
	MaxValue             types.Int64  `tfsdk:"max_value"`
}

func (d *OfferDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhosting_offer"
}

func (d *OfferDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		MarkdownDescription: "Finds a Scaleway WebHosting offer by name or ID.",
		Attributes: map[string]fwschema.Attribute{
			"name": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exact name of the desired offer",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("offer_id")),
				},
			},
			"control_panel": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Name of the control panel (Cpanel or Plesk)",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("offer_id")),
				},
			},
			"offer_id": fwschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the desired offer",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("name"), path.MatchRoot("control_panel")),
				},
			},
			"id": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The regional ID of the offer",
			},
			"billing_operation_path": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Billing operation path",
			},
			"product": fwschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Product list. Deprecated: use the `offer` field instead.",
				NestedObject: fwschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"name": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the product",
						},
						"option": fwschema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether or not the product is in the offer",
						},
						"email_accounts_quota": fwschema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Limit number of email accounts",
						},
						"email_storage_quota": fwschema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Limit quantity of email storage in gigabytes",
						},
						"databases_quota": fwschema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Limit number of databases",
						},
						"hosting_storage_quota": fwschema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Limit quantity of hosting storage in gigabytes",
						},
						"support_included": fwschema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether or not support is included",
						},
						"v_cpu": fwschema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Limit number of virtual CPU",
						},
						"ram": fwschema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Limit quantity of memory in gigabytes",
						},
					},
				},
			},
			"offer": fwschema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The offer details of the hosting",
				NestedObject: fwschema.NestedAttributeObject{
					Attributes: map[string]fwschema.Attribute{
						"id": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the offer",
						},
						"name": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the offer",
						},
						"billing_operation_path": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Billing operation path",
						},
						"available": fwschema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether or not the offer is available",
						},
						"control_panel_name": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the control panel",
						},
						"end_of_life": fwschema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether this offer is in end of life",
						},
						"quota_warning": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Quota warning for this offer",
						},
						"price": fwschema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Price of the offer",
						},
						"options": fwschema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "List of options available for a particular offer",
							NestedObject: fwschema.NestedAttributeObject{
								Attributes: map[string]fwschema.Attribute{
									"id": fwschema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "ID of the option",
									},
									"name": fwschema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Name of the option",
									},
									"billing_operation_path": fwschema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Billing operation path",
									},
									"min_value": fwschema.Int64Attribute{
										Computed:            true,
										MarkdownDescription: "Minimum value for this option",
									},
									"current_value": fwschema.Int64Attribute{
										Computed:            true,
										MarkdownDescription: "Current value",
									},
									"max_value": fwschema.Int64Attribute{
										Computed:            true,
										MarkdownDescription: "The maximum allowed value of this option",
									},
									"quota_warning": fwschema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Quota warning",
									},
									"price": fwschema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Price of the option",
									},
								},
							},
						},
					},
				},
			},
			"price": fwschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Price of the offer",
			},
			"region": fwschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The region of the offer",
			},
		},
	}
}

func (d *OfferDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.api = webhosting.NewOfferAPI(d.meta.ScwClient())
}

func (d *OfferDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config offerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(config.Region, d.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	listResp, err := d.api.ListOffers(&webhosting.OfferAPIListOffersRequest{
		Region: region,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Failed to list webhosting offers", err.Error())

		return
	}

	if len(listResp.Offers) == 0 {
		resp.Diagnostics.AddError("No offers found", fmt.Sprintf("no offer found in region %s", region))

		return
	}

	var filteredOffer *webhosting.Offer

	offerID := config.OfferID.ValueString()
	name := config.Name.ValueString()
	controlPanel := config.ControlPanel.ValueString()

	for _, offer := range listResp.Offers {
		if offer.ID == offerID || (offer.Name == name && (controlPanel == "" || strings.EqualFold(offer.ControlPanelName, controlPanel))) {
			filteredOffer = offer

			break
		}
	}

	if filteredOffer == nil {
		resp.Diagnostics.AddError(
			"Offer not found",
			fmt.Sprintf("no offer found with the name or id: %s %s in region %s", name, offerID, region),
		)

		return
	}

	regionalID := scwds.NewRegionalID(filteredOffer.ID, region)
	state := offerDataSourceModel{
		ID:                   types.StringValue(regionalID),
		OfferID:              types.StringValue(regionalID),
		Name:                 types.StringValue(filteredOffer.Name),
		Region:               types.StringValue(region.String()),
		BillingOperationPath: types.StringValue(filteredOffer.BillingOperationPath),
		Price:                types.StringValue(flattenOfferPrice(filteredOffer.Price)),
		Product:              types.ListNull(types.ObjectType{AttrTypes: offerProductAttrTypes()}),
		Offer:                flattenOfferDataSource(ctx, filteredOffer, &resp.Diagnostics),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenOfferDataSource(ctx context.Context, offer *webhosting.Offer, diags *diag.Diagnostics) types.List {
	if offer == nil {
		return types.ListNull(types.ObjectType{AttrTypes: offerOfferAttrTypes()})
	}

	model := offerDataSourceOfferModel{
		ID:                   types.StringValue(offer.ID),
		Name:                 types.StringValue(offer.Name),
		BillingOperationPath: types.StringValue(offer.BillingOperationPath),
		Available:            types.BoolValue(offer.Available),
		ControlPanelName:     types.StringValue(offer.ControlPanelName),
		EndOfLife:            types.BoolValue(offer.EndOfLife),
		QuotaWarning:         types.StringValue(string(offer.QuotaWarning)),
		Price:                types.StringValue(flattenOfferPrice(offer.Price)),
		Options:              flattenOfferOptionsDataSource(ctx, offer.Options, diags),
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: offerOfferAttrTypes()}, []offerDataSourceOfferModel{model})
	diags.Append(d...)

	return list
}

func flattenOfferOptionsDataSource(ctx context.Context, options []*webhosting.OfferOption, diags *diag.Diagnostics) types.List {
	if len(options) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: offerOptionAttrTypes()})
	}

	models := make([]offerDataSourceOptionModel, 0, len(options))
	for _, option := range options {
		models = append(models, offerDataSourceOptionModel{
			ID:                   types.StringValue(option.ID),
			Name:                 types.StringValue(string(option.Name)),
			BillingOperationPath: types.StringValue(option.BillingOperationPath),
			MinValue:             types.Int64Value(int64(option.MinValue)),
			CurrentValue:         types.Int64Value(int64(option.CurrentValue)),
			MaxValue:             types.Int64Value(int64(option.MaxValue)),
			QuotaWarning:         types.StringValue(string(option.QuotaWarning)),
			Price:                types.StringValue(flattenOfferPrice(option.Price)),
		})
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: offerOptionAttrTypes()}, models)
	diags.Append(d...)

	return list
}

func offerOfferAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"name":                   types.StringType,
		"billing_operation_path": types.StringType,
		"available":              types.BoolType,
		"control_panel_name":     types.StringType,
		"end_of_life":            types.BoolType,
		"quota_warning":          types.StringType,
		"price":                  types.StringType,
		"options":                types.ListType{ElemType: types.ObjectType{AttrTypes: offerOptionAttrTypes()}},
	}
}

func offerOptionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"name":                   types.StringType,
		"billing_operation_path": types.StringType,
		"min_value":              types.Int64Type,
		"current_value":          types.Int64Type,
		"max_value":              types.Int64Type,
		"quota_warning":          types.StringType,
		"price":                  types.StringType,
	}
}

func offerProductAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                  types.StringType,
		"option":                types.BoolType,
		"email_accounts_quota":  types.Int64Type,
		"email_storage_quota":   types.Int64Type,
		"databases_quota":       types.Int64Type,
		"hosting_storage_quota": types.Int64Type,
		"support_included":      types.BoolType,
		"v_cpu":                 types.Int64Type,
		"ram":                   types.Int64Type,
	}
}
