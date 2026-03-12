package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceSnapshot() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasource.SchemaFromResourceSchema(ResourceSnapshot().SchemaFunc())

	// Set 'Optional' schema elements
	datasource.AddOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["snapshot_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the snapshot",
		ConflictsWith:    []string{"name"},
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
	}
	dsSchema["name"].ConflictsWith = []string{"snapshot_id"}

	return &schema.Resource{
		ReadContext: DataSourceInstanceSnapshotRead,
		Schema:      dsSchema,
	}
}

func DataSourceInstanceSnapshotRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	var snapshot *instance.Snapshot

	snapshotID, ok := d.GetOk("snapshot_id")
	if !ok {
		snapshotName := d.Get("name").(string)

		res, err := instanceAPI.ListSnapshots(&instance.ListSnapshotsRequest{
			Zone:    zone,
			Name:    types.ExpandStringPtr(snapshotName),
			Project: types.ExpandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundSnapshot, err := datasource.FindExact(
			res.Snapshots,
			func(s *instance.Snapshot) bool { return s.Name == snapshotName },
			snapshotName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		snapshot = foundSnapshot
		snapshotID = foundSnapshot.ID
	} else {
		id, err := locality.ExtractUUID(snapshotID.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		res, err := instanceAPI.GetSnapshot(&instance.GetSnapshotRequest{
			Zone:       zone,
			SnapshotID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		snapshot = res.Snapshot
	}

	zonedID := datasource.NewZonedID(snapshotID, zone)

	d.SetId(zonedID)

	err = d.Set("snapshot_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setSnapshotState(d, snapshot)
}
