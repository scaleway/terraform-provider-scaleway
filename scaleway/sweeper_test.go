package scaleway

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sweepZones(region string, f func(scwClient *scw.Client) error) error {
	scwRegion, err := scw.ParseRegion(region)
	if err != nil {
		return err
	}
	for _, zone := range scwRegion.GetZones() {
		client, err := sharedClientForZone(zone.String())
		if err != nil {
			return err
		}
		err = f(client)
		if err != nil {
			l.Warningf("error running sweepZones, ignoring: %s", err)
		}
	}
	return nil
}

// sharedClientForRegion returns a Scaleway client needed for the sweeper
// functions for a given region {fr-par,nl-ams}
func sharedClientForRegion(region string) (*scw.Client, error) {
	meta, err := buildMeta(&metaConfig{
		terraformVersion: "test",
		forceRegion:      region,
	})
	if err != nil {
		return nil, err
	}

	return meta.scwClient, nil
}

// sharedClientForZone returns a Scaleway client needed for the sweeper
// functions for a given zone {fr-par-1,fr-par-2,nl-ams-1}
func sharedClientForZone(zone string) (*scw.Client, error) {
	meta, err := buildMeta(&metaConfig{
		terraformVersion: "test",
		forceZone:        zone,
	})
	if err != nil {
		return nil, err
	}
	return meta.scwClient, nil
}

// sharedS3ClientForRegion returns a common S3 client needed for the sweeper
func sharedS3ClientForRegion(region string) (*s3.S3, error) {
	meta, err := buildMeta(&metaConfig{
		terraformVersion: "test",
		forceRegion:      region,
	})
	if err != nil {
		return nil, err
	}

	return newS3ClientFromMeta(meta)
}
