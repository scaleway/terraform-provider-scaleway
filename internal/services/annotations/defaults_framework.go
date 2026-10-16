package annotations

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// frameworkBindingWrapper wraps a Framework resource.Resource so that default
// annotation bindings are created after Create and deleted before Delete,
// based on the resource's "srn" attribute.
//
// The wrapper transparently delegates all Framework optional interfaces
// (Configure, ImportState, Identity, ModifyPlan, etc.) to the inner resource
// when it implements them. Resources without a Computed "srn" attribute (e.g.
// the annotations binding resource whose "srn" is Required) or with an empty
// SRN are no-ops.
type frameworkBindingWrapper struct {
	inner       resource.Resource
	meta        *meta.Meta
	srnComputed bool
	srnChecked  bool
}

// WithDefaultBindingsFramework wraps a Framework resource factory so that
// default annotation bindings are created on Create and deleted on Delete for
// any resource that exposes a Computed "srn" attribute. It is the Framework
// counterpart to WithDefaultBindings (which targets SDKv2 resources).
func WithDefaultBindingsFramework(factory func() resource.Resource) func() resource.Resource {
	return func() resource.Resource {
		return &frameworkBindingWrapper{inner: factory()}
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
	w.inner.Create(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		return
	}

	if !w.hasComputedSrn(ctx) {
		return
	}

	srn := stateStringAttribute(ctx, &resp.State, "srn")
	if srn == "" {
		return
	}

	if _, err := CreateDefaultBindingsFromMeta(ctx, w.meta, srn); err != nil {
		resp.Diagnostics.AddError("Failed to create default annotation bindings", err.Error())
	}
}

func (w *frameworkBindingWrapper) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	w.inner.Read(ctx, req, resp)
}

func (w *frameworkBindingWrapper) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	w.inner.Update(ctx, req, resp)
}

func (w *frameworkBindingWrapper) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if w.hasComputedSrn(ctx) {
		srn := stateStringAttribute(ctx, &req.State, "srn")
		if srn != "" {
			if err := DeleteDefaultBindingsFromMeta(ctx, w.meta, srn); err != nil {
				resp.Diagnostics.AddError("Failed to delete default annotation bindings", err.Error())

				return
			}
		}
	}

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

func (w *frameworkBindingWrapper) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	if r, ok := w.inner.(resource.ResourceWithIdentity); ok {
		r.IdentitySchema(ctx, req, resp)

		return
	}

	resp.Diagnostics.AddError(
		"Resource Identity Not Implemented",
		"This resource does not support identity.",
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

func (w *frameworkBindingWrapper) UpgradeIdentity(ctx context.Context) map[int64]resource.IdentityUpgrader {
	if r, ok := w.inner.(resource.ResourceWithUpgradeIdentity); ok {
		return r.UpgradeIdentity(ctx)
	}

	return nil
}

// hasComputedSrn checks whether the inner resource's schema has a "srn"
// attribute that is Computed (i.e. generated by the API, not user-provided).
// This is used to skip binding creation/deletion for resources like the
// annotations binding resource, whose "srn" is Required (user-provided).
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
	_ resource.ResourceWithIdentity         = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithModifyPlan       = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithUpgradeState     = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithConfigValidators = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithValidateConfig   = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithMoveState        = (*frameworkBindingWrapper)(nil)
	_ resource.ResourceWithUpgradeIdentity  = (*frameworkBindingWrapper)(nil)
)
