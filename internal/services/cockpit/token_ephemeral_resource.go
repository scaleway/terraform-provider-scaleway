package cockpit

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

var (
	_ ephemeral.EphemeralResource              = (*TokenEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*TokenEphemeralResource)(nil)
	_ ephemeral.EphemeralResourceWithClose     = (*TokenEphemeralResource)(nil)
)

type TokenEphemeralResource struct {
	cockpitAPI *cockpit.RegionalAPI
	meta       *meta.Meta
}

func NewTokenEphemeralResource() ephemeral.EphemeralResource {
	return &TokenEphemeralResource{}
}

func (r *TokenEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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
	r.cockpitAPI = cockpit.NewRegionalAPI(client)
	r.meta = m
}

func (r *TokenEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_token"
}

type TokenEphemeralResourceModel struct {
	Name      types.String `tfsdk:"name"`
	ProjectID types.String `tfsdk:"project_id"`
	Region    types.String `tfsdk:"region"`
	Scopes    types.List   `tfsdk:"scopes"`
	// Output
	SecretKey types.String `tfsdk:"secret_key"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

//go:embed descriptions/token_ephemeral_resource.md
var tokenEphemeralResourceDescription string

func (r *TokenEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         tokenEphemeralResourceDescription,
		MarkdownDescription: tokenEphemeralResourceDescription,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the token",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the Scaleway project the token belongs to",
			},
			"region": regional.SchemaAttribute("Region of the token. If not set, the region is derived from the provider configuration."),
			"scopes": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Token permission scopes. If not set, defaults to write_metrics and write_logs.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"query_metrics": schema.BoolAttribute{
							Optional:    true,
							Description: "Query metrics",
						},
						"write_metrics": schema.BoolAttribute{
							Optional:    true,
							Description: "Write metrics",
						},
						"setup_metrics_rules": schema.BoolAttribute{
							Optional:    true,
							Description: "Setup metrics rules",
						},
						"query_logs": schema.BoolAttribute{
							Optional:    true,
							Description: "Query logs",
						},
						"write_logs": schema.BoolAttribute{
							Optional:    true,
							Description: "Write logs",
						},
						"setup_logs_rules": schema.BoolAttribute{
							Optional:    true,
							Description: "Setup logs rules",
						},
						"setup_alerts": schema.BoolAttribute{
							Optional:    true,
							Description: "Setup alerts",
						},
						"query_traces": schema.BoolAttribute{
							Optional:    true,
							Description: "Query traces",
						},
						"write_traces": schema.BoolAttribute{
							Optional:    true,
							Description: "Write traces",
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"secret_key": schema.StringAttribute{
				Computed:    true,
				Description: "The secret key of the token",
				Sensitive:   true,
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the creation of the Cockpit Token (Format ISO 8601)",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the last update of the Cockpit Token (Format ISO 8601)",
			},
		},
	}
}

func (r *TokenEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data TokenEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if r.cockpitAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured cockpitAPI",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	// Extract project_id
	var projectID string
	if !data.ProjectID.IsNull() && !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	} else {
		var exists bool

		projectID, exists = r.meta.ScwClient().GetDefaultProjectID()
		if !exists {
			resp.Diagnostics.AddError(
				"Missing project_id",
				"Please provide a project_id explicitly or configure a default project in the provider.",
			)

			return
		}
	}

	// Extract region
	var region scw.Region
	if !data.Region.IsNull() && !data.Region.IsUnknown() && data.Region.ValueString() != "" {
		region = scw.Region(data.Region.ValueString())
	} else {
		var exists bool

		region, exists = r.meta.ScwClient().GetDefaultRegion()
		if !exists {
			resp.Diagnostics.AddError(
				"Missing region",
				"Please provide a region explicitly or configure a default region in the provider.",
			)

			return
		}
	}

	// Expand scopes
	var tokenScopes []cockpit.TokenScope

	if !data.Scopes.IsNull() && !data.Scopes.IsUnknown() {
		var scopesMap []map[string]bool

		diags := data.Scopes.ElementsAs(ctx, &scopesMap, false)
		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		if len(scopesMap) > 0 {
			for key, tokenScope := range scopeMapping {
				if value, ok := scopesMap[0][key]; ok && value {
					tokenScopes = append(tokenScopes, tokenScope)
				}
			}
		}
	}

	// Default scopes if none specified
	if len(tokenScopes) == 0 {
		tokenScopes = []cockpit.TokenScope{
			cockpit.TokenScopeWriteOnlyMetrics,
			cockpit.TokenScopeWriteOnlyLogs,
		}
	}

	name := data.Name.ValueString()

	// Create the token
	res, err := r.cockpitAPI.CreateToken(&cockpit.RegionalAPICreateTokenRequest{
		Region:      region,
		ProjectID:   projectID,
		Name:        name,
		TokenScopes: tokenScopes,
	}, scw.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Cockpit Token",
			fmt.Sprintf("Failed to create Cockpit token %q in region %s: %s", name, region, err),
		)

		return
	}

	data.CreatedAt = types.StringValue(res.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	data.UpdatedAt = types.StringValue(res.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
	data.SecretKey = types.StringValue(*res.SecretKey)

	// Store token ID in Private so Close() can delete it
	// Private data lives only in memory for the duration of a single phase
	// (plan or apply) and is never persisted to state or plan files.
	if err := resp.Private.SetKey(ctx, "token_id", []byte(res.ID)); err != nil {
		resp.Diagnostics.AddError(
			"Error storing private data",
			fmt.Sprintf("Failed to store token ID for later cleanup: %s", err),
		)

		return
	}

	if err := resp.Private.SetKey(ctx, "region", []byte(string(res.Region))); err != nil {
		resp.Diagnostics.AddError(
			"Error storing private data",
			fmt.Sprintf("Failed to store region for later cleanup: %s", err),
		)

		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

func (r *TokenEphemeralResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	tokenIDBytes, diags := req.Private.GetKey(ctx, "token_id")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	regionBytes, diags := req.Private.GetKey(ctx, "region")

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tokenID := string(tokenIDBytes)
	region := scw.Region(string(regionBytes))

	// Delete the token
	if err := r.cockpitAPI.DeleteToken(&cockpit.RegionalAPIDeleteTokenRequest{
		Region:  region,
		TokenID: tokenID,
	}, scw.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError(
			"Error closing Cockpit Token",
			fmt.Sprintf("Failed to delete Cockpit token %s in region %s: %s", tokenID, region, err),
		)
	}
}
