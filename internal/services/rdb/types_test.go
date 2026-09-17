package rdb_test

import (
	"testing"
	"time"

	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb"
	"github.com/stretchr/testify/require"
)

func TestFlattenInstanceMaintenances(t *testing.T) {
	t.Parallel()

	startsAt := time.Date(2026, 5, 26, 4, 0, 0, 0, time.UTC)
	stopsAt := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	forcedAt := time.Date(2026, 5, 27, 4, 0, 0, 0, time.UTC)

	got := rdb.FlattenInstanceMaintenances([]*rdbSDK.Maintenance{
		{
			StartsAt:     &startsAt,
			StopsAt:      &stopsAt,
			Reason:       "Minor version upgrade",
			Status:       rdbSDK.MaintenanceStatusPending,
			ForcedAt:     &forcedAt,
			IsApplicable: true,
		},
	})

	maintenances, ok := got.([]map[string]any)
	require.True(t, ok)
	require.Len(t, maintenances, 1)
	require.Equal(t, startsAt.Format(time.RFC3339), maintenances[0]["starts_at"])
	require.Equal(t, stopsAt.Format(time.RFC3339), maintenances[0]["stops_at"])
	require.Empty(t, maintenances[0]["closed_at"])
	require.Equal(t, "Minor version upgrade", maintenances[0]["reason"])
	require.Equal(t, "pending", maintenances[0]["status"])
	require.Equal(t, forcedAt.Format(time.RFC3339), maintenances[0]["forced_at"])
	require.Equal(t, true, maintenances[0]["is_applicable"])

	empty := rdb.FlattenInstanceMaintenances(nil)
	emptyMaintenances, ok := empty.([]map[string]any)
	require.True(t, ok)
	require.Empty(t, emptyMaintenances)
}
