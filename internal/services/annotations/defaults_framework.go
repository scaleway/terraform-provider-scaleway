package annotations

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// frameworkBindingWrapper wraps a Framework resource.Resource so that default
// annotation bindings are managed based on the resource's "srn" attribute:
// after Create and after Update, the bindings are fully reconciled (missing
// defaults are created, stale values under default keys are deleted); on
// Read, no bindings are created or deleted (Read must not perform writes
// since it runs during plan and refresh); before Delete, default bindings
// of the resource are deleted. On Create, Update and Delete, the
// best-effort cleanup of orphaned default annotation keys and values is run
// (deferred from Configure so that no deletes happen during plan).
//
// The wrapper transparently delegates all Framework optional interfaces
// (Configure, ImportState, ModifyPlan, etc.) to the inner resource when it
// implements them. Identity support is handled by
// frameworkBindingIdentityWrapper, which only wraps resources that already
// implement resource.ResourceWithIdentity.
//
// Resources without a Computed "srn" attribute or with an empty SRN are
// no-ops. The annotations binding resource (whose "srn" is Computed) is
// excluded from wrapping at registration time in provider/framework.go so
// that it does not receive default bindings on its own SRN.
type frameworkBindingWrapper struct {
	inner       resource.Resource
	meta        *meta.Meta
	srnComputed bool
	srnChecked  bool
}

// frameworkBindingIdentityWrapper extends frameworkBindingWrapper for inner
// resources that support identity. It must be the only wrapper type that
// implements resource.ResourceWithIdentity: the Framework unconditionally
// queries IdentitySchema on every resource that implements the interface, and
// Terraform core (>= 1.13) treats the resulting error diagnostic as fatal for
// every provider RPC.
type frameworkBindingIdentityWrapper struct {
	*frameworkBindingWrapper
}

// WithDefaultBindingsFramework wraps a Framework resource factory so that
// default annotation bindings are reconciled on Create, Read and Update, and
// all bindings are deleted on Delete, for any resource that exposes a Computed
// "srn" attribute. It is the Framework counterpart to WithDefaultBindings
// (which targets SDKv2 resources).
func WithDefaultBindingsFramework(factory func() resource.Resource) func() resource.Resource {
	return func() resource.Resource {
		inner := factory()

		base := &frameworkBindingWrapper{inner: inner}

		if _, ok := inner.(resource.ResourceWithIdentity); ok {
			return &frameworkBindingIdentityWrapper{frameworkBindingWrapper: base}
		}

		return base
	}
}

// --- Core resource.Resource interface ---

func (w *frameworkBindingWrapper) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	w.inner.Metadata(ctx, req, resp)
}

func (w *frameworkBindingWrapper) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	w.inner.Schema(ctx, req, resp)
}

func (w *frameworkBindingWrapper) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	w.runCleanup(ctx, &resp.Diagnostics)

	w.inner.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := w.reconcileFromState(ctx, &resp.State); err != nil {
		resp.Diagnostics.AddError("Failed to reconcile default annotation bindings", err.Error())
	}
}

func (w *frameworkBindingWrapper) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	w.inner.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	// The resource no longer exists: state has been cleared.
	if resp.State.Raw.IsNull() {
		return
	}

	// Read is called during plan (refresh) and must not perform writes.
	// Binding reconciliation is done on Create and Update.
}

func (w *frameworkBindingWrapper) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	w.runCleanup(ctx, &resp.Diagnostics)

	w.inner.Update(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := w.reconcileFromState(ctx, &resp.State); err != nil {
		resp.Diagnostics.AddError("Failed to reconcile default annotation bindings", err.Error())
	}
}

func (w *frameworkBindingWrapper) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// If deleting the bindings fails, abort the destroy before the inner
	// Delete runs, so the resource stays in state and the destroy can be
	// retried instead of annotations errors being silently ignored.
	if w.hasComputedSrn(ctx) {
		srn := stateStringAttribute(ctx, &req.State, "srn")
		if srn != "" {
			if err := DeleteDefaultBindingsFromMeta(ctx, w.meta, srn); err != nil {
				resp.Diagnostics.AddError("Failed to delete default annotation bindings", err.Error())

				return
			}
		}
	}

	w.runCleanup(ctx, &resp.Diagnostics)

	w.inner.Delete(ctx, req, resp)
}

// --- Optional interfaces (delegate to inner when supported) ---

func (w *frameworkBindingWrapper) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if r, ok := w.inner.(resource.ResourceWithConfigure); ok {
		r.Configure(ctx, req, resp)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if m, ok := req.ProviderData.(*meta.Meta); ok {
		w.meta = m
	}
}

func (w *frameworkBindingWrapper) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r, ok := w.inner.(resource.ResourceWithImportState); ok {
		r.ImportState(ctx, req, resp)

		return
	}

	resp.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import.",
	)
}

func (w *frameworkBindingWrapper) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if r, ok := w.inner.(resource.ResourceWithModifyPlan); ok {
		r.ModifyPlan(ctx, req, resp)
	}
}

func (w *frameworkBindingWrapper) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if r, ok := w.inner.(resource.ResourceWithUpgradeState); ok {
		return r.UpgradeState(ctx)
	}

	return nil
}

func (w *frameworkBindingWrapper) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if r, ok := w.inner.(resource.ResourceWithConfigValidators); ok {
		return r.ConfigValidators(ctx)
	}

	return nil
}

func (w *frameworkBindingWrapper) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r, ok := w.inner.(resource.ResourceWithValidateConfig); ok {
		r.ValidateConfig(ctx, req, resp)
	}
}

func (w *frameworkBindingWrapper) MoveState(ctx context.Context) []resource.StateMover {
	if r, ok := w.inner.(resource.ResourceWithMoveState); ok {
		return r.MoveState(ctx)
	}

	return nil
}

// IdentitySchema delegates to the inner resource. This method is safe to
// always delegate because the wrapper is only used when the inner resource
// implements resource.ResourceWithIdentity.
func (w *frameworkBindingIdentityWrapper) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if r, ok := w.inner.(resource.ResourceWithIdentity); ok {
		r.IdentitySchema(ctx, req, resp)

		return
	}

	resp.Diagnostics.AddError(
		"Resource Identity Not Implemented",
		"This resource does not support identity.",
	)
}

func (w *frameworkBindingIdentityWrapper) UpgradeIdentity(ctx context.Context) map[int64]resource.IdentityUpgrader {
	if r, ok := w.inner.(resource.ResourceWithUpgradeIdentity); ok {
		return r.UpgradeIdentity(ctx)
	}

	return nil
}

// reconcileFromState reconciles the default annotation bindings of the
// resource described by state (creates missing bindings and deletes stale
// ones). It is a no-op for resources without a Computed "srn" attribute or
// with an empty SRN. Used on Create and Update.
func (w *frameworkBindingWrapper) reconcileFromState(ctx context.Context, state *tfsdk.State) error {
	if !w.hasComputedSrn(ctx) {
		return nil
	}

	srn := stateStringAttribute(ctx, state, "srn")
	if srn == "" {
		return nil
	}

	return ReconcileDefaultBindingsFromMeta(ctx, w.meta, srn)
}

// runCleanup runs the best-effort cleanup of orphaned default annotation keys
// and values. It is deferred from Configure to Create/Update/Delete so that
// no deletes happen during plan. The cleanup runs at most once per Meta.
func (w *frameworkBindingWrapper) runCleanup(ctx context.Context, diags *diag.Diagnostics) {
	if w.meta == nil {
		return
	}

	if err := CleanupOrphanedDefaultsOnMeta(ctx, w.meta); err != nil {
		diags.AddWarning("Failed to clean up orphaned default annotation keys and values", err.Error())
	}
}

// hasComputedSrn checks whether the inner resource's schema has a "srn"
// attribute that is Computed (i.e. generated by the API, not user-provided).
// This is used to skip binding creation/deletion for resources whose "srn"
// is user-provided (Required) rather than the resource's own SRN.
// The result is cached after the first call.
func (w *frameworkBindingWrapper) hasComputedSrn(ctx context.Context) bool {
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

	if strAttr, ok := attr.(schema.StringAttribute); ok {
		w.srnComputed = strAttr.Computed

		return w.srnComputed
	}

	return false
}

// stateStringAttribute reads a string attribute from a Framework State. It
// returns an empty string if the attribute is missing (no "srn" in schema),
// null, or unknown — making it a safe no-op for resources without an SRN.
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

// Compile-time interface checks.
var (
	_ resource.Resource                     = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithConfigure        = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithImportState      = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithModifyPlan       = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithUpgradeState     = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithConfigValidators = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithValidateConfig   = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithMoveState        = (*frameworkBindingWrapper)(nil)

	_ resource.Resource                    = (*frameworkBindingIdentityWrapper)(nil)
	_ resource.ResourceWithIdentity        = (*frameworkBindingIdentityWrapper)(nil)
	_ resource.ResourceWithUpgradeIdentity = (*frameworkBindingIdentityWrapper)(nil)
)
