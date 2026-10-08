package tags

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

// FrameworkModifyPlanTagsAll recomputes the tags_all attribute in the
// planned state from the planned tags and the provider's default tags.
// It should be called from a Framework resource's ModifyPlan method.
//
// When tags or default tags change, tags_all is updated so the plan
// reflects the new merged set and triggers an Update if needed.
func FrameworkModifyPlanTagsAll(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, m *meta.Meta) {
	if m == nil {
		return
	}

	// If the plan is null (e.g., during resource deletion), do not modify it.
	if req.Plan.Raw.IsNull() {
		return
	}

	defaultTags := m.DefaultTags()

	// Get the planned tags value.
	var planTags types.List
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("tags"), &planTags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the state tags_all value (if any).
	var stateTagsAll types.List
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("tags_all"), &stateTagsAll)...)

	// Compute the new tags_all from planned tags + default tags.
	resourceTags := frameworkListToStrings(ctx, planTags)
	allTags := MergeTags(defaultTags, resourceTags)
	newTagsAll := stringsToFrameworkList(ctx, allTags)

	// If the new tags_all is the same as state, no change needed.
	if stateTagsAll.Equal(newTagsAll) {
		return
	}

	// Set the new tags_all in the plan.
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("tags_all"), newTagsAll)...)
}

// FrameworkSetTagsAllAndTags sets tags_all and tags in a Framework state model
// from the full tag list returned by the API. tags_all receives the complete
// list; tags receives the list with default tags removed.
func FrameworkSetTagsAllAndTags(ctx context.Context, allTagsFromAPI []string, defaultTags []string, tags *types.List, tagsAll *types.List) {
	*tagsAll = stringsToFrameworkList(ctx, allTagsFromAPI)
	resourceTags := RemoveDefaultTags(allTagsFromAPI, defaultTags)
	*tags = stringsToFrameworkList(ctx, resourceTags)
}

// FrameworkExpandTagsAll merges default tags with resource tags from a
// Framework types.List, returning the combined slice for API requests.
func FrameworkExpandTagsAll(ctx context.Context, planTags types.List, defaultTags []string) []string {
	resourceTags := frameworkListToStrings(ctx, planTags)
	return MergeTags(defaultTags, resourceTags)
}

func frameworkListToStrings(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	_ = list.ElementsAs(ctx, &result, false)
	return result
}

func stringsToFrameworkList(_ context.Context, tags []string) types.List {
	elems := make([]attr.Value, 0, len(tags))
	for _, t := range tags {
		elems = append(elems, types.StringValue(t))
	}
	list, _ := types.ListValue(types.StringType, elems)
	return list
}
