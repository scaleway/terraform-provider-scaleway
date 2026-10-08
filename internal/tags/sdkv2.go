package tags

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

// TagsAllSchema returns the schema definition for the computed tags_all
// attribute. This attribute holds the complete set of tags applied to a
// resource: provider-level default tags merged with resource-specific tags.
//
// Every tag-compatible SDKv2 resource should include this attribute alongside
// its existing tags attribute.
func TagsAllSchema() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Computed:    true,
		Description: "The tags associated with the resource, including default tags from the provider configuration.",
	}
}

// ExtractDefaultConfig retrieves the provider's DefaultConfig from the meta
// value passed to SDKv2 resources. Returns nil when no default tags are
// configured.
func ExtractDefaultConfig(m any) *DefaultConfig {
	defaultTags := meta.ExtractDefaultTags(m)
	if len(defaultTags) == 0 {
		return nil
	}

	return &DefaultConfig{Tags: defaultTags}
}

// CustomizeDiffTagsAll is an SDKv2 CustomizeDiffFunc that recomputes the
// tags_all attribute from the resource's tags attribute and the provider's
// default tags. It should be added to every tag-compatible resource's
// CustomizeDiff chain.
//
// When the planned tags contain unknown values, tags_all is marked as
// computed so it will be resolved during apply.
func CustomizeDiffTagsAll(_ context.Context, d *schema.ResourceDiff, m any) error {
	dc := ExtractDefaultConfig(m)

	// If tags is not wholly known in the plan, defer to apply.
	if planTags := d.GetRawPlan(); !planTags.IsNull() && planTags.IsKnown() {
		if tagsAttr := planTags.GetAttr("tags"); tagsAttr.IsKnown() && !tagsAttr.IsNull() {
			if !tagsAttr.IsWhollyKnown() {
				return d.SetNewComputed("tags_all")
			}
		}
	}

	resourceTags := types.ExpandStrings(d.Get("tags"))
	allTags := dc.MergeTags(resourceTags)

	return d.SetNew("tags_all", allTags)
}

// ExpandTagsAll returns the merged set of default tags and the resource's
// configured tags, suitable for sending to the Scaleway API on Create.
func ExpandTagsAll(d *schema.ResourceData, m any) []string {
	dc := ExtractDefaultConfig(m)
	resourceTags := types.ExpandStrings(d.Get("tags"))

	return dc.MergeTags(resourceTags)
}

// ExpandTagsAllPtr is like ExpandTagsAll but returns a pointer, suitable
// for API request structs that use *[]string. Returns nil when the result
// is empty.
func ExpandTagsAllPtr(d *schema.ResourceData, m any) *[]string {
	allTags := ExpandTagsAll(d, m)
	if len(allTags) == 0 {
		return nil
	}

	return &allTags
}

// ExpandTagsAllUpdatedPtr is like ExpandTagsAll but always returns a
// non-nil pointer (even for an empty list), suitable for Update requests
// where an empty list must be sent to clear tags.
func ExpandTagsAllUpdatedPtr(d *schema.ResourceData, m any) *[]string {
	allTags := ExpandTagsAll(d, m)

	return &allTags
}

// SetTagsAllAndTags sets both the tags_all and tags attributes from the
// full tag list returned by the API. tags_all receives the complete list;
// tags receives the list with default tags removed (resource-specific only).
//
// This should be called in every tag-compatible resource's Read function.
func SetTagsAllAndTags(d *schema.ResourceData, m any, apiTags []string) {
	dc := ExtractDefaultConfig(m)

	_ = d.Set("tags_all", apiTags)

	resourceTags := dc.RemoveDefaultTags(apiTags)
	if len(resourceTags) > 0 {
		_ = d.Set("tags", resourceTags)
	} else {
		_ = d.Set("tags", []string{})
	}
}
