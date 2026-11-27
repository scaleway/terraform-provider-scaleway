package annotations

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	annotationsSDK "github.com/scaleway/scaleway-sdk-go/api/annotations/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSDKv2Resource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"srn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
		},
		CreateContext: func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			d.SetId("test-id")

			return diag.FromErr(d.Set("srn", testTargetSrn))
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			return nil
		},
		DeleteContext: func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			return nil
		},
	}
}

func TestWithDefaultBindingsSDKv2(t *testing.T) {
	t.Parallel()

	t.Run("returns the resource as-is without a computed srn", func(t *testing.T) {
		t.Parallel()

		r := &schema.Resource{Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
		}}

		assert.Same(t, r, WithDefaultBindings(r))

		r = &schema.Resource{Schema: map[string]*schema.Schema{
			"srn": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
		}}

		assert.Same(t, r, WithDefaultBindings(r))
	})

	t.Run("wraps a resource with a computed srn", func(t *testing.T) {
		t.Parallel()

		assert.NotSame(t, newTestSDKv2Resource(), WithDefaultBindings(newTestSDKv2Resource()))
	})

	t.Run("create without default annotations is a no-op", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{}
		m := newTestMeta(t, server, nil)
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{})

		diags := wrapped.CreateContext(t.Context(), d, m)

		assert.False(t, diags.HasError())
		assert.Empty(t, diags)
		assert.Empty(t, server.created)
	})

	t.Run("create without meta is a no-op", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{}
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{})

		diags := wrapped.CreateContext(t.Context(), d, "not-a-meta")

		assert.False(t, diags.HasError())
		assert.Empty(t, diags)
		assert.Empty(t, server.created)
	})

	t.Run("create reconciles the default annotations", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")}}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{})

		diags := wrapped.CreateContext(t.Context(), d, m)

		assert.False(t, diags.HasError())

		assertCreatedBinding(t, server, testValueName("team", "platform"))
	})

	t.Run("create reconciles bindings spread over multiple pages", func(t *testing.T) {
		t.Parallel()

		// The stale binding is on the second page: it can only be found (and
		// deleted) if all pages are fetched.
		server := &fakeAnnotationsServer{
			keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")},
			bindings: []*annotationsSDK.Binding{
				testBinding("b-env", "env", "platform"),
				testBinding("b-other", "other", "platform"),
				testBinding("b-stale", "team", "legacy"),
			},
		}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{})

		diags := wrapped.CreateContext(t.Context(), d, m)

		assert.False(t, diags.HasError())

		assert.Equal(t, []string{"b-stale"}, server.deleted)

		assertCreatedBinding(t, server, testValueName("team", "platform"))
	})

	t.Run("create warns when a key cannot be resolved", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{}
		m := newTestMeta(t, server, map[string]string{"missing": "platform"})
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{})

		diags := wrapped.CreateContext(t.Context(), d, m)

		assert.False(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, diag.Warning, diags[0].Severity)
		assert.Contains(t, diags[0].Summary, "Failed to apply default annotations")
		assert.Contains(t, diags[0].Detail, `annotation key "missing" not found`)
	})

	t.Run("create warns when the binding creation fails", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{
			keys:        []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")},
			createError: http.StatusInternalServerError,
		}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{})

		diags := wrapped.CreateContext(t.Context(), d, m)

		require.False(t, diags.HasError())
		require.Len(t, diags, 1)
		assert.Equal(t, diag.Warning, diags[0].Severity)
		assert.Contains(t, diags[0].Summary, "Failed to apply default annotations")
	})

	t.Run("update reconciles the default annotations", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{keys: []*annotationsSDK.ListAllKeysAndValuesResponseKey{testKey("team")}}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{"srn": testTargetSrn})

		diags := wrapped.UpdateContext(t.Context(), d, m)

		assert.False(t, diags.HasError())

		assertCreatedBinding(t, server, testValueName("team", "platform"))
	})

	t.Run("update without default annotations is a no-op", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{}
		m := newTestMeta(t, server, nil)
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{"srn": testTargetSrn})

		diags := wrapped.UpdateContext(t.Context(), d, m)

		assert.False(t, diags.HasError())
		assert.Empty(t, diags)
		assert.Empty(t, server.created)
	})

	t.Run("delete removes the default bindings", func(t *testing.T) {
		t.Parallel()

		server := &fakeAnnotationsServer{bindings: []*annotationsSDK.Binding{testBinding("b-team", "team", "platform")}}
		m := newTestMeta(t, server, map[string]string{"team": "platform"})
		wrapped := WithDefaultBindings(newTestSDKv2Resource())
		d := schema.TestResourceDataRaw(t, newTestSDKv2Resource().Schema, map[string]any{"srn": testTargetSrn})

		diags := wrapped.DeleteContext(t.Context(), d, m)

		assert.False(t, diags.HasError())

		assert.Equal(t, []string{"b-team"}, server.deleted)
	})
}
