package acctest_test

import (
	"encoding/json"
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestIsNilOrEmptySlice(t *testing.T) {
	tests := []struct {
		v    any
		name string
		want bool
	}{
		{nil, "nil", true},
		{[]any{}, "empty slice", true},
		{[]any{"a"}, "non-empty slice", false},
		{"", "empty string", false},
		{map[string]any{}, "empty map", false},
		{float64(0), "number", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := acctest.IsNilOrEmptySlice(tt.v); got != tt.want {
				t.Fatalf("isNilOrEmptySlice(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestCompareJSONBodies_NullVsEmptySlice(t *testing.T) {
	// Reproduces the failure seen in TestAccInstanceTemplateResource_Volumes and
	// TestAccAction_InstanceExportSnapshot_SBS: the cassette recorded "tags":[] but
	// the SDK now serializes an unset slice as "tags":null.
	cassetteBody := `{"volume_id":"ff23892f-61b7-486e-94c2-328e0bfe2d07","name":"tf-snapshot-eager-sanderson","project_id":"105bdce1-64c0-48ab-899d-868455867ecf","tags":[],"public":false}`
	requestBody := `{"volume_id":"ff23892f-61b7-486e-94c2-328e0bfe2d07","name":"tf-snapshot-jovial-franklin","project_id":"105bdce1-64c0-48ab-899d-868455867ecf","tags":null,"public":false}`

	var requestJSON, cassetteJSON map[string]any
	if err := json.Unmarshal([]byte(requestBody), &requestJSON); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if err := json.Unmarshal([]byte(cassetteBody), &cassetteJSON); err != nil {
		t.Fatalf("unmarshal cassette: %v", err)
	}

	// project_id is ignored by BodyMatcherIgnore; remove it to mirror the matcher.
	for _, key := range acctest.BodyMatcherIgnore {
		acctest.RemoveKeyRecursive(requestJSON, key)
		acctest.RemoveKeyRecursive(cassetteJSON, key)
	}

	if !acctest.CompareJSONBodies(requestJSON, cassetteJSON, false) {
		t.Fatal("expected null tags to match empty [] tags")
	}
}

func TestCompareJSONBodies_NonEmptyVsNullSliceStillFails(t *testing.T) {
	// A non-empty slice must NOT be treated as equal to null: this would be a
	// genuine mismatch and should not silently match a cassette interaction.
	requestJSON := map[string]any{"tags": []any{"a"}}
	cassetteJSON := map[string]any{"tags": nil}

	if acctest.CompareJSONBodies(requestJSON, cassetteJSON, false) {
		t.Fatal("expected non-empty slice to not match null")
	}
}

func TestCompareJSONFields_NestedNullVsEmptySlice(t *testing.T) {
	// Ensure the tolerance also works for nested values reached via
	// compareJSONFields (e.g. slices of slices).
	if !acctest.CompareJSONFields(nil, []any{}, false) {
		t.Fatal("expected nested nil to match empty []")
	}

	if acctest.CompareJSONFields([]any{"a"}, nil, false) {
		t.Fatal("expected nested non-empty slice to not match nil")
	}
}
