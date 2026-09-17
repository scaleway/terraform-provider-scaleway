package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

// ExpandStringList builds a Go type string array from a Terraform types.List
func ExpandStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)

	return result
}

// ExpandUpdatedStringList builds a Go type string array from a Terraform types.List.
// If list is nil, returns a pointer on an empty array to trigger the update of the field in the request.
func ExpandUpdatedStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	result := make([]string, 0)

	if list.IsNull() || list.IsUnknown() {
		return result
	}

	diags.Append(list.ElementsAs(ctx, &result, false)...)

	return result
}

// FlattenStringList builds a Terraform types.List from a Go type string array.
func FlattenStringList(ctx context.Context, attribute string, items []string, reference any) (types.List, diag.Diagnostics) {
	if len(items) == 0 && ListIsNullInReference(ctx, attribute, reference) {
		return types.ListNull(types.StringType), nil
	}

	return types.ListValueFrom(ctx, types.StringType, items)
}

////
//  IDs
////

// ExpandRawID returns the raw UUID (without locality) in string pointer form, from an ID in Terraform types.String form.
func ExpandRawID(str types.String, attributeName string, diags *diag.Diagnostics) *string {
	rawID, err := locality.ExtractUUID(str.ValueString())
	if rawID == "" {
		return nil
	}

	if err != nil {
		diags.AddAttributeError(path.Root(attributeName), "Failed to parse "+attributeName, err.Error())

		return nil
	}

	return new(rawID)
}

// ExpandRawIDSet returns the raw UUIDs (without locality) in string array form, from a Terraform types.Set.
func ExpandRawIDSet(ctx context.Context, set types.Set, attributeName string, diags *diag.Diagnostics) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var result []string
	diags.Append(set.ElementsAs(ctx, &result, false)...)

	rawIDs := make([]string, 0, len(result))
	for _, id := range result {
		rawID, err := locality.ExtractUUID(id)
		if err != nil {
			diags.AddAttributeError(path.Root(attributeName), "Failed to parse "+attributeName, err.Error())

			return nil
		}

		rawIDs = append(rawIDs, rawID)
	}

	return rawIDs
}

// FlattenIDSet builds a Terraform types.Set of IDs in either raw or localized form, from a list of raw UUIDs.
// The reference (config or state) is included to write the IDs in the same form.
func FlattenIDSet(ctx context.Context, attribute string, rawIDs []string, locality string, reference any) (types.Set, diag.Diagnostics) {
	if len(rawIDs) == 0 {
		if SetIsNullInReference(ctx, attribute, reference) {
			return types.SetNull(types.StringType), nil
		}

		return types.SetValueFrom(ctx, types.StringType, rawIDs)
	}

	idsFlat := make([]string, 0, len(rawIDs))
	for _, rawID := range rawIDs {
		localized, diags := IDUsesLocalizedFormatInSet(ctx, reference, attribute, rawID)
		if diags.HasError() {
			return types.SetNull(types.StringType), diags
		}

		if localized {
			idsFlat = append(idsFlat, fmt.Sprintf("%s/%s", locality, rawID))
		} else {
			idsFlat = append(idsFlat, rawID)
		}
	}

	return types.SetValueFrom(ctx, types.StringType, idsFlat)
}

////
// Reference Checking
////

// IDUsesZonedFormat checks if an ID attribute is in its raw or zoned form in the reference (state or config).
func IDUsesZonedFormat(ctx context.Context, reference any, attributePath path.Path, diags *diag.Diagnostics) bool {
	var refValue basetypes.StringValue

	switch req := reference.(type) {
	case resource.CreateRequest:
		d := req.Config.GetAttribute(ctx, attributePath, &refValue)
		diags.Append(d...)
	case resource.ReadRequest:
		d := req.State.GetAttribute(ctx, attributePath, &refValue)
		diags.Append(d...)
	case resource.UpdateRequest:
		d := req.Config.GetAttribute(ctx, attributePath, &refValue)
		diags.Append(d...)
	default:
		return false
	}

	if diags.HasError() {
		return false
	}

	if refValue.IsNull() {
		// If the value in the reference is null, we assume that we are in an Import context (no state/config)
		return true
	}

	return zonal.ExpandID(refValue.ValueString()).Zone != ""
}

// IDUsesRegionalFormat checks if an ID attribute is in its raw or zoned form in the reference (state or config).
func IDUsesRegionalFormat(ctx context.Context, reference any, attributePath path.Path, diags *diag.Diagnostics) bool {
	var refValue basetypes.StringValue

	switch req := reference.(type) {
	case resource.CreateRequest:
		d := req.Config.GetAttribute(ctx, attributePath, &refValue)
		diags.Append(d...)
	case resource.ReadRequest:
		d := req.State.GetAttribute(ctx, attributePath, &refValue)
		diags.Append(d...)
	case resource.UpdateRequest:
		d := req.Config.GetAttribute(ctx, attributePath, &refValue)
		diags.Append(d...)
	default:
		return false
	}

	if diags.HasError() {
		return false
	}

	if refValue.IsNull() {
		// If the value in the reference is null, we assume that we are in an Import context (no state/config)
		return true
	}

	return regional.ExpandID(refValue.ValueString()).Region != ""
}

// SetIsNullInReference checks if a Set attribute is null (return true) or empty (return false).
func SetIsNullInReference(ctx context.Context, attribute string, reference any) bool {
	var (
		diags  diag.Diagnostics
		refSet basetypes.SetValue
	)

	switch req := reference.(type) {
	case resource.CreateRequest:
		diags = req.Config.GetAttribute(ctx, path.Root(attribute), &refSet)
	case resource.ReadRequest:
		diags = req.State.GetAttribute(ctx, path.Root(attribute), &refSet)
	case resource.UpdateRequest:
		diags = req.Config.GetAttribute(ctx, path.Root(attribute), &refSet)
	}

	if diags.HasError() {
		return false
	}

	if refSet.IsNull() {
		return true
	}

	return false
}

// IDUsesLocalizedFormatInSet checks if an ID contained in a Set attribute is in its raw or zoned form in the reference (state or config).
func IDUsesLocalizedFormatInSet(ctx context.Context, reference any, attribute, idToFind string) (bool, diag.Diagnostics) {
	var (
		diags  diag.Diagnostics
		refSet basetypes.SetValue
	)

	switch req := reference.(type) {
	case resource.CreateRequest:
		diags = req.Config.GetAttribute(ctx, path.Root(attribute), &refSet)
	case resource.ReadRequest:
		diags = req.State.GetAttribute(ctx, path.Root(attribute), &refSet)
	case resource.UpdateRequest:
		diags = req.Config.GetAttribute(ctx, path.Root(attribute), &refSet)
	}

	if diags.HasError() {
		return false, diags
	}

	if refSet.IsNull() {
		// If the set in the reference is null, we assume that we are in an Import context (no state/config)
		return true, nil
	}

	var setAsStringArray []string
	diags.Append(refSet.ElementsAs(ctx, &setAsStringArray, false)...)

	for _, id := range setAsStringArray {
		if locality.ExpandID(id) == idToFind {
			loc, _, _ := locality.ParseLocalizedID(id)
			if loc != "" {
				return true, nil
			}

			return false, nil
		}
	}

	return false, diag.Diagnostics{diag.NewErrorDiagnostic(
		fmt.Sprintf("failed to read %q", attribute),
		fmt.Sprintf("API response contains ID %q that cannot be found in the config", idToFind),
	)}
}

// ListIsNullInReference checks if a List attribute is null (return true) or empty (return false).
func ListIsNullInReference(ctx context.Context, attribute string, reference any) bool {
	var (
		diags   diag.Diagnostics
		reflist basetypes.ListValue
	)

	switch req := reference.(type) {
	case resource.CreateRequest:
		diags = req.Config.GetAttribute(ctx, path.Root(attribute), &reflist)
	case resource.ReadRequest:
		diags = req.State.GetAttribute(ctx, path.Root(attribute), &reflist)
	case resource.UpdateRequest:
		diags = req.Config.GetAttribute(ctx, path.Root(attribute), &reflist)
	}

	if diags.HasError() {
		return false
	}

	if reflist.IsNull() {
		return true
	}

	return false
}
