package types_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenFrameworkStringValue(t *testing.T) {
	t.Run("empty string returns null", func(t *testing.T) {
		result := scwtypes.FlattenFrameworkStringValue("")
		assert.True(t, result.IsNull(), "expected null for empty string")
	})

	t.Run("non-empty string returns value", func(t *testing.T) {
		result := scwtypes.FlattenFrameworkStringValue("hello")
		assert.False(t, result.IsNull())
		assert.Equal(t, "hello", result.ValueString())
	})
}

func TestFlattenFrameworkStringList(t *testing.T) {
	ctx := context.Background()

	t.Run("nil slice returns null list", func(t *testing.T) {
		result, diags := scwtypes.FlattenFrameworkStringList(ctx, nil)
		require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
		assert.True(t, result.IsNull(), "expected null for nil slice")
	})

	t.Run("empty slice returns null list", func(t *testing.T) {
		result, diags := scwtypes.FlattenFrameworkStringList(ctx, []string{})
		require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
		assert.True(t, result.IsNull(), "expected null for empty slice")
	})

	t.Run("populated slice returns list value", func(t *testing.T) {
		items := []string{"a", "b", "c"}
		result, diags := scwtypes.FlattenFrameworkStringList(ctx, items)
		require.False(t, diags.HasError(), "unexpected diagnostics: %s", diags)
		assert.False(t, result.IsNull())
		assert.Len(t, result.Elements(), 3)

		var got []string

		diags = result.ElementsAs(ctx, &got, false)
		require.False(t, diags.HasError())
		assert.Equal(t, items, got)
	})

	t.Run("single element slice", func(t *testing.T) {
		items := []string{"only"}
		result, diags := scwtypes.FlattenFrameworkStringList(ctx, items)
		require.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.Len(t, result.Elements(), 1)

		expected := types.StringValue("only")
		assert.Equal(t, expected, result.Elements()[0])
	})
}
