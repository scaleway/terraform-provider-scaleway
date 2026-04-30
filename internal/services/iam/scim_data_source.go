package iam

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ datasource.DataSource              = (*ScimDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ScimDataSource)(nil)
)

func NewScimDataSource() datasource.DataSource {
	return &ScimDataSource{}
}

type ScimDataSource struct {
	iamAPI *iam.API
	meta   *meta.Meta
}

func (d *ScimDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_scim"
}

//go:embed descriptions/scim_data_source.md
var scimDataSourceDescription string

func (d *ScimDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: scimDataSourceDescription,
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID. If not provided, the default organization configured in the provider is used.",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the SCIM configuration",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The date and time of SCIM configuration creation",
				Computed:            true,
			},
		},
	}
}

func (d *ScimDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.iamAPI = iam.NewAPI(d.meta.ScwClient())
}

func (d *ScimDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state scimResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	orgID := state.OrganizationID.ValueString()
	if orgID == "" {
		defaultOrgID, exists := d.meta.ScwClient().GetDefaultOrganizationID()
		if exists {
			orgID = defaultOrgID
		} else {
			resp.Diagnostics.AddAttributeError(
				path.Root("organization_id"),
				"Organization ID is required",
				"Either set organization_id or configure a default organization",
			)

			return
		}
	}

	scim, err := d.iamAPI.GetOrganizationScim(&iam.GetOrganizationScimRequest{
		OrganizationID: orgID,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read SCIM",
			err.Error(),
		)

		return
	}

	state = convertScimToState(scim, orgID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
