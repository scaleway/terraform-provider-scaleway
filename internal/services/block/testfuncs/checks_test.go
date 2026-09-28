package blocktestfuncs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	blocktestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block/testfuncs"
)

func TestMatchAttrPairIgnorePrefix(t *testing.T) {
	mockState := &terraform.State{
		Modules: []*terraform.ModuleState{
			{
				Path: []string{"root"}, // RootModule() looks for this path
				Resources: map[string]*terraform.ResourceState{
					"example_resource.prefix_first": {
						Primary: &terraform.InstanceState{
							Attributes: map[string]string{"attr": "fr-par-1/6b0bf87b-a911-4a0b-beb6-10c250c2ce1f"},
						},
					},
					"example_resource.prefix_second": {
						Primary: &terraform.InstanceState{
							Attributes: map[string]string{"attr": "6b0bf87b-a911-4a0b-beb6-10c250c2ce1f"},
						},
					},
					"example_resource.prefix_both_first": {
						Primary: &terraform.InstanceState{
							Attributes: map[string]string{"attr": "fr-par-1/6b0bf87b-a911-4a0b-beb6-10c250c2ce1f"},
						},
					},
					"example_resource.prefix_both_second": {
						Primary: &terraform.InstanceState{
							Attributes: map[string]string{"attr": "fr-par-1/6b0bf87b-a911-4a0b-beb6-10c250c2ce1f"},
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
			name:        "Successful match when first has prefix and second does not",
			resFirst:    "example_resource.prefix_first",
			resSecond:   "example_resource.prefix_second",
			expectError: false,
		},
		{
			name:        "Successful match when second has prefix and first does not",
			resFirst:    "example_resource.prefix_second",
			resSecond:   "example_resource.prefix_first",
			expectError: false,
		},
		{
			name:        "Successful match when both have identical values with prefix",
			resFirst:    "example_resource.prefix_both_first",
			resSecond:   "example_resource.prefix_both_second",
			expectError: false,
		},
		{
			name:        "Values do not match",
			resFirst:    "example_resource.prefix_first",
			resSecond:   "example_resource.mismatch",
			expectError: true,
		},
		{
			name:        "Resource missing from state",
			resFirst:    "example_resource.does_not_exist",
			resSecond:   "example_resource.prefix_first",
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
