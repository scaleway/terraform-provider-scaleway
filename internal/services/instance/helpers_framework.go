package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	instanceV2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	types2 "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

// expandStringList builds an array of strings from a Terraform types.List
func expandStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)

	return result
}

// expandUpdatedStringList builds an array of strings from a Terraform types.List.
// If list is nil, returns a pointer on an empty array to trigger the update of the field in the request.
func expandUpdatedStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	result := make([]string, 0)

	if list.IsNull() || list.IsUnknown() {
		return result
	}

	diags.Append(list.ElementsAs(ctx, &result, false)...)

	return result
}

func flattenStringList(ctx context.Context, items []string) (types.List, diag.Diagnostics) {
	if len(items) == 0 {
		return types.ListNull(types.StringType), nil
	}

	return types.ListValueFrom(ctx, types.StringType, items)
}

func flattenLocalizedIDList(ctx context.Context, ids []string, locality string) (types.List, diag.Diagnostics) {
	if len(ids) == 0 {
		return types.ListNull(types.StringType), nil
	}

	localizedIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		localizedIDs = append(localizedIDs, locality+"/"+id)
	}

	return types.ListValueFrom(ctx, types.StringType, localizedIDs)
}

// expandRawIDList takes a Terraform types.String and returns a raw UUID in string pointer from (without locality).  // TODO : comment does not match function
// If the parameter string is empty, it returns nil.
func expandRawID(str types.String, attributeName string, diags *diag.Diagnostics) *string {
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

// expandRawIDList builds an array of raw UUIDs in string from (without locality) from a Terraform types.List
func expandRawIDList(ctx context.Context, list types.List, attributeName string, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)

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

type privateNetworkSpecs struct {
	privateNetworkIDs []string
}

func expandPrivateNetworks(ctx context.Context, networks types.List, d *diag.Diagnostics) privateNetworkSpecs {
	// specs := privateNetworkSpecs{}

	rawIDsList := expandRawIDList(ctx, networks, "private_networks", d)
	pnIDs := make([]string, 0, len(rawIDsList))
	for _, rawID := range rawIDsList {
		pnIDs = append(pnIDs, rawID)
	}

	return privateNetworkSpecs{
		privateNetworkIDs: rawIDsList,
	}
}

func (p privateNetworkSpecs) ToCreateRequest() []*instanceV2.CreateTemplateRequestPrivateNetworkTemplate {
	var createReq []*instanceV2.CreateTemplateRequestPrivateNetworkTemplate

	for _, pnID := range p.privateNetworkIDs {
		createReq = append(createReq, &instanceV2.CreateTemplateRequestPrivateNetworkTemplate{PrivateNetworkID: pnID})
	}

	return createReq
}

func (p privateNetworkSpecs) ToUpdateRequest() *instanceV2.UpdateTemplateRequestUpdatePrivateNetworks {
	return &instanceV2.UpdateTemplateRequestUpdatePrivateNetworks{PrivateNetworks: p.ToCreateRequest()}
}

func flattenPrivateNetworks(ctx context.Context, pns []*instanceV2.CreateTemplateRequestPrivateNetworkTemplate, zone scw.Zone) (types.List, diag.Diagnostics) {
	if len(pns) == 0 {
		return types.ListNull(types.StringType), nil
	}

	region, err := zone.Region()
	if err != nil {
		return types.ListNull(types.StringType), diag.Diagnostics{diag.NewErrorDiagnostic(fmt.Sprintf("failed to infer region from zone %q", zone), err.Error())}
	}

	regionalPNIDs := make([]string, 0, len(pns))
	for _, pn := range pns {
		regionalPNIDs = append(regionalPNIDs, fmt.Sprintf("%s/%s", region, pn.PrivateNetworkID))
	}

	return types.ListValueFrom(ctx, types.StringType, regionalPNIDs)
}

type volumeSpecs struct {
	VolumeType     instanceV2.CreateServerRequestServerVolumeVolumeType `tfsdk:"volume_type"`
	Name           string                                               `tfsdk:"name"`
	Tags           []string                                             `tfsdk:"tags"`
	Size           *scw.Size                                            `tfsdk:"size"`
	BaseSnapshotID *string                                              `tfsdk:"base_snapshot_id"`
	ImageLabel     *string                                              `tfsdk:"image_label"`
	PerfIops       *uint32                                              `tfsdk:"perf_iops"`
}

//type volumeSpecs struct {
//	VolumeType instanceV2.CreateServerRequestServerVolumeVolumeType
//	Name string
//	Tags []string
//	Size *scw.Size
//	BaseSnapshotID *string
//	ImageLabel *string
//	PerfIops *uint32
//}

type volumesSpecs struct {
	volumeSpecs []volumeSpecs
}

func expandVolumes(ctx context.Context, volumes types.List, diags *diag.Diagnostics) volumesSpecs {
	specs := volumesSpecs{}

	if volumes.IsNull() || volumes.IsUnknown() {
		return specs
	}

	var result []volumeSpecs
	diags.Append(volumes.ElementsAs(ctx, &result, false)...)

	specs.volumeSpecs = result

	return specs
}

func (p volumesSpecs) ToCreateRequest() []*instanceV2.CreateTemplateRequestVolumeTemplate {
	var createReq []*instanceV2.CreateTemplateRequestVolumeTemplate

	for _, vol := range p.volumeSpecs {
		createReq = append(createReq, &instanceV2.CreateTemplateRequestVolumeTemplate{
			VolumeType:     vol.VolumeType,
			Name:           types2.ExpandOrGenerateString(vol.Name, "tf-tmpl-vol"),
			Tags:           vol.Tags,
			Size:           vol.Size,
			BaseSnapshotID: vol.BaseSnapshotID,
			ImageLabel:     vol.ImageLabel,
			PerfIops:       vol.PerfIops,
		})
	}

	return createReq
}

func (p volumesSpecs) ToUpdateRequest() *instanceV2.UpdateTemplateRequestUpdateVolumes {
	return &instanceV2.UpdateTemplateRequestUpdateVolumes{Volumes: p.ToCreateRequest()}
}

func flattenVolumes(ctx context.Context, volumes []*instanceV2.CreateTemplateRequestVolumeTemplate) (types.List, diag.Diagnostics) {
	if len(volumes) == 0 {
		return types.ListNull(types.ObjectType{}), nil
	}

	//region, err := zone.Region()
	//if err != nil {
	//	return types.ListNull(types.ObjectType{}), diag.Diagnostics{diag.NewErrorDiagnostic(fmt.Sprintf("failed to infer region from zone %q", zone), err.Error())}
	//}

	volumesFlat := make([]volumeSpecs, 0, len(volumes))
	for _, vol := range volumes {
		volumesFlat = append(volumesFlat, volumeSpecs{
			VolumeType:     vol.VolumeType,
			Name:           vol.Name,
			Tags:           vol.Tags,
			Size:           vol.Size,
			BaseSnapshotID: vol.BaseSnapshotID,
			ImageLabel:     vol.ImageLabel,
			PerfIops:       vol.PerfIops,
		})
	}

	return types.ListValueFrom(ctx, types.ObjectType{}, volumesFlat)
}

func idUsesZonedFormat(ctx context.Context, reference any, attribute string, diags *diag.Diagnostics) bool {
	switch req := reference.(type) {
	case resource.CreateRequest:
		var configValue basetypes.StringValue
		d := req.Config.GetAttribute(ctx, path.Root(attribute), &configValue)
		diags.Append(d...)

		if d.HasError() {
			return false
		}

		if configValue.IsNull() {
			return true
		}

		return zonal.ExpandID(configValue.ValueString()).Zone != ""
	case resource.ReadRequest:
		var stateValue basetypes.StringValue
		d := req.State.GetAttribute(ctx, path.Root(attribute), &stateValue)
		diags.Append(d...)

		if d.HasError() {
			return false
		}

		if stateValue.IsNull() {
			return true
		}

		return zonal.ExpandID(stateValue.ValueString()).Zone != ""
	case resource.UpdateRequest:
		var configValue basetypes.StringValue
		d := req.Config.GetAttribute(ctx, path.Root(attribute), &configValue)
		diags.Append(d...)

		if d.HasError() {
			return false
		}

		if configValue.IsNull() {
			return true
		}

		return zonal.ExpandID(configValue.ValueString()).Zone != ""
	default:
		return false
	}
}
