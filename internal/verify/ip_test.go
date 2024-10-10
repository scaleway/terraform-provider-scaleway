package verify_test

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
	"github.com/stretchr/testify/assert"
)

func TestValidateStandaloneIPorCIDRWithValidIPReturnNothing(t *testing.T) {
	for _, ip := range []string{"192.168.1.1", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "10.0.0.0/24", "2001:0db8:85a3::8a2e:0370:7334/64"} {
		diags := verify.IsStandaloneIPorCIDR()(ip, cty.Path{})
		assert.Empty(t, diags)
	}
}

func TestValidateStandaloneIPorCIDRWithInvalidIPReturnError(t *testing.T) {
	for _, ip := range []string{"10.0.0", "256.256.256.256", "2001::85a3::8a2e:0370:7334", "10.0.0.0/34"} {
		diags := verify.IsStandaloneIPorCIDR()(ip, cty.Path{})
		assert.Len(t, diags, 1)
	}
}
