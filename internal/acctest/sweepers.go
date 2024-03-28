package acctest

import (
	"context"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/logging"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

func Sweep(f func(scwClient *scw.Client) error) error {
	ctx := context.Background()
	m, err := meta.NewMeta(ctx, &meta.Config{
		TerraformVersion: "terraform-tests",
	})
	if err != nil {
		return err
	}
	return f(m.ScwClient())
}

func SweepZones(zones []scw.Zone, f func(scwClient *scw.Client, zone scw.Zone) error) error {
	for _, zone := range zones {
		client, err := sharedClientForZone(zone)
		if err != nil {
			return err
		}
		err = f(client, zone)
		if err != nil {
			logging.L.Warningf("error running sweepZones, ignoring: %s", err)
		}
	}
	return nil
}

func SweepRegions(regions []scw.Region, f func(scwClient *scw.Client, region scw.Region) error) error {
	zones := make([]scw.Zone, len(regions))
	for i, region := range regions {
		zones[i] = region.GetZones()[0]
	}

	return SweepZones(zones, func(scwClient *scw.Client, zone scw.Zone) error {
		r, _ := zone.Region()
		return f(scwClient, r)
	})
}

// sharedClientForZone returns a Scaleway client needed for the sweeper
// functions for a given zone
func sharedClientForZone(zone scw.Zone) (*scw.Client, error) {
	ctx := context.Background()
	m, err := meta.NewMeta(ctx, &meta.Config{
		TerraformVersion: "terraform-tests",
		ForceZone:        zone,
	})
	if err != nil {
		return nil, err
	}
	return m.ScwClient(), nil
}
