package tem

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	accountSDK "github.com/scaleway/scaleway-sdk-go/api/account/v3"
	tem "github.com/scaleway/scaleway-sdk-go/api/tem/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	DefaultDomainTimeout           = 5 * time.Minute
	defaultDomainValidationTimeout = 60 * time.Minute
	defaultDomainRetryInterval     = 15 * time.Second
)

// temAPIWithRegion returns a new Tem API and the region for a Create request
func temAPIWithRegion(d *schema.ResourceData, m any) (*tem.API, scw.Region, error) {
	api := tem.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// NewAPIWithRegionAndID returns a Tem API with zone and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*tem.API, scw.Region, string, error) {
	api := tem.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, id, nil
}

func getDefaultProjectID(ctx context.Context, m any) (string, error) {
	accountAPI := account.NewProjectAPI(m)

	res, err := accountAPI.ListProjects(&accountSDK.ProjectAPIListProjectsRequest{
		Name: types.ExpandStringPtr("default"),
	}, scw.WithContext(ctx))
	if err != nil {
		return "", err
	}

	return res.Projects[0].ID, nil
}
