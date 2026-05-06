package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

type ZonalModel interface {
	GetZones() types.List
}

// ExtractZones determines zones to query.
// If zones is null, returns the default zone from the provider config.
func ExtractZones(ctx context.Context, model ZonalModel, m *meta.Meta) ([]scw.Zone, error) {
	zonesList := model.GetZones()
	if zonesList.IsNull() {
		defaultZone, exists := m.ScwClient().GetDefaultZone()
		if !exists {
			return nil, errors.New("no zones specified and no default zone configured")
		}

		return []scw.Zone{defaultZone}, nil
	}

	var zoneStrings []string

	diags := zonesList.ElementsAs(ctx, &zoneStrings, false)
	if diags.HasError() {
		return nil, fmt.Errorf("converting zones: %s", diags.Errors()[0].Detail())
	}

	var res []scw.Zone

	for _, zone := range zoneStrings {
		if zone == "*" {
			return scw.AllZones, nil
		}

		parsedZone, err := scw.ParseZone(zone)
		if err != nil {
			return nil, err
		}

		res = append(res, parsedZone)
	}

	return res, nil
}

type ZonalFetchTarget struct {
	Zone      scw.Zone
	ProjectID string
}
