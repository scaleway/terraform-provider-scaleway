package annotations

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// WithDefaultBindings wraps an SDKv2 resource so that, after a successful
// Create or Update, the default annotation bindings of the resource's SRN are
// reconciled and, before Delete, the default bindings are removed. Annotation
// failures only produce warnings and never fail the resource operation.
// Resources without a computed top-level "srn" attribute are returned
// unchanged.
// TODO: handle different error levels, set from a config attribute.
func WithDefaultBindings(r *schema.Resource) *schema.Resource {
	srnSchema, ok := r.SchemaMap()["srn"]
	if !ok || !srnSchema.Computed {
		return r
	}

	w := *r

	if w.CreateContext != nil {
		original := w.CreateContext

		w.CreateContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			diags := original(ctx, d, m)
			if !diags.HasError() {
				diags = append(diags, reconcileDefaultBindingsDiags(ctx, d, m)...)
			}

			return diags
		}
	}

	if w.UpdateContext != nil {
		original := w.UpdateContext

		w.UpdateContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			diags := original(ctx, d, m)
			if !diags.HasError() {
				diags = append(diags, reconcileDefaultBindingsDiags(ctx, d, m)...)
			}

			return diags
		}
	}

	if w.DeleteContext != nil {
		original := w.DeleteContext

		w.DeleteContext = func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			diags := deleteDefaultBindingsDiags(ctx, d, m)

			return append(diags, original(ctx, d, m)...)
		}
	}

	return &w
}

func reconcileDefaultBindingsDiags(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	srn, _ := d.Get("srn").(string)
	if srn == "" {
		return nil
	}

	metaVal, ok := m.(*meta.Meta)
	if !ok {
		return nil
	}

	err := ReconcileDefaultBindingsForSrn(ctx, metaVal, srn)
	if err == nil {
		return nil
	}

	severity := diag.Warning

	if IsResolutionError(err) {
		severity = diag.Error
	}

	return diag.Diagnostics{{
		Severity: severity,
		Summary:  "Failed to apply default annotations",
		Detail:   err.Error(),
	}}
}

func deleteDefaultBindingsDiags(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	srn, _ := d.Get("srn").(string)
	if srn == "" {
		return nil
	}

	metaVal, ok := m.(*meta.Meta)
	if !ok {
		return nil
	}

	if err := DeleteDefaultBindingsForSrn(ctx, metaVal, srn); err != nil {
		return diag.Diagnostics{{
			Severity: diag.Warning,
			Summary:  "Failed to remove default annotation bindings",
			Detail:   err.Error(),
		}}
	}

	return nil
}
