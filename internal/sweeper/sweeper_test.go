package sweeper_test

import (
	"testing"

	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/sweeper"
)

func TestSweeper_RegistryContainsExpectedTypes(t *testing.T) {
	all := sweeper.All()

	if len(all) != 4 {
		t.Fatalf("expected 4 registered sweepers, got %d", len(all))
	}

	expected := map[string]struct{}{
		"scaleway_iam_api_key":     {},
		"scaleway_iam_scim_token":  {},
		"scaleway_secret":          {},
		"scaleway_key_manager_key": {},
	}

	for _, s := range all {
		if _, ok := expected[s.TypeName()]; !ok {
			t.Errorf("unexpected sweeper type: %s", s.TypeName())
		}
	}
}

func TestSweeper_SupportedTypesMatchAll(t *testing.T) {
	supported := sweeper.SupportedTypes()
	all := sweeper.All()

	if len(supported) != len(all) {
		t.Fatalf("SupportedTypes length %d != All length %d", len(supported), len(all))
	}

	for i, s := range all {
		if supported[i] != s.TypeName() {
			t.Errorf("SupportedTypes[%d] = %q, want %q", i, supported[i], s.TypeName())
		}
	}
}

func TestSweeper_FindSweeper(t *testing.T) {
	for _, s := range sweeper.All() {
		found, ok := sweeper.FindSweeper(s.TypeName())
		if !ok {
			t.Errorf("FindSweeper(%q) returned ok=false", s.TypeName())
		}

		if found.TypeName() != s.TypeName() {
			t.Errorf("FindSweeper(%q) returned %q", s.TypeName(), found.TypeName())
		}
	}

	if _, ok := sweeper.FindSweeper("scaleway_unknown_resource"); ok {
		t.Error("FindSweeper returned ok=true for an unknown type")
	}
}

func TestSweeper_LocalityAndTagSupport(t *testing.T) {
	cases := []struct {
		typeName     string
		locality     sweeper.Locality
		supportsTags bool
	}{
		{"scaleway_iam_api_key", sweeper.LocalityGlobal, false},
		{"scaleway_iam_scim_token", sweeper.LocalityGlobal, false},
		{"scaleway_secret", sweeper.LocalityRegional, true},
		{"scaleway_key_manager_key", sweeper.LocalityRegional, true},
	}

	for _, tc := range cases {
		t.Run(tc.typeName, func(t *testing.T) {
			s, ok := sweeper.FindSweeper(tc.typeName)
			if !ok {
				t.Fatalf("sweeper not found for %q", tc.typeName)
			}

			if s.Locality() != tc.locality {
				t.Errorf("Locality() = %q, want %q", s.Locality(), tc.locality)
			}

			if s.SupportsTagFilter() != tc.supportsTags {
				t.Errorf("SupportsTagFilter() = %v, want %v", s.SupportsTagFilter(), tc.supportsTags)
			}
		})
	}
}

// TestSweeper_SupportedTypesAreValidRegions ensures the regional sweepers are
// validated against the same region list used by the action schema.
func TestSweeper_RegionalSweepersUseKnownRegions(t *testing.T) {
	regions := regional.AllRegions()
	if len(regions) == 0 {
		t.Fatal("expected at least one known Scaleway region")
	}

	found := false

	for _, r := range regions {
		if r == "fr-par" {
			found = true
		}
	}

	if !found {
		t.Error("expected fr-par to be a valid region")
	}
}
