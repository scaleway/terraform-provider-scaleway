package file

import (
	"fmt"
	"time"

	file "github.com/scaleway/scaleway-sdk-go/api/file/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultFileSystemTimeout       = 5 * time.Minute
	defaultFileSystemRetryInterval = 5 * time.Second
)

func NewAPIWithRegionAndID(m any, regionID string) (*file.API, scw.Region, string, error) {
	fileAPI := file.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(regionID)
	if err != nil {
		return nil, "", "", fmt.Errorf("parsing regional id %q: %w", regionID, err)
	}

	return fileAPI, region, ID, nil
}
