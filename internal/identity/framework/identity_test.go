package framework_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity/framework"
	"github.com/stretchr/testify/assert"
)

func TestResolveZonalIdentity(t *testing.T) {
	t.Parallel()

	const uuid = "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name     string
		identity framework.ZonalIdentity
		wantZone scw.Zone
		wantID   string
	}{
		{
			name:     "legacy zoned id zone/uuid",
			identity: framework.ZonalIdentity{ID: types.StringValue("fr-par-1/" + uuid), Zone: types.StringValue("fr-par-1")},
			wantZone: scw.ZoneFrPar1,
			wantID:   uuid,
		},
		{
			name:     "bare uuid uses identity zone",
			identity: framework.ZonalIdentity{ID: types.StringValue(uuid), Zone: types.StringValue("nl-ams-1")},
			wantZone: scw.ZoneNlAms1,
			wantID:   uuid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			zone, id := framework.ResolveZonalIdentity(tt.identity)
			assert.Equal(t, tt.wantZone, zone)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestResolveRegionalIdentity(t *testing.T) {
	t.Parallel()

	const uuid = "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name       string
		identity   framework.RegionalIdentity
		wantRegion scw.Region
		wantID     string
	}{
		{
			name:       "legacy regional id region/uuid",
			identity:   framework.RegionalIdentity{ID: types.StringValue("fr-par/" + uuid), Region: types.StringValue("fr-par")},
			wantRegion: scw.RegionFrPar,
			wantID:     uuid,
		},
		{
			name:       "bare uuid uses identity region",
			identity:   framework.RegionalIdentity{ID: types.StringValue(uuid), Region: types.StringValue("nl-ams")},
			wantRegion: scw.RegionNlAms,
			wantID:     uuid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			region, id := framework.ResolveRegionalIdentity(tt.identity)
			assert.Equal(t, tt.wantRegion, region)
			assert.Equal(t, tt.wantID, id)
		})
	}
}
