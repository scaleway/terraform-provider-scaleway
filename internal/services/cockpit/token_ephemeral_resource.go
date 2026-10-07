package cockpit

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
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

func (r *TokenEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

	r.cockpitAPI = cockpit.NewRegionalAPI(m.ScwClient())
	r.meta = m
}

func (r *TokenEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
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

type tokenEphemeralScopesModel struct {
	QueryMetrics      types.Bool `tfsdk:"query_metrics"`
	WriteMetrics      types.Bool `tfsdk:"write_metrics"`
	SetupMetricsRules types.Bool `tfsdk:"setup_metrics_rules"`
	QueryLogs         types.Bool `tfsdk:"query_logs"`
	WriteLogs         types.Bool `tfsdk:"write_logs"`
	SetupLogsRules    types.Bool `tfsdk:"setup_logs_rules"`
	SetupAlerts       types.Bool `tfsdk:"setup_alerts"`
	QueryTraces       types.Bool `tfsdk:"query_traces"`
	WriteTraces       types.Bool `tfsdk:"write_traces"`
}

//go:embed descriptions/token_ephemeral_resource.md
var tokenEphemeralResourceDescription string

func (r *TokenEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
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
				Computed:    true,
				Description: "ID of the Scaleway project the token belongs to",
				Validators: []validator.String{
					verify.IsStringUUID(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Region of the token. If not set, the region is derived from the provider configuration.",
				Validators: []validator.String{
					verify.IsStringOneOfWithWarning(regional.AllRegions()),
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
		Blocks: map[string]schema.Block{
			"scopes": schema.ListNestedBlock{
				Description: "Token permission scopes. If not set, defaults to write_metrics and write_logs.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"query_metrics": schema.BoolAttribute{
							Optional:    true,
							Description: "Query metrics. Defaults to false (same as scaleway_cockpit_token).",
						},
						"write_metrics": schema.BoolAttribute{
							Optional:    true,
							Description: "Write metrics. Defaults to true (same as scaleway_cockpit_token).",
						},
						"setup_metrics_rules": schema.BoolAttribute{
							Optional:    true,
							Description: "Setup metrics rules. Defaults to false (same as scaleway_cockpit_token).",
						},
						"query_logs": schema.BoolAttribute{
							Optional:    true,
							Description: "Query logs. Defaults to false (same as scaleway_cockpit_token).",
						},
						"write_logs": schema.BoolAttribute{
							Optional:    true,
							Description: "Write logs. Defaults to true (same as scaleway_cockpit_token).",
						},
						"setup_logs_rules": schema.BoolAttribute{
							Optional:    true,
							Description: "Setup logs rules. Defaults to false (same as scaleway_cockpit_token).",
						},
						"setup_alerts": schema.BoolAttribute{
							Optional:    true,
							Description: "Setup alerts. Defaults to false (same as scaleway_cockpit_token).",
						},
						"query_traces": schema.BoolAttribute{
							Optional:    true,
							Description: "Query traces. Defaults to false (same as scaleway_cockpit_token).",
						},
						"write_traces": schema.BoolAttribute{
							Optional:    true,
							Description: "Write traces. Defaults to false (same as scaleway_cockpit_token).",
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func expandTokenEphemeralScopes(ctx context.Context, scopes types.List) ([]cockpit.TokenScope, diag.Diagnostics) {
	var diags diag.Diagnostics

	if scopes.IsNull() || scopes.IsUnknown() {
		return nil, diags
	}

	var scopeModels []tokenEphemeralScopesModel

	diags.Append(scopes.ElementsAs(ctx, &scopeModels, false)...)

	if diags.HasError() {
		return nil, diags
	}

	if len(scopeModels) == 0 {
		return nil, diags
	}

	s := scopeModels[0]
	// Defaults match the managed scaleway_cockpit_token resource (write_metrics / write_logs = true).
	flags := map[string]any{
		"query_metrics":       ephemeralScopeBool(s.QueryMetrics, false),
		"write_metrics":       ephemeralScopeBool(s.WriteMetrics, true),
		"setup_metrics_rules": ephemeralScopeBool(s.SetupMetricsRules, false),
		"query_logs":          ephemeralScopeBool(s.QueryLogs, false),
		"write_logs":          ephemeralScopeBool(s.WriteLogs, true),
		"setup_logs_rules":    ephemeralScopeBool(s.SetupLogsRules, false),
		"setup_alerts":        ephemeralScopeBool(s.SetupAlerts, false),
		"query_traces":        ephemeralScopeBool(s.QueryTraces, false),
		"write_traces":        ephemeralScopeBool(s.WriteTraces, false),
	}

	return expandCockpitTokenScopesFromFlags(flags), diags
}

func ephemeralScopeBool(value types.Bool, defaultValue bool) bool {
	if value.IsNull() || value.IsUnknown() {
		return defaultValue
	}

	return value.ValueBool()
}

func (r *TokenEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data TokenEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if r.cockpitAPI == nil || r.meta == nil {
		resp.Diagnostics.AddError(
			"Unconfigured cockpitAPI",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(data.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError(
			"Missing project_id",
			"Please provide a project_id explicitly or configure a default project in the provider.",
		)

		return
	}

	region, err := meta.ExtractFrameworkRegion(data.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError(
			"Missing region",
			"Please provide a region explicitly or configure a default region in the provider.",
		)

		return
	}

	tokenScopes, diags := expandTokenEphemeralScopes(ctx, data.Scopes)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(tokenScopes) == 0 {
		tokenScopes = []cockpit.TokenScope{
			cockpit.TokenScopeWriteOnlyMetrics,
			cockpit.TokenScopeWriteOnlyLogs,
		}
	}

	name := data.Name.ValueString()

	res, err := retryOn403Value(ctx, func() (*cockpit.Token, error) {
		return r.cockpitAPI.CreateToken(&cockpit.RegionalAPICreateTokenRequest{
			Region:      region,
			ProjectID:   projectID,
			Name:        name,
			TokenScopes: tokenScopes,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Cockpit Token",
			fmt.Sprintf("Failed to create Cockpit token %q in region %s: %s", name, region, err),
		)

		return
	}

	cleanup := func() {
		_ = r.deleteToken(ctx, res.Region, res.ID)
	}

	tokenIDPrivate, err := json.Marshal(res.ID)
	if err != nil {
		cleanup()
		resp.Diagnostics.AddError(
			"Error storing private data",
			fmt.Sprintf("Failed to encode token ID for later cleanup: %s", err),
		)

		return
	}

	if err := resp.Private.SetKey(ctx, "token_id", tokenIDPrivate); err != nil {
		cleanup()
		resp.Diagnostics.AddError(
			"Error storing private data",
			fmt.Sprintf("Failed to store token ID for later cleanup: %s", err),
		)

		return
	}

	regionPrivate, err := json.Marshal(res.Region.String())
	if err != nil {
		cleanup()
		resp.Diagnostics.AddError(
			"Error storing private data",
			fmt.Sprintf("Failed to encode region for later cleanup: %s", err),
		)

		return
	}

	if err := resp.Private.SetKey(ctx, "region", regionPrivate); err != nil {
		cleanup()
		resp.Diagnostics.AddError(
			"Error storing private data",
			fmt.Sprintf("Failed to store region for later cleanup: %s", err),
		)

		return
	}

	data.ProjectID = types.StringValue(res.ProjectID)
	data.Region = types.StringValue(res.Region.String())
	data.CreatedAt = types.StringValue(scwtypes.FlattenTime(res.CreatedAt).(string))
	data.UpdatedAt = types.StringValue(scwtypes.FlattenTime(res.UpdatedAt).(string))

	if res.SecretKey != nil {
		data.SecretKey = types.StringValue(*res.SecretKey)
	} else {
		data.SecretKey = types.StringNull()
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		cleanup()
	}
}

func (r *TokenEphemeralResource) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if r.cockpitAPI == nil {
		resp.Diagnostics.AddError(
			"Unconfigured cockpitAPI",
			"The ephemeral resource was not properly configured. The Scaleway client is missing. "+
				"This is usually a bug in the provider. Please report it to the maintainers.",
		)

		return
	}

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

	var tokenID string
	if err := json.Unmarshal(tokenIDBytes, &tokenID); err != nil {
		resp.Diagnostics.AddError(
			"Error reading private data",
			fmt.Sprintf("Failed to decode token ID for cleanup: %s", err),
		)

		return
	}

	var regionStr string
	if err := json.Unmarshal(regionBytes, &regionStr); err != nil {
		resp.Diagnostics.AddError(
			"Error reading private data",
			fmt.Sprintf("Failed to decode region for cleanup: %s", err),
		)

		return
	}

	region := scw.Region(regionStr)

	// Do not retryOn403 here: Cockpit returns 403 when the token is already gone,
	// which must not be treated as IAM propagation.
	err := r.deleteToken(ctx, region, tokenID)
	if err != nil && !httperrors.Is404(err) && !httperrors.Is403(err) {
		resp.Diagnostics.AddError(
			"Error closing Cockpit Token",
			fmt.Sprintf("Failed to delete Cockpit token %s in region %s: %s", tokenID, region, err),
		)
	}
}

func (r *TokenEphemeralResource) deleteToken(ctx context.Context, region scw.Region, tokenID string) error {
	return r.cockpitAPI.DeleteToken(&cockpit.RegionalAPIDeleteTokenRequest{
		Region:  region,
		TokenID: tokenID,
	}, scw.WithContext(ctx))
}
