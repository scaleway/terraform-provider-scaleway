package functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// diagsFuncError converts diags into a single function error, if any.
func diagsFuncError(diags diag.Diagnostics) *function.FuncError {
	if !diags.HasError() {
		return nil
	}

	texts := make([]string, 0, diags.ErrorsCount())

	for _, d := range diags.Errors() {
		texts = append(texts, d.Summary())
	}

	return function.NewFuncError(strings.Join(texts, "; "))
}

var _ function.Function = &AnnotationsBindings{}

type AnnotationsBindings struct{}

func NewAnnotationsBindings() function.Function {
	return &AnnotationsBindings{}
}

var (
	annotationsBindingAttrTypes = map[string]attr.Type{
		"value_id":   types.StringType,
		"target_srn": types.StringType,
	}

	annotationsBindingType = basetypes.ObjectType{AttrTypes: annotationsBindingAttrTypes}
)

func (f *AnnotationsBindings) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "annotations_bindings"
}

func (f *AnnotationsBindings) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Compute the cross-product of annotation values and targets",
		Description: "Given a map of annotation values (key -> value_id) and a map of targets (key -> SRN), returns the N x M map of scaleway_annotations_binding entries, keyed by \"<value key>|<target key>\", ready to be used as the for_each of a scaleway_annotations_binding resource.",
		Parameters: []function.Parameter{
			function.MapParameter{
				Name:        "values",
				Description: "annotation values: maps a stable key to a scaleway_annotations_value value_id",
				ElementType: types.StringType,
			},
			function.MapParameter{
				Name:        "targets",
				Description: "annotation targets: maps a stable key to the SRN of the resource to annotate",
				ElementType: types.StringType,
			},
		},
		Return: function.MapReturn{
			ElementType: annotationsBindingType,
		},
	}
}

func (f *AnnotationsBindings) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var (
		values  types.Map
		targets types.Map
	)

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &values, &targets))

	if resp.Error != nil {
		return
	}

	if values.IsNull() || targets.IsNull() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.MapNull(annotationsBindingType)))

		return
	}

	if values.IsUnknown() || targets.IsUnknown() {
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.MapUnknown(annotationsBindingType)))

		return
	}

	entries := make(map[string]attr.Value, len(values.Elements())*len(targets.Elements()))

	for valueKey, valueID := range values.Elements() {
		for targetKey, targetSRN := range targets.Elements() {
			key := valueKey + "|" + targetKey

			if _, exists := entries[key]; exists {
				resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(fmt.Sprintf("conflicting binding key %q: distinct (value, target) pairs map to the same key", key)))
				resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, types.MapUnknown(annotationsBindingType)))

				return
			}

			entry, diags := basetypes.NewObjectValue(
				annotationsBindingAttrTypes,
				map[string]attr.Value{
					"value_id":   valueID,
					"target_srn": targetSRN,
				},
			)

			if fe := diagsFuncError(diags); fe != nil {
				resp.Error = function.ConcatFuncErrors(resp.Error, fe, resp.Result.Set(ctx, types.MapUnknown(annotationsBindingType)))

				return
			}

			entries[key] = entry
		}
	}

	result, diags := basetypes.NewMapValue(annotationsBindingType, entries)

	if fe := diagsFuncError(diags); fe != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, fe, resp.Result.Set(ctx, types.MapUnknown(annotationsBindingType)))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, result))
}
