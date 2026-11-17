package annotations

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	frameworkSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

type frameworkDefaultBindingsWrapper struct {
	inner       resource.Resource
	meta        *meta.Meta
	srnComputed bool
	srnChecked  bool
}

// frameworkDefaultBindingsIdentityWrapper extends frameworkDefaultBindingsWrapper for
// inner resources that support identity. The Framework unconditionally queries
// IdentitySchema on every resource implementing resource.ResourceWithIdentity, so
// only the wrappers of identity resources implement the identity interfaces.
type frameworkDefaultBindingsIdentityWrapper struct {
	*frameworkDefaultBindingsWrapper
}

// WithDefaultBindingsFramework wraps a Framework resource factory so that, after
// a successful Create or Update, the default annotation bindings of the
// resource's SRN are reconciled and, before Delete, the default bindings are
// removed. Annotation failures produce warnings and never fail the resource
// operation, except when a default annotation key or value cannot be resolved,
// which is a configuration error. Resources without a Computed "srn" attribute
// are no-ops.
func WithDefaultBindingsFramework(factory func() resource.Resource) func() resource.Resource {
	return func() resource.Resource {
		inner := factory()

		base := &frameworkDefaultBindingsWrapper{inner: inner}

		if _, ok := inner.(resource.ResourceWithIdentity); ok {
			return &frameworkDefaultBindingsIdentityWrapper{frameworkDefaultBindingsWrapper: base}
		}

		return base
	}
}

func (w *frameworkDefaultBindingsWrapper) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	w.inner.Metadata(ctx, req, resp)
}

func (w *frameworkDefaultBindingsWrapper) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	w.inner.Schema(ctx, req, resp)
}

func (w *frameworkDefaultBindingsWrapper) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	w.inner.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(w.reconcileDiags(ctx, &resp.State)...)
}

func (w *frameworkDefaultBindingsWrapper) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	w.inner.Read(ctx, req, resp)
}

func (w *frameworkDefaultBindingsWrapper) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	w.inner.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(w.reconcileDiags(ctx, &resp.State)...)
}

func (w *frameworkDefaultBindingsWrapper) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if srn := w.resourceSrn(ctx, &req.State); srn != "" {
		if err := DeleteDefaultBindingsForSrn(ctx, w.meta, srn); err != nil {
			resp.Diagnostics.AddWarning("Failed to remove default annotation bindings", err.Error())
		}
	}

	w.inner.Delete(ctx, req, resp)
}

func (w *frameworkDefaultBindingsWrapper) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r, ok := w.inner.(resource.ResourceWithConfigure); ok {
		r.Configure(ctx, req, resp)
	}

	if m, ok := req.ProviderData.(*meta.Meta); ok {
		w.meta = m
	}
}

func (w *frameworkDefaultBindingsWrapper) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	w.inner.(resource.ResourceWithImportState).ImportState(ctx, req, resp)
}

func (w *frameworkDefaultBindingsWrapper) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if r, ok := w.inner.(resource.ResourceWithModifyPlan); ok {
		r.ModifyPlan(ctx, req, resp)
	}
}

func (w *frameworkDefaultBindingsWrapper) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if r, ok := w.inner.(resource.ResourceWithUpgradeState); ok {
		return r.UpgradeState(ctx)
	}

	return nil
}

func (w *frameworkDefaultBindingsWrapper) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if r, ok := w.inner.(resource.ResourceWithConfigValidators); ok {
		return r.ConfigValidators(ctx)
	}

	return nil
}

func (w *frameworkDefaultBindingsWrapper) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r, ok := w.inner.(resource.ResourceWithValidateConfig); ok {
		r.ValidateConfig(ctx, req, resp)
	}
}

func (w *frameworkDefaultBindingsWrapper) MoveState(ctx context.Context) []resource.StateMover {
	if r, ok := w.inner.(resource.ResourceWithMoveState); ok {
		return r.MoveState(ctx)
	}

	return nil
}

func (w *frameworkDefaultBindingsIdentityWrapper) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	w.inner.(resource.ResourceWithIdentity).IdentitySchema(ctx, req, resp)
}

func (w *frameworkDefaultBindingsIdentityWrapper) UpgradeIdentity(ctx context.Context) map[int64]resource.IdentityUpgrader {
	if r, ok := w.inner.(resource.ResourceWithUpgradeIdentity); ok {
		return r.UpgradeIdentity(ctx)
	}

	return nil
}

func (w *frameworkDefaultBindingsWrapper) resourceSrn(ctx context.Context, state *tfsdk.State) string {
	if w.meta == nil || !w.hasComputedSrn(ctx) {
		return ""
	}

	return stateStringAttribute(ctx, state, "srn")
}

func (w *frameworkDefaultBindingsWrapper) reconcileDiags(ctx context.Context, state *tfsdk.State) diag.Diagnostics {
	srn := w.resourceSrn(ctx, state)
	if srn == "" {
		return nil
	}

	err := ReconcileDefaultBindingsForSrn(ctx, w.meta, srn)
	if err == nil {
		return nil
	}

	if IsResolutionError(err) {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Failed to apply default annotations", err.Error())}
	}

	return diag.Diagnostics{diag.NewWarningDiagnostic("Failed to apply default annotations", err.Error())}
}

func (w *frameworkDefaultBindingsWrapper) hasComputedSrn(ctx context.Context) bool {
	if w.srnChecked {
		return w.srnComputed
	}

	w.srnChecked = true

	var schemaResp resource.SchemaResponse

	w.inner.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	if schemaResp.Diagnostics.HasError() {
		return false
	}

	attr, ok := schemaResp.Schema.Attributes["srn"]
	if !ok {
		return false
	}

	strAttr, ok := attr.(frameworkSchema.StringAttribute)

	w.srnComputed = ok && strAttr.Computed

	return w.srnComputed
}

func stateStringAttribute(ctx context.Context, state *tfsdk.State, attr string) string {
	if state == nil || state.Raw.IsNull() {
		return ""
	}

	var s string

	diags := state.GetAttribute(ctx, path.Root(attr), &s)
	if diags.HasError() {
		return ""
	}

	return s
}

var (
	_ resource.Resource                     = (*frameworkDefaultBindingsWrapper)(nil)
	_ resource.ResourceWithConfigure        = (*frameworkDefaultBindingsWrapper)(nil)
	_ resource.ResourceWithImportState      = (*frameworkDefaultBindingsWrapper)(nil)
	_ resource.ResourceWithModifyPlan       = (*frameworkDefaultBindingsWrapper)(nil)
	_ resource.ResourceWithUpgradeState     = (*frameworkDefaultBindingsWrapper)(nil)
	_ resource.ResourceWithConfigValidators = (*frameworkDefaultBindingsWrapper)(nil)
	_ resource.ResourceWithValidateConfig   = (*frameworkDefaultBindingsWrapper)(nil)
	_ resource.ResourceWithMoveState        = (*frameworkDefaultBindingsWrapper)(nil)

	_ resource.ResourceWithIdentity        = (*frameworkDefaultBindingsIdentityWrapper)(nil)
	_ resource.ResourceWithUpgradeIdentity = (*frameworkDefaultBindingsIdentityWrapper)(nil)
)
