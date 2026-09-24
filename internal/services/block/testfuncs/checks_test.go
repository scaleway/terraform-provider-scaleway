package blocktestfuncs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestMatchAttrPairIgnoreCase(t *testing.T) {
	mockState := &terraform.State{
		Modules: []*terraform.ModuleState{
			{
				Path: []string{"root"}, // RootModule() looks for this path
				Resources: map[string]*terraform.ResourceState{
					"example_resource.match_1": {
						Primary: &terraform.InstanceState{
							Attributes: map[string]string{"attr": "fr-par-1/6b0bf87b-a911-4a0b-beb6-10c250c2ce1f"},
						},
					},
					"example_resource.match_2": {
						Primary: &terraform.InstanceState{
							Attributes: map[string]string{"attr": "6b0bf87b-a911-4a0b-beb6-10c250c2ce1f"},
						},
					},
					"example_resource.mismatch": {
						Primary: &terraform.InstanceState{
							Attributes: map[string]string{"attr": "6b0bf87b-a911-4a0b-beb6-10c250c2ce1e"},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name        string
		resFirst    string
		resSecond   string
		expectError bool
	}{
		{
			name:        "Successful case-insensitive match",
			resFirst:    "example_resource.match_1",
			resSecond:   "example_resource.match_2",
			expectError: false,
		},
		{
			name:        "Values do not match",
			resFirst:    "example_resource.match_1",
			resSecond:   "example_resource.mismatch",
			expectError: true,
		},
		{
			name:        "Resource missing from state",
			resFirst:    "example_resource.does_not_exist",
			resSecond:   "example_resource.match_1",
			expectError: true,
		},
	}

	// 3. Run the test cases against your helper function
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			checkFunc := blocktestfuncs.MatchAttrPairIgnorePrefix(tc.resFirst, "attr", tc.resSecond, "attr")

			err := checkFunc(mockState)

			if tc.expectError && err == nil {
				t.Fatalf("expected an error but got none")
			}

			if !tc.expectError && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}
		})
	}
}
