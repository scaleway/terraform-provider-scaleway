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

func TestFlattenHighAvailabilityMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mode     rdbSDK.HighAvailabilityMode
		isHA     bool
		wantMode string
		wantIsHA bool
	}{
		{
			name:     "single_zone",
			mode:     rdbSDK.HighAvailabilityModeSingleZone,
			isHA:     false,
			wantMode: "single_zone",
			wantIsHA: true,
		},
		{
			name:     "multiple_zone",
			mode:     rdbSDK.HighAvailabilityModeMultipleZone,
			isHA:     false,
			wantMode: "multiple_zone",
			wantIsHA: true,
		},
		{
			name:     "disabled",
			mode:     rdbSDK.HighAvailabilityModeDisabled,
			isHA:     true,
			wantMode: "disabled",
			wantIsHA: false,
		},
		{
			name:     "unknown falls back to is_ha true",
			mode:     rdbSDK.HighAvailabilityModeUnknownHighAvailabilityMode,
			isHA:     true,
			wantMode: "single_zone",
			wantIsHA: true,
		},
		{
			name:     "unknown falls back to is_ha false",
			mode:     rdbSDK.HighAvailabilityModeUnknownHighAvailabilityMode,
			isHA:     false,
			wantMode: "disabled",
			wantIsHA: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMode, gotIsHA := rdb.FlattenHighAvailabilityMode(tt.mode, tt.isHA)
			require.Equal(t, tt.wantMode, gotMode)
			require.Equal(t, tt.wantIsHA, gotIsHA)
		})
	}
}

func TestInstanceHasHA(t *testing.T) {
	t.Parallel()

	require.False(t, rdb.InstanceHasHA(nil))
	require.True(t, rdb.InstanceHasHA(&rdbSDK.Instance{
		HighAvailabilityMode: rdbSDK.HighAvailabilityModeSingleZone,
	}))
	require.False(t, rdb.InstanceHasHA(&rdbSDK.Instance{
		HighAvailabilityMode: rdbSDK.HighAvailabilityModeDisabled,
	}))
	require.True(t, rdb.InstanceHasHA(&rdbSDK.Instance{
		IsHaCluster: true,
	}))
}
