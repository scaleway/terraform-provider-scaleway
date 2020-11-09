package scaleway

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sweepZones(zones []scw.Zone, f func(scwClient *scw.Client, zone scw.Zone) error) error {
	for _, zone := range zones {
		client, err := sharedClientForZone(zone)
		if err != nil {
			return err
		}
		err = f(client, zone)
		if err != nil {
			l.Warningf("error running sweepZones, ignoring: %s", err)
		}
	}
	return nil
}

func sweepRegions(regions []scw.Region, f func(scwClient *scw.Client, region scw.Region) error) error {
	for _, region := range regions {
		return sweepZones(region.GetZones(), func(scwClient *scw.Client, zone scw.Zone) error {
			r, _ := zone.Region()
			return f(scwClient, r)
		})
	}
	return nil
}

// sharedClientForZone returns a Scaleway client needed for the sweeper
// functions for a given zone
func sharedClientForZone(zone scw.Zone) (*scw.Client, error) {
	meta, err := buildMeta(&MetaConfig{
		terraformVersion: "terraform-tests",
		forceZone:        zone,
	})
	if err != nil {
		return nil, err
	}
	return meta.scwClient, nil
}

// sharedS3ClientForRegion returns a common S3 client needed for the sweeper
func sharedS3ClientForRegion(region scw.Region) (*s3.S3, error) {
	meta, err := buildMeta(&MetaConfig{
		terraformVersion: "terraform-tests",
		forceZone:        region.GetZones()[0],
	})
	if err != nil {
		return nil, err
	}
	return newS3ClientFromMeta(meta)
}
