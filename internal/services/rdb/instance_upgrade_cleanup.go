package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

// deleteReplacedRDBInstance deletes a previous blue/green instance once it is ready.
// A 404 means the instance is already gone and is treated as success.
func deleteReplacedRDBInstance(ctx context.Context, api *rdb.API, region scw.Region, instanceID string, timeout time.Duration) error {
	// Wait until the old instance leaves any transient state so DeleteInstance can succeed.
	_, err := waitForRDBInstance(ctx, api, region, instanceID, timeout)
	if httperrors.Is404(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("old instance %s not ready for deletion: %w", instanceID, err)
	}

	_, err = api.DeleteInstance(&rdb.DeleteInstanceRequest{
		Region:     region,
		InstanceID: instanceID,
	}, scw.WithContext(ctx))
	if httperrors.Is404(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to delete old instance %s: %w", instanceID, err)
	}

	// Wait until the deleted instance disappears from the API.
	_, err = waitForRDBInstance(ctx, api, region, instanceID, timeout)
	if err != nil && !httperrors.Is404(err) {
		return fmt.Errorf("error waiting for old instance %s deletion: %w", instanceID, err)
	}

	return nil
}

// resumeReplacedRDBInstanceCleanup deletes a previous blue/green instance if
// replaced_from_instance_id is still set in state (e.g. after an apply timeout).
func resumeReplacedRDBInstanceCleanup(ctx context.Context, d *schema.ResourceData, api *rdb.API, timeout time.Duration) {
	raw, ok := d.GetOk("replaced_from_instance_id")
	if !ok {
		return
	}

	regionalID, ok := raw.(string)
	if !ok || regionalID == "" {
		return
	}

	region, instanceID, err := regional.ParseID(regionalID)
	if err != nil {
		// Keep the attribute so the invalid value stays visible in state and can be fixed.
		tflog.Warn(ctx, fmt.Sprintf("Invalid replaced_from_instance_id %q, skipping cleanup: %v", regionalID, err))

		return
	}

	tflog.Info(ctx, "Resuming cleanup of replaced RDB instance "+regionalID)

	err = deleteReplacedRDBInstance(ctx, api, region, instanceID, timeout)
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Failed to cleanup replaced instance %s: %v", regionalID, err))

		return
	}

	_ = d.Set("replaced_from_instance_id", "")

	tflog.Info(ctx, "Successfully cleaned up replaced RDB instance "+regionalID)
}
