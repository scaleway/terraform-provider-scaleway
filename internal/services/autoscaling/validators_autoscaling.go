package autoscaling

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// scalingPolicyTargetValidator validates that exactly one of fixed_size, cpu_target, or memory_target is set.
type scalingPolicyTargetValidator struct{}

func (v scalingPolicyTargetValidator) Description(_ context.Context) string {
	return "Exactly one of fixed_size, cpu_target, or memory_target must be set"
}

func (v scalingPolicyTargetValidator) MarkdownDescription(_ context.Context) string {
	return "Exactly one of `fixed_size`, `cpu_target`, or `memory_target` must be set"
}

func (v scalingPolicyTargetValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	setCount := 0
	attrs := req.ConfigValue.Attributes()

	for _, target := range []string{"fixed_size", "cpu_target", "memory_target"} {
		targetValue, targetSet := attrs[target].(types.Int32)
		if !targetSet {
			continue
		}

		if !targetValue.IsNull() && !targetValue.IsUnknown() {
			setCount++
		}
	}

	if setCount == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("scaling_policy"),
			"Missing scaling policy target",
			"Exactly one of fixed_size, cpu_target, or memory_target must be set.",
		)
	} else if setCount > 1 {
		resp.Diagnostics.AddAttributeError(
			path.Root("scaling_policy"),
			"Invalid scaling policy target",
			"Only one of fixed_size, cpu_target, or memory_target can be set.",
		)
	}
}
