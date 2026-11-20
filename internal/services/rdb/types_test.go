package rdb_test

import (
	"testing"
	"time"

	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/rdb"
	"github.com/stretchr/testify/require"
)

func TestMergeInstanceSettings(t *testing.T) {
	t.Parallel()

	current := map[string]string{
		"max_connections":      "100",
		"work_mem":             "4",
		"effective_cache_size": "1300",
	}

	t.Run("create overlay preserves defaults", func(t *testing.T) {
		t.Parallel()

		got := rdb.MergeInstanceSettings(current, nil, map[string]string{
			"max_connections": "200",
		})
		require.Equal(t, "200", got["max_connections"])
		require.Equal(t, "4", got["work_mem"])
		require.Equal(t, "1300", got["effective_cache_size"])
	})

	t.Run("removing a previously managed key drops it from the merge", func(t *testing.T) {
		t.Parallel()

		oldManaged := map[string]string{
			"max_connections": "200",
			"work_mem":        "8",
		}
		newManaged := map[string]string{
			"max_connections": "200",
		}
		api := map[string]string{
			"max_connections":      "200",
			"work_mem":             "8",
			"effective_cache_size": "1300",
		}

		got := rdb.MergeInstanceSettings(api, oldManaged, newManaged)
		require.Equal(t, "200", got["max_connections"])
		require.Equal(t, "1300", got["effective_cache_size"])
		_, hasWorkMem := got["work_mem"]
		require.False(t, hasWorkMem)
	})
}

func TestFilterInstanceSettings(t *testing.T) {
	t.Parallel()

	all := map[string]string{
		"max_connections": "200",
		"work_mem":        "4",
	}
	got := rdb.FilterInstanceSettings(all, map[string]bool{"max_connections": true})
	require.Equal(t, map[string]string{"max_connections": "200"}, got)
	require.Empty(t, rdb.FilterInstanceSettings(all, map[string]bool{}))
}

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
