package annotations_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDefaults_EmptyInput(t *testing.T) {
	dc, err := annotations.ResolveDefaults(t.Context(), nil, nil)
	require.NoError(t, err)
	assert.Nil(t, dc)
}

func TestResolveDefaults_EmptyMap(t *testing.T) {
	dc, err := annotations.ResolveDefaults(t.Context(), nil, map[string]string{})
	require.NoError(t, err)
	assert.Nil(t, dc)
}

func TestCreateDefaultBindings_NilConfig(t *testing.T) {
	err := annotations.CreateDefaultBindings(t.Context(), nil, "srn://test", nil)
	require.NoError(t, err)
}

func TestCreateDefaultBindings_EmptySrn(t *testing.T) {
	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", ValueID: "val-1"},
		},
	}
	err := annotations.CreateDefaultBindings(t.Context(), nil, "", dc)
	require.NoError(t, err)
}

func TestReconcileDefaultBindings_NilConfig(t *testing.T) {
	err := annotations.ReconcileDefaultBindings(t.Context(), nil, "srn://test", nil)
	require.NoError(t, err)
}

func TestReconcileDefaultBindings_EmptySrn(t *testing.T) {
	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", ValueID: "val-1"},
		},
	}
	err := annotations.ReconcileDefaultBindings(t.Context(), nil, "", dc)
	require.NoError(t, err)
}

func TestDeleteDefaultBindings_EmptySrn(t *testing.T) {
	err := annotations.DeleteDefaultBindings(t.Context(), nil, "", nil)
	require.NoError(t, err)
}

func TestResolveDefaultsOnMeta_NoAnnotations(t *testing.T) {
	m := &meta.Meta{}
	err := annotations.ResolveDefaultsOnMeta(t.Context(), m)
	require.NoError(t, err)
	assert.Nil(t, m.DefaultAnnotationsConfig())
}

func TestResolveDefaultsOnMeta_AlreadyResolved(t *testing.T) {
	m := &meta.Meta{}
	m.SetDefaultAnnotations(map[string]string{"env": "dev"})
	m.SetDefaultAnnotationsConfig(&meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "dev", ValueID: "val-1"},
		},
	})

	err := annotations.ResolveDefaultsOnMeta(t.Context(), m)
	require.NoError(t, err)
	assert.Equal(t, "key-1", m.DefaultAnnotationsConfig().Items[0].KeyID)
}

func TestWithDefaultBindings_WrapsCreateAndDelete(t *testing.T) {
	createCalled := false
	deleteCalled := false

	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"srn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the resource",
			},
		},
		CreateContext: func(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
			createCalled = true

			d.SetId("id-1")
			_ = d.Set("srn", "srn://test/id-1")

			return nil
		},
		DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
			deleteCalled = true

			return nil
		},
	}

	wrapped := annotations.WithDefaultBindings(r)

	// With no default annotations configured, the wrapper is a no-op for
	// bindings but still calls the original Create/Delete.
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	diags := wrapped.CreateContext(t.Context(), d, &meta.Meta{})
	assert.False(t, diags.HasError())
	assert.True(t, createCalled)

	// SRN should be set by the original Create, and the wrapper should not
	// have cleared it.
	assert.Equal(t, "srn://test/id-1", d.Get("srn"))

	diags = wrapped.DeleteContext(t.Context(), d, &meta.Meta{})
	assert.False(t, diags.HasError())
	assert.True(t, deleteCalled)
}

func TestWithDefaultBindings_NoSrnSchemaIsNoOp(t *testing.T) {
	createCalled := false

	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the resource",
			},
		},
		CreateContext: func(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
			createCalled = true

			d.SetId("id-1")

			return nil
		},
		DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
			return nil
		},
	}

	wrapped := annotations.WithDefaultBindings(r)

	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	diags := wrapped.CreateContext(t.Context(), d, &meta.Meta{})
	assert.False(t, diags.HasError())
	assert.True(t, createCalled)
}

func TestWithDefaultBindings_DoubleWrapIsNoOp(t *testing.T) {
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"srn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the resource",
			},
		},
	}

	wrapped := annotations.WithDefaultBindings(r)

	// The wrapper mutates the resource in place, so the first wrap returns
	// the same pointer.
	assert.Same(t, r, wrapped)

	// Wrapping the same pointer a second time must be a no-op.
	assert.Same(t, r, annotations.WithDefaultBindings(r))
}

func TestValidateAnnotationKeyName(t *testing.T) {
	validNames := []string{"env", "_private", "team2", "Accès", strings.Repeat("é", annotations.MaxAnnotationKeyNameLength)}
	for _, name := range validNames {
		require.NoError(t, annotations.ValidateAnnotationKeyName(name), "name %q should be valid", name)
	}

	invalidNames := []string{
		"",
		"1abc",
		"-abc",
		"a-b",
		"a b",
		strings.Repeat("a", annotations.MaxAnnotationKeyNameLength+1),
		strings.Repeat("é", annotations.MaxAnnotationKeyNameLength+1),
	}
	for _, name := range invalidNames {
		require.Error(t, annotations.ValidateAnnotationKeyName(name), "name %q should be invalid", name)
	}
}

func TestValidateAnnotationValueName(t *testing.T) {
	validNames := []string{"prod", "_internal", "value42"}
	for _, name := range validNames {
		require.NoError(t, annotations.ValidateAnnotationValueName(name), "name %q should be valid", name)
	}

	invalidNames := []string{
		"",
		"9lives",
		"-value",
		"va lue",
		strings.Repeat("a", annotations.MaxAnnotationValueNameLength+1),
	}
	for _, name := range invalidNames {
		require.Error(t, annotations.ValidateAnnotationValueName(name), "name %q should be invalid", name)
	}
}

func TestCleanupOrphanedDefaultAnnotations_NilClient(t *testing.T) {
	require.Error(t, annotations.CleanupOrphanedDefaultAnnotations(t.Context(), nil, nil))
}

func TestCleanupOrphanedDefaultsOnMeta_NoClient(t *testing.T) {
	m := &meta.Meta{}
	m.SetDefaultAnnotations(map[string]string{"env": "dev"})

	require.Error(t, annotations.CleanupOrphanedDefaultsOnMeta(t.Context(), m))
}

func TestCleanupOrphanedDefaultsOnMeta_SkipsWhenNotConfigured(t *testing.T) {
	m := &meta.Meta{}

	require.NoError(t, annotations.CleanupOrphanedDefaultsOnMeta(t.Context(), m))
}

func TestWithDefaultBindings_ReachesBindingPathOnCreate(t *testing.T) {
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"srn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the resource",
			},
		},
		CreateContext: func(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
			d.SetId("id-1")
			_ = d.Set("srn", "srn://test/id-1")

			return nil
		},
		DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
			return nil
		},
	}

	wrapped := annotations.WithDefaultBindings(r)

	m := &meta.Meta{}
	m.SetDefaultAnnotationsConfig(&meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-1"},
		},
	})

	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})

	// With DefaultAnnotationsConfig set and SRN non-empty, the wrapper
	// reaches ReconcileDefaultBindingsFromMeta, which calls the annotations
	// API with a nil scw.Client and panics. This proves the binding code
	// path was not short-circuited by a nil config guard.
	assert.Panics(t, func() {
		_ = wrapped.CreateContext(t.Context(), d, m)
	})
}

func TestWithDefaultBindings_ReachesBindingPathOnDelete(t *testing.T) {
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"srn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Scaleway Resource Name (SRN) of the resource",
			},
		},
		CreateContext: func(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
			d.SetId("id-1")
			_ = d.Set("srn", "srn://test/id-1")

			return nil
		},
		DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
			return nil
		},
	}

	wrapped := annotations.WithDefaultBindings(r)

	m := &meta.Meta{}
	m.SetDefaultAnnotationsConfig(&meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", KeyID: "key-1", ValueName: "prod", ValueID: "val-1"},
		},
	})

	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{})
	_ = d.Set("srn", "srn://test/id-1")

	// With DefaultAnnotationsConfig set and SRN non-empty, the wrapper
	// reaches DeleteDefaultBindingsFromMeta, which calls the annotations
	// API with a nil scw.Client and panics. This proves the binding delete
	// path was reached.
	assert.Panics(t, func() {
		_ = wrapped.DeleteContext(t.Context(), d, m)
	})
}
