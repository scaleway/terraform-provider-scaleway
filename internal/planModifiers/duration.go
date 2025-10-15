package planModifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type Duration struct{}

func (d Duration) Description(ctx context.Context) string {
	return "Verify that two duration are equals"
}

func (d Duration) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d Duration) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	if request.StateValue == request.PlanValue {
		response.PlanValue = request.PlanValue
	}
}
