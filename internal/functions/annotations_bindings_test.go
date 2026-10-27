package functions_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/functions"
)

var (
	testBindingAttrTypes = map[string]attr.Type{
		"value_id":   types.StringType,
		"target_srn": types.StringType,
	}

	testBindingObjectType = basetypes.ObjectType{AttrTypes: testBindingAttrTypes}
)

func testBindingEntry(valueID, targetSRN string) attr.Value {
	return testBindingEntryValues(types.StringValue(valueID), types.StringValue(targetSRN))
}

func testBindingEntryValues(valueID, targetSRN attr.Value) attr.Value {
	entry, _ := basetypes.NewObjectValue(testBindingAttrTypes, map[string]attr.Value{
		"value_id":   valueID,
		"target_srn": targetSRN,
	})

	return entry
}

func testBindingMapValue(elements map[string]attr.Value) types.Map {
	m, _ := basetypes.NewMapValue(testBindingObjectType, elements)

	return m
}

func TestAnnotationsBindingsFunctionRun(t *testing.T) {
	t.Parallel()

	values := func(elements map[string]attr.Value) attr.Value {
		return types.MapValueMust(types.StringType, elements)
	}

	testCases := map[string]struct {
		expected function.RunResponse
		request  function.RunRequest
	}{
		"null-values": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{types.MapNull(types.StringType), types.MapValueMust(types.StringType, map[string]attr.Value{"vpc": types.StringValue("srn-1")})}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.MapNull(testBindingObjectType)),
			},
		},
		"unknown-targets": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{values(map[string]attr.Value{"environment": types.StringValue("vid-1")}), types.MapUnknown(types.StringType)}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(types.MapUnknown(testBindingObjectType)),
			},
		},
		"empty-targets": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{values(map[string]attr.Value{"environment": types.StringValue("vid-1")}), values(map[string]attr.Value{})}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(testBindingMapValue(map[string]attr.Value{})),
			},
		},
		"single": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{values(map[string]attr.Value{"environment": types.StringValue("vid-1")}), values(map[string]attr.Value{"vpc": types.StringValue("srn-1")})}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(testBindingMapValue(map[string]attr.Value{
					"environment|vpc": testBindingEntry("vid-1", "srn-1"),
				})),
			},
		},
		"cross-product": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{values(map[string]attr.Value{
					"environment": types.StringValue("vid-1"),
					"team":        types.StringValue("vid-2"),
				}), values(map[string]attr.Value{
					"vpc": types.StringValue("srn-1"),
					"pn":  types.StringValue("srn-2"),
				})}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(testBindingMapValue(map[string]attr.Value{
					"environment|vpc": testBindingEntry("vid-1", "srn-1"),
					"environment|pn":  testBindingEntry("vid-1", "srn-2"),
					"team|vpc":        testBindingEntry("vid-2", "srn-1"),
					"team|pn":         testBindingEntry("vid-2", "srn-2"),
				})),
			},
		},
		"unknown-value-id": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{values(map[string]attr.Value{"environment": types.StringUnknown()}), values(map[string]attr.Value{"vpc": types.StringValue("srn-1")})}),
			},
			expected: function.RunResponse{
				Result: function.NewResultData(testBindingMapValue(map[string]attr.Value{
					"environment|vpc": testBindingEntryValues(types.StringUnknown(), types.StringValue("srn-1")),
				})),
			},
		},
		"conflicting-keys": {
			request: function.RunRequest{
				Arguments: function.NewArgumentsData([]attr.Value{values(map[string]attr.Value{
					"a":   types.StringValue("vid-1"),
					"a|b": types.StringValue("vid-2"),
				}), values(map[string]attr.Value{
					"b|c": types.StringValue("srn-1"),
					"c":   types.StringValue("srn-2"),
				})}),
			},
			expected: function.RunResponse{
				Error:  function.NewFuncError(`conflicting binding key "a|b|c": distinct (value, target) pairs map to the same key`),
				Result: function.NewResultData(types.MapUnknown(testBindingObjectType)),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := function.RunResponse{
				Result: function.NewResultData(types.MapUnknown(testBindingObjectType)),
			}

			functions.NewAnnotationsBindings().Run(context.Background(), testCase.request, &got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestAccProviderFunction_Annotations_Bindings(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				output "bindings" {
					value = provider::scaleway::annotations_bindings(
						{
							"environment" = "value-id-production"
							"team"        = "value-id-team"
						},
						{
							"scaleway_vpc.core" = "srn://scaleway-vpc/core"
							"scaleway_vpc.app"  = "srn://scaleway-vpc/app"
						},
					)
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValueAtPath("bindings", tfjsonpath.New("environment|scaleway_vpc.core").AtMapKey("value_id"), knownvalue.StringExact("value-id-production")),
					statecheck.ExpectKnownOutputValueAtPath("bindings", tfjsonpath.New("environment|scaleway_vpc.core").AtMapKey("target_srn"), knownvalue.StringExact("srn://scaleway-vpc/core")),
					statecheck.ExpectKnownOutputValueAtPath("bindings", tfjsonpath.New("environment|scaleway_vpc.app").AtMapKey("value_id"), knownvalue.StringExact("value-id-production")),
					statecheck.ExpectKnownOutputValueAtPath("bindings", tfjsonpath.New("environment|scaleway_vpc.app").AtMapKey("target_srn"), knownvalue.StringExact("srn://scaleway-vpc/app")),
					statecheck.ExpectKnownOutputValueAtPath("bindings", tfjsonpath.New("team|scaleway_vpc.core").AtMapKey("value_id"), knownvalue.StringExact("value-id-team")),
					statecheck.ExpectKnownOutputValueAtPath("bindings", tfjsonpath.New("team|scaleway_vpc.core").AtMapKey("target_srn"), knownvalue.StringExact("srn://scaleway-vpc/core")),
					statecheck.ExpectKnownOutputValueAtPath("bindings", tfjsonpath.New("team|scaleway_vpc.app").AtMapKey("value_id"), knownvalue.StringExact("value-id-team")),
					statecheck.ExpectKnownOutputValueAtPath("bindings", tfjsonpath.New("team|scaleway_vpc.app").AtMapKey("target_srn"), knownvalue.StringExact("srn://scaleway-vpc/app")),
				},
			},
		},
	})
}
