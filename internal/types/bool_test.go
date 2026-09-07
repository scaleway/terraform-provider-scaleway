package types_test

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/stretchr/testify/assert"
)

type rawConfigStub struct {
	rawConfig cty.Value
}

func (s rawConfigStub) GetRawConfig() cty.Value {
	return s.rawConfig
}

func objectConfig(attributes map[string]cty.Value) cty.Value {
	if len(attributes) == 0 {
		return cty.EmptyObjectVal
	}

	return cty.ObjectVal(attributes)
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		rawConfig cty.Value
		expected  any
		name      string
		key       string
	}{
		{
			name:      "explicit true is reported",
			rawConfig: objectConfig(map[string]cty.Value{"assign_flexible_ip": cty.True}),
			key:       "assign_flexible_ip",
			expected:  true,
		},
		{
			name:      "explicit false is reported and not mistaken for an absent attribute",
			rawConfig: objectConfig(map[string]cty.Value{"assign_flexible_ip": cty.False}),
			key:       "assign_flexible_ip",
			expected:  false,
		},
		{
			name:      "attribute left null is absent",
			rawConfig: objectConfig(map[string]cty.Value{"assign_flexible_ip": cty.NullVal(cty.Bool)}),
			key:       "assign_flexible_ip",
			expected:  nil,
		},
		{
			name:      "attribute unknown at plan time is absent",
			rawConfig: objectConfig(map[string]cty.Value{"assign_flexible_ip": cty.UnknownVal(cty.Bool)}),
			key:       "assign_flexible_ip",
			expected:  nil,
		},
		{
			name:      "attribute missing from the configuration is absent",
			rawConfig: objectConfig(map[string]cty.Value{"name": cty.StringVal("lb")}),
			key:       "assign_flexible_ip",
			expected:  nil,
		},
		{
			name:      "null configuration is absent",
			rawConfig: cty.NullVal(cty.EmptyObject),
			key:       "assign_flexible_ip",
			expected:  nil,
		},
		{
			name:      "absent configuration is absent",
			rawConfig: cty.NilVal,
			key:       "assign_flexible_ip",
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, types.GetBool(rawConfigStub{rawConfig: tt.rawConfig}, tt.key))
		})
	}
}

// Explicit false must be preserved so the API does not fall back to its default.
func TestGetBoolExpandsExplicitFalseToPointer(t *testing.T) {
	rawConfig := objectConfig(map[string]cty.Value{"assign_flexible_ip": cty.False})

	assigned := types.ExpandBoolPtr(types.GetBool(rawConfigStub{rawConfig: rawConfig}, "assign_flexible_ip"))

	assert.NotNil(t, assigned)
	assert.False(t, *assigned)
}
