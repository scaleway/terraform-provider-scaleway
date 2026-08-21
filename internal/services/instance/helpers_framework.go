package instance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	instanceV2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	scwtypes "github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

type privateNetworkSpecs struct {
	privateNetworkIDs []string
}

func expandPrivateNetworks(ctx context.Context, networks types.Set, d *diag.Diagnostics) privateNetworkSpecs {
	rawIDsList := scwtypes.ExpandRawIDSet(ctx, networks, "private_networks", d)

	pnIDs := make([]string, 0, len(rawIDsList))
	pnIDs = append(pnIDs, rawIDsList...)

	return privateNetworkSpecs{
		privateNetworkIDs: pnIDs,
	}
}

func (p privateNetworkSpecs) ToCreateRequest() []*instanceV2.CreateTemplateRequestPrivateNetworkTemplate {
	createReq := make([]*instanceV2.CreateTemplateRequestPrivateNetworkTemplate, 0, len(p.privateNetworkIDs))

	for _, pnID := range p.privateNetworkIDs {
		createReq = append(createReq, &instanceV2.CreateTemplateRequestPrivateNetworkTemplate{PrivateNetworkID: pnID})
	}

	return createReq
}

func (p privateNetworkSpecs) ToUpdateRequest() *instanceV2.UpdateTemplateRequestUpdatePrivateNetworks {
	return &instanceV2.UpdateTemplateRequestUpdatePrivateNetworks{PrivateNetworks: p.ToCreateRequest()}
}

func flattenPrivateNetworks(ctx context.Context, pns []*instanceV2.CreateTemplateRequestPrivateNetworkTemplate, zone scw.Zone, reference any) (types.Set, diag.Diagnostics) {
	if len(pns) == 0 {
		if scwtypes.SetIsNullInReference(ctx, "private_networks", reference) {
			return types.SetNull(types.StringType), nil
		}

		return types.SetValueFrom(ctx, types.StringType, pns)
	}

	region, err := zone.Region()
	if err != nil {
		return types.SetNull(types.StringType), diag.Diagnostics{diag.NewErrorDiagnostic(fmt.Sprintf("failed to infer region from zone %q", zone), err.Error())}
	}

	pnIDs := make([]string, 0, len(pns))
	for _, pn := range pns {
		localized, diags := scwtypes.IDUsesLocalizedFormatInSet(ctx, reference, "private_networks", pn.PrivateNetworkID)
		if diags.HasError() {
			return types.SetNull(types.StringType), diags
		}

		if localized {
			pnIDs = append(pnIDs, fmt.Sprintf("%s/%s", region, pn.PrivateNetworkID))
		} else {
			pnIDs = append(pnIDs, pn.PrivateNetworkID)
		}
	}

	return types.SetValueFrom(ctx, types.StringType, pnIDs)
}

func volumeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"volume_type":      types.StringType,
			"name":             types.StringType,
			"tags":             types.ListType{ElemType: types.StringType},
			"size_in_gb":       types.Int64Type,
			"base_snapshot_id": types.StringType,
			"image_label":      types.StringType,
			"perf_iops":        types.Int32Type,
		},
	}
}

type volumeSpecs struct {
	Size           *scw.Size                                            `tfsdk:"size_in_gb"`
	BaseSnapshotID *string                                              `tfsdk:"base_snapshot_id"`
	ImageLabel     *string                                              `tfsdk:"image_label"`
	PerfIops       *uint32                                              `tfsdk:"perf_iops"`
	VolumeType     instanceV2.CreateServerRequestServerVolumeVolumeType `tfsdk:"volume_type"`
	Name           string                                               `tfsdk:"name"`
	Tags           []string                                             `tfsdk:"tags"`
}

type templateVolumesSpecs struct {
	volumeSpecs []volumeSpecs
}

func expandVolumes(ctx context.Context, volumes types.List, diags *diag.Diagnostics) templateVolumesSpecs {
	specs := templateVolumesSpecs{}

	if volumes.IsNull() || volumes.IsUnknown() {
		return specs
	}

	var volumeObjects []types.Object
	diags.Append(volumes.ElementsAs(ctx, &volumeObjects, false)...)

	if diags.HasError() {
		return specs
	}

	for _, volObj := range volumeObjects {
		vol := volumeSpecs{}

		volumeType := volObj.Attributes()["volume_type"].(types.String)
		if !volumeType.IsNull() && !volumeType.IsUnknown() {
			vol.VolumeType = instanceV2.CreateServerRequestServerVolumeVolumeType(volumeType.ValueString())
		}

		name := volObj.Attributes()["name"].(types.String)
		if !name.IsNull() && !name.IsUnknown() {
			vol.Name = name.ValueString()
		}

		tags := volObj.Attributes()["tags"].(types.List)
		if !tags.IsNull() && !tags.IsUnknown() {
			vol.Tags = scwtypes.ExpandStringList(ctx, tags, diags)
		}

		size := volObj.Attributes()["size_in_gb"].(types.Int64)
		if !size.IsNull() && !size.IsUnknown() {
			sizeVal := scw.Size(size.ValueInt64()) * scw.GB
			vol.Size = &sizeVal
		}

		baseSnapshotID := volObj.Attributes()["base_snapshot_id"].(types.String)
		if !baseSnapshotID.IsNull() && !baseSnapshotID.IsUnknown() {
			vol.BaseSnapshotID = scwtypes.ExpandRawID(baseSnapshotID, "base_snapshot_id", diags)
		}

		imageLabel := volObj.Attributes()["image_label"].(types.String)
		if !imageLabel.IsNull() && !imageLabel.IsUnknown() {
			vol.ImageLabel = new(imageLabel.ValueString())
		}

		perfIops := volObj.Attributes()["perf_iops"].(types.Int32)
		if !perfIops.IsNull() && !perfIops.IsUnknown() {
			perfIopsVal := uint32(perfIops.ValueInt32())
			vol.PerfIops = &perfIopsVal
		}

		specs.volumeSpecs = append(specs.volumeSpecs, vol)
	}

	return specs
}

func (p templateVolumesSpecs) ToCreateRequest() []*instanceV2.CreateTemplateRequestVolumeTemplate {
	createReq := make([]*instanceV2.CreateTemplateRequestVolumeTemplate, 0, len(p.volumeSpecs))

	for _, vol := range p.volumeSpecs {
		createReq = append(createReq, &instanceV2.CreateTemplateRequestVolumeTemplate{
			VolumeType:     vol.VolumeType,
			Name:           scwtypes.ExpandOrGenerateString(vol.Name, "tmpl-vol"),
			Tags:           vol.Tags,
			Size:           vol.Size,
			BaseSnapshotID: vol.BaseSnapshotID,
			ImageLabel:     vol.ImageLabel,
			PerfIops:       vol.PerfIops,
		})
	}

	return createReq
}

func (p templateVolumesSpecs) ToUpdateRequest() *instanceV2.UpdateTemplateRequestUpdateVolumes {
	return &instanceV2.UpdateTemplateRequestUpdateVolumes{Volumes: p.ToCreateRequest()}
}

func flattenVolumes(ctx context.Context, volumes []*instanceV2.CreateTemplateRequestVolumeTemplate, reference any, zone scw.Zone) (types.List, diag.Diagnostics) {
	if len(volumes) == 0 {
		return types.ListNull(volumeObjectType()), nil
	}

	volumeObjects := make([]attr.Value, 0, len(volumes))
	for index, vol := range volumes {
		volAttrs := map[string]attr.Value{
			"volume_type":      types.StringValue(string(vol.VolumeType)),
			"name":             types.StringValue(vol.Name),
			"size_in_gb":       types.Int64Null(),
			"base_snapshot_id": types.StringNull(),
			"image_label":      types.StringNull(),
			"perf_iops":        types.Int32Null(),
			"tags":             types.ListNull(types.StringType),
		}

		if len(vol.Tags) > 0 {
			tagsList, diags := types.ListValueFrom(ctx, types.StringType, vol.Tags)
			if diags.HasError() {
				return types.ListNull(volumeObjectType()), diags
			}

			volAttrs["tags"] = tagsList
		}

		if vol.Size != nil {
			volAttrs["size_in_gb"] = types.Int64Value(int64(*vol.Size / scw.GB))
		}

		if vol.BaseSnapshotID != nil {
			var diags diag.Diagnostics

			if scwtypes.IDUsesZonedFormat(ctx, reference, path.Root("volumes").AtListIndex(index).AtName("base_snapshot_id"), &diags) {
				volAttrs["base_snapshot_id"] = types.StringValue(zonal.NewIDString(zone, *vol.BaseSnapshotID))
			} else {
				volAttrs["base_snapshot_id"] = types.StringValue(*vol.BaseSnapshotID)
			}
		}

		if vol.ImageLabel != nil {
			volAttrs["image_label"] = types.StringValue(*vol.ImageLabel)
		}

		if vol.PerfIops != nil {
			volAttrs["perf_iops"] = types.Int32Value(int32(*vol.PerfIops))
		}

		volObj, diags := types.ObjectValue(volumeObjectType().AttrTypes, volAttrs)
		if diags.HasError() {
			return types.ListNull(volumeObjectType()), diags
		}

		volumeObjects = append(volumeObjects, volObj)
	}

	return types.ListValueFrom(ctx, volumeObjectType(), volumeObjects)
}
