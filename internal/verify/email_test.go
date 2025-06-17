package verify_test

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func TestIsEmail(t *testing.T) {
	validateFunc := verify.IsEmail()

	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"invalid-email", false},
		{"", false},
	}

	for _, test := range tests {
		diags := validateFunc(test.email, cty.Path{})
		if (len(diags) == 0) != test.valid {
			t.Errorf("IsEmail() test failed for input %s, expected valid: %v, got errors: %v", test.email, test.valid, diags)
		}
	}
}

func TestIsEmailList(t *testing.T) {
	validateFunc := verify.IsEmailList()

	tests := []struct {
		emails []any
		valid  bool
	}{
		{[]any{"test@example.com", "test2@example.com"}, true},
		{[]any{"test@example.com", "invalid-email"}, false},
		{[]any{123, "test@example.com"}, false},
		{[]any{}, true},
	}

	for _, test := range tests {
		diags := validateFunc(test.emails, cty.Path{})
		if (len(diags) == 0) != test.valid {
			t.Errorf("IsEmailList() test failed for input %v, expected valid: %v, got errors: %v", test.emails, test.valid, diags)
		}
	}
}
