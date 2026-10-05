package annotations_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/annotations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDefaults_EmptyInput(t *testing.T) {
	dc, err := annotations.ResolveDefaults(context.Background(), nil, nil)
	require.NoError(t, err)
	assert.Nil(t, dc)
}

func TestResolveDefaults_EmptyMap(t *testing.T) {
	dc, err := annotations.ResolveDefaults(context.Background(), nil, map[string]string{})
	require.NoError(t, err)
	assert.Nil(t, dc)
}

func TestCreateDefaultBindings_NilConfig(t *testing.T) {
	result, err := annotations.CreateDefaultBindings(context.Background(), nil, "srn://test", nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestCreateDefaultBindings_EmptySrn(t *testing.T) {
	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", ValueID: "val-1"},
		},
	}
	result, err := annotations.CreateDefaultBindings(context.Background(), nil, "", dc)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestReadDefaultBindings_NilConfig(t *testing.T) {
	result, err := annotations.ReadDefaultBindings(context.Background(), nil, "srn://test", nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestReadDefaultBindings_EmptySrn(t *testing.T) {
	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", ValueID: "val-1"},
		},
	}
	result, err := annotations.ReadDefaultBindings(context.Background(), nil, "", dc)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestDeleteDefaultBindings_NilConfig(t *testing.T) {
	err := annotations.DeleteDefaultBindings(context.Background(), nil, "srn://test", nil)
	require.NoError(t, err)
}

func TestDeleteDefaultBindings_EmptySrn(t *testing.T) {
	dc := &meta.AnnotationDefaultsConfig{
		Items: []meta.AnnotationDefault{
			{KeyName: "env", ValueID: "val-1"},
		},
	}
	err := annotations.DeleteDefaultBindings(context.Background(), nil, "", dc)
	require.NoError(t, err)
}

func TestResolveDefaultsOnMeta_NoAnnotations(t *testing.T) {
	m := &meta.Meta{}
	err := annotations.ResolveDefaultsOnMeta(context.Background(), m)
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

	err := annotations.ResolveDefaultsOnMeta(context.Background(), m)
	require.NoError(t, err)
	assert.Equal(t, "key-1", m.DefaultAnnotationsConfig().Items[0].KeyID)
}

func TestWithDefaultBindings_WrapsCreateAndDelete(t *testing.T) {
	createCalled := false
	deleteCalled := false

	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"srn": {
				Type:     schema.TypeString,
				Computed: true,
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
	diags := wrapped.CreateContext(context.Background(), d, &meta.Meta{})
	assert.False(t, diags.HasError())
	assert.True(t, createCalled)

	// SRN should be set by the original Create, and the wrapper should not
	// have cleared it.
	assert.Equal(t, "srn://test/id-1", d.Get("srn"))

	diags = wrapped.DeleteContext(context.Background(), d, &meta.Meta{})
	assert.False(t, diags.HasError())
	assert.True(t, deleteCalled)
}

func TestWithDefaultBindings_NoSrnSchemaIsNoOp(t *testing.T) {
	createCalled := false

	r := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
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
	diags := wrapped.CreateContext(context.Background(), d, &meta.Meta{})
	assert.False(t, diags.HasError())
	assert.True(t, createCalled)
}
