package cockpit

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cockpit "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

//go:embed descriptions/cockpit_products_data_source.md
var cockpitProductsDataSourceDescription string

var (
	_ datasource.DataSource              = (*ProductsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ProductsDataSource)(nil)
)

func NewProductsDataSource() datasource.DataSource {
	return &ProductsDataSource{}
}

type ProductsDataSource struct {
	dataSourceBase
}

type productModel struct {
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	FamilyName  types.String `tfsdk:"family_name"`
}

type productsDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Region   types.String   `tfsdk:"region"`
	Products []productModel `tfsdk:"products"`
	Names    []string       `tfsdk:"names"`
}

func (d *ProductsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_products"
}

func (d *ProductsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: cockpitProductsDataSourceDescription,
		Description:         cockpitProductsDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"id":     schema.StringAttribute{Computed: true, MarkdownDescription: "The region of the products."},
			"region": dataSourceRegionAttribute(),
			"products": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of Cockpit products available for exported_products in scaleway_cockpit_exporter.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":         schema.StringAttribute{Computed: true, MarkdownDescription: "Product name to use in exported_products (e.g. cockpit, LB, object-storage)."},
						"display_name": schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable display name of the product."},
						"family_name":  schema.StringAttribute{Computed: true, MarkdownDescription: "Product family name."},
					},
				},
			},
			"names": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of product names for use in scaleway_cockpit_exporter.exported_products.",
			},
		},
	}
}

func (d *ProductsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data productsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	api, region, err := d.regionalAPI(data.Region)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	res, err := transport.RetryOnStatusValue(ctx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.ListProductsResponse, error) {
		return api.ListProducts(&cockpit.RegionalAPIListProductsRequest{
			Region:  region,
			OrderBy: cockpit.ListProductsRequestOrderByDisplayNameAsc,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list Cockpit products", err.Error())

		return
	}

	data.Products = make([]productModel, 0, len(res.ProductsList))
	data.Names = make([]string, 0, len(res.ProductsList))

	for _, p := range res.ProductsList {
		data.Products = append(data.Products, productModel{
			Name:        types.StringValue(p.Name),
			DisplayName: types.StringValue(p.DisplayName),
			FamilyName:  types.StringValue(p.FamilyName),
		})
		data.Names = append(data.Names, p.Name)
	}

	data.ID = types.StringValue(region.String())
	data.Region = types.StringValue(region.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
