package cockpit

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	identityfw "github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

//go:embed descriptions/cockpit_token_resource.md
var cockpitTokenResourceDescription string

var (
	_ resource.Resource                 = (*TokenResource)(nil)
	_ resource.ResourceWithConfigure    = (*TokenResource)(nil)
	_ resource.ResourceWithImportState  = (*TokenResource)(nil)
	_ resource.ResourceWithIdentity     = (*TokenResource)(nil)
	_ resource.ResourceWithModifyPlan   = (*TokenResource)(nil)
	_ resource.ResourceWithUpgradeState = (*TokenResource)(nil)
)

func NewTokenResource() resource.Resource {
	return &TokenResource{}
}

type TokenResource struct {
	api  *cockpit.RegionalAPI
	meta *meta.Meta
}

type tokenResourceModel struct {
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
	Scopes    types.List     `tfsdk:"scopes"`
	ID        types.String   `tfsdk:"id"`
	Name      types.String   `tfsdk:"name"`
	SecretKey types.String   `tfsdk:"secret_key"`
	CreatedAt types.String   `tfsdk:"created_at"`
	UpdatedAt types.String   `tfsdk:"updated_at"`
	ProjectID types.String   `tfsdk:"project_id"`
	Region    types.String   `tfsdk:"region"`
}

// tokenScopeDescriptions lists the scope flags with their default value, in a stable order.
var tokenScopeDescriptions = map[string]struct {
	description string
	value       bool
}{
	"query_metrics":       {"Query metrics", false},
	"write_metrics":       {"Write metrics", true},
	"setup_metrics_rules": {"Setup metrics rules", false},
	"query_logs":          {"Query logs", false},
	"write_logs":          {"Write logs", true},
	"setup_logs_rules":    {"Setup logs rules", false},
	"setup_alerts":        {"Setup alerts", false},
	"query_traces":        {"Query traces", false},
	"write_traces":        {"Write traces", false},
}

func tokenScopesAttributes() map[string]schema.Attribute {
	attributes := make(map[string]schema.Attribute, len(tokenScopeDescriptions))

	for name, scope := range tokenScopeDescriptions {
		attributes[name] = schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(scope.value),
			MarkdownDescription: scope.description,
		}
	}

	return attributes
}

func tokenScopesObjectType() types.ObjectType {
	attrTypes := make(map[string]attr.Type, len(tokenScopeDescriptions))
	for name := range tokenScopeDescriptions {
		attrTypes[name] = types.BoolType
	}

	return types.ObjectType{AttrTypes: attrTypes}
}

func (r *TokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cockpit_token"
}

func (r *TokenResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             1,
		MarkdownDescription: cockpitTokenResourceDescription,
		Description:         cockpitTokenResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the token",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secret_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The secret key of the token",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time of the creation of the Cockpit Token (Format ISO 8601)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time of the last update of the Cockpit Token (Format ISO 8601)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": account.ProjectIDSchemaFw("The project_id you want to attach the resource to"),
			"region":     regional.SchemaAttributeComputed(),
		},
		Blocks: map[string]schema.Block{
			// Not Computed (Framework blocks cannot be): the defaults are applied on create and
			// the stored scopes are kept in the plan by ModifyPlan when the block is omitted.
			"scopes": schema.ListNestedBlock{
				MarkdownDescription: "Endpoints. If omitted, the token gets the write_metrics and write_logs scopes. Changing it forces a new resource.",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: tokenScopesAttributes(),
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

func (r *TokenResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityfw.DefaultRegional()
}

func (r *TokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = cockpit.NewRegionalAPI(m.ScwClient())
}

// ModifyPlan forces a replacement when the scopes change: every attribute of a token is immutable.
// Terraform requires the planned block count to match the configuration, so an omitted block is
// planned (and stored) empty. Dropping the block of a token that already has the default scopes
// (e.g. a state written by the SDKv2 resource) is not a change, so it is updated in place.
func (r *TokenResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state tokenResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() || plan.Scopes.IsUnknown() || plan.Scopes.Equal(state.Scopes) {
		return
	}

	defaults, diags := flattenTokenScopes(defaultCockpitTokenScopes())
	resp.Diagnostics.Append(diags...)

	if len(plan.Scopes.Elements()) == 0 && state.Scopes.Equal(defaults) {
		return
	}

	resp.RequiresReplace = append(resp.RequiresReplace, path.Root("scopes"))
}

func (r *TokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tokenResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, err := meta.ExtractFrameworkRegion(plan.Region, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve region", err.Error())

		return
	}

	projectID, err := meta.ExtractFrameworkProjectID(plan.ProjectID, r.meta.ScwClient())
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve project ID", err.Error())

		return
	}

	scopes, diags := expandTokenEphemeralScopes(ctx, plan.Scopes)
	resp.Diagnostics.Append(diags...)

	timeout, diags := plan.Timeouts.Create(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	createCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := transport.RetryOnStatusValue(createCtx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.Token, error) {
		return r.api.CreateToken(&cockpit.RegionalAPICreateTokenRequest{
			Name:        plan.Name.ValueString(),
			TokenScopes: scopes,
			ProjectID:   projectID,
			Region:      region,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Cockpit token", err.Error())

		return
	}

	// The secret key is only returned on creation.
	plan.SecretKey = types.StringPointerValue(res.SecretKey)

	state, diags := flattenToken(res, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(res.Region, res.ID))...)
}

// flattenToken builds the state from the API token; prior provides the secret key and timeouts.
func flattenToken(res *cockpit.Token, prior tokenResourceModel) (tokenResourceModel, diag.Diagnostics) {
	// An omitted scopes block stays empty; import / pre-identity upgrades (null scopes) refresh from the API.
	scopes := prior.Scopes

	var diags diag.Diagnostics

	if prior.Name.IsNull() || prior.Scopes.IsNull() || len(scopes.Elements()) > 0 {
		scopes, diags = flattenTokenScopes(res.Scopes)
	}

	model := tokenResourceModel{
		Timeouts:  prior.Timeouts,
		Scopes:    scopes,
		ID:        types.StringValue(regional.NewIDString(res.Region, res.ID)),
		Name:      types.StringValue(res.Name),
		SecretKey: prior.SecretKey,
		CreatedAt: flattenCockpitTime(res.CreatedAt),
		UpdatedAt: flattenCockpitTime(res.UpdatedAt),
		ProjectID: types.StringValue(res.ProjectID),
		Region:    types.StringValue(res.Region.String()),
	}

	return model, diags
}

func flattenTokenScopes(scopes []cockpit.TokenScope) (types.List, diag.Diagnostics) {
	flags := flattenCockpitTokenScopes(scopes)
	values := make(map[string]attr.Value, len(flags))

	for name, enabled := range flags {
		values[name] = types.BoolValue(enabled)
	}

	object, diags := types.ObjectValue(tokenScopesObjectType().AttrTypes, values)
	if diags.HasError() {
		return types.ListNull(tokenScopesObjectType()), diags
	}

	return types.ListValue(tokenScopesObjectType(), []attr.Value{object})
}

func (r *TokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	identity, identityAvailable := getRegionalIdentity(ctx, req.Identity)

	var state tokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := resolveRegionalID(identity, identityAvailable, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Cockpit token ID", err.Error())

		return
	}

	timeout, diags := state.Timeouts.Read(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	readCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := transport.RetryOnStatusValue(readCtx, transport.IAMPropagation, func(ctx context.Context) (*cockpit.Token, error) {
		return r.api.GetToken(&cockpit.RegionalAPIGetTokenRequest{
			Region:  region,
			TokenID: id,
		}, scw.WithContext(ctx))
	})
	if err != nil {
		if httperrors.Is404(err) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read Cockpit token", err.Error())

		return
	}

	// The API never returns the secret key again: flattenToken keeps the one from the state.
	newState, diags := flattenToken(res, state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, identityfw.SetRegionalIdentity(res.Region, res.ID))...)
}

// Update only runs when the omitted scopes block is dropped from the state or on a timeouts change.
func (r *TokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tokenResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	identity, identityAvailable := getRegionalIdentity(ctx, req.Identity)

	var state tokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	region, id, err := resolveRegionalID(identity, identityAvailable, state.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse Cockpit token ID", err.Error())

		return
	}

	timeout, diags := state.Timeouts.Delete(ctx, DefaultCockpitTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	deleteCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = transport.RetryOnStatus(deleteCtx, transport.IAMPropagation, func(ctx context.Context) error {
		return r.api.DeleteToken(&cockpit.RegionalAPIDeleteTokenRequest{
			Region:  region,
			TokenID: id,
		}, scw.WithContext(ctx))
	})
	if err != nil && !httperrors.Is404(err) {
		resp.Diagnostics.AddError("Failed to delete Cockpit token", err.Error())
	}
}

func (r *TokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("id"), path.Root("id"), req, resp)
}

// UpgradeState upgrades the states written before the region was part of the ID (schema version 0).
func (r *TokenResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: upgradeTokenStateV0},
	}
}

func upgradeTokenStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var rawState map[string]any

	err := json.Unmarshal(req.RawState.JSON, &rawState)
	if err != nil {
		resp.Diagnostics.AddError("Failed to decode Cockpit token state", err.Error())

		return
	}

	upgraded, err := cockpitTokenV1UpgradeFunc(ctx, rawState, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to upgrade Cockpit token state", err.Error())

		return
	}

	// scopes stay null so the next Read (prior.Scopes.IsNull) refreshes them from the API.
	for _, key := range []string{"id", "region", "name", "project_id", "secret_key", "created_at", "updated_at"} {
		value, ok := upgraded[key].(string)
		if !ok {
			continue
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(key), value)...)
	}
}
