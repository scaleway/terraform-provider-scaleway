package block

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	block "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceSnapshot() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceSnapshot().Schema)

	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "volume_id", "project_id")

	dsSchema["snapshot_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the snapshot",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}

	return &schema.Resource{
		ReadContext: DataSourceBlockSnapshotRead,
		Schema:      dsSchema,
		Identity:    blockIdentity(),
	}
}

func DataSourceBlockSnapshotRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := blockAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshotID, snapshotIDExists := d.GetOk("snapshot_id")
	if !snapshotIDExists {
		res, err := api.ListSnapshots(&block.ListSnapshotsRequest{
			Zone:      zone,
			Name:      types.ExpandStringPtr(d.Get("name")),
			ProjectID: types.ExpandStringPtr(d.Get("project_id")),
			VolumeID:  types.ExpandStringPtr(d.Get("volume_id")),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		for _, snapshot := range res.Snapshots {
			if snapshot.Name == d.Get("name").(string) {
				if snapshotID != "" {
					return diag.Errorf("more than 1 snapshot found with the same name %s", d.Get("name"))
				}

				snapshotID = snapshot.ID
			}
		}

		if snapshotID == "" {
			return diag.Errorf("no snapshot found with the name %s", d.Get("name"))
		}
	}

	zoneID := datasource.NewZonedID(snapshotID, zone)
	d.SetId(zoneID)

	err = d.Set("snapshot_id", zoneID)
	if err != nil {
		return diag.FromErr(err)
	}

	diags := ResourceBlockSnapshotRead(ctx, d, m)
	if diags != nil {
		return append(diags, diag.Errorf("failed to read snapshot state")...)
	}

	if d.Id() == "" {
		return diag.Errorf("snapshot (%s) not found", zoneID)
	}

	return nil
}
