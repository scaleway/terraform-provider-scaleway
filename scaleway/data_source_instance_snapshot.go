package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceSnapshot() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstanceSnapshot().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "zone", "project_id")

	dsSchema["snapshot_id"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Description:   "The ID of the snapshot",
		ConflictsWith: []string{"name"},
		ValidateFunc:  validationUUIDorUUIDWithLocality(),
	}
	dsSchema["name"].ConflictsWith = []string{"snapshot_id"}

	return &schema.Resource{
		ReadContext: dataSourceScalewayInstanceSnapshotRead,
		Schema:      dsSchema,
	}
}

func dataSourceScalewayInstanceSnapshotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshotID, ok := d.GetOk("snapshot_id")
	if !ok {
		snapshotName := d.Get("name").(string)
		res, err := instanceAPI.ListSnapshots(&instance.ListSnapshotsRequest{
			Zone:    zone,
			Name:    expandStringPtr(snapshotName),
			Project: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		foundSnapshot, err := findExact(
			res.Snapshots,
			func(s *instance.Snapshot) bool { return s.Name == snapshotName },
			snapshotName,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		snapshotID = foundSnapshot.ID
	}

	zonedID := datasourceNewZonedID(snapshotID, zone)

	d.SetId(zonedID)

	err = d.Set("snapshot_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	diags := resourceScalewayInstanceSnapshotRead(ctx, d, m)
	if len(diags) > 0 {
		return diags
	}

	if d.Id() == "" {
		return diag.Errorf("instance snapshot (%s) not found", zonedID)
	}

	return nil
}
