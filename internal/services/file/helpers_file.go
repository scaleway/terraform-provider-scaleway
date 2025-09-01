package file

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultFileSystemTimeout       = 5 * time.Minute
	defaultFileSystemRetryInterval = 5 * time.Second
)

func fileSystemAPIWithZone(d *schema.ResourceData, m any) (*file.API, scw.Region, error) {
	fileAPI := file.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return fileAPI, region, nil
}

func NewAPIWithRegionAndID(m any, regionID string) (*file.API, scw.Region, string, error) {
	fileAPI := file.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionID)
	if err != nil {
		return nil, "", "", err
	}

	return fileAPI, region, ID, nil
}
