package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceSnapshot() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceScalewayInstanceSnapshot().Schema)

	// Set 'Optional' schema elements
	addOptionalFieldsToSchema(dsSchema, "name", "zone")

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

func dataSourceScalewayInstanceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshotID, ok := d.GetOk("snapshot_id")
	if !ok {
		res, err := instanceAPI.ListSnapshots(&instance.ListSnapshotsRequest{
			Zone:    zone,
			Name:    expandStringPtr(d.Get("name")),
			Project: expandStringPtr(d.Get("project_id")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
		for _, snapshot := range res.Snapshots {
			if snapshot.Name == d.Get("name").(string) {
				if snapshotID != "" {
					return diag.FromErr(fmt.Errorf("more than 1 snapshot found with the same name %s", d.Get("name")))
				}
				snapshotID = snapshot.ID
			}
		}
		if snapshotID == "" {
			return diag.FromErr(fmt.Errorf("no snapshot found with the name %s", d.Get("name")))
		}
	}

	zonedID := datasourceNewZonedID(snapshotID, zone)

	d.SetId(zonedID)

	err = d.Set("snapshot_id", zonedID)
	if err != nil {
		return diag.FromErr(err)
	}
	diags := resourceScalewayInstanceSnapshotRead(ctx, d, meta)
	if len(diags) > 0 {
		return diags
	}

	if d.Id() == "" {
		return diag.Errorf("instance snapshot (%s) not found", zonedID)
	}

	return nil
}
