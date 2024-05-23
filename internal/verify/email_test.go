package verify_test

import (
	"testing"

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
		_, errors := validateFunc(test.email, "email")
		if (len(errors) == 0) != test.valid {
			t.Errorf("IsEmail() test failed for input %s, expected valid: %v, got errors: %v", test.email, test.valid, errors)
		}
	}
}

func TestIsEmailList(t *testing.T) {
	validateFunc := verify.IsEmailList()

	tests := []struct {
		emails []interface{}
		valid  bool
	}{
		{[]interface{}{"test@example.com", "test2@example.com"}, true},
		{[]interface{}{"test@example.com", "invalid-email"}, false},
		{[]interface{}{123, "test@example.com"}, false},
		{[]interface{}{}, true},
	}

	for _, test := range tests {
		_, errors := validateFunc(test.emails, "emails")
		if (len(errors) == 0) != test.valid {
			t.Errorf("IsEmailList() test failed for input %v, expected valid: %v, got errors: %v", test.emails, test.valid, errors)
		}
	}
}
