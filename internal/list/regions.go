package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

type RegionalModel interface {
	GetRegions() types.List
}

// ExtractRegions determines regions to query.
// If regions is null, returns the default region from the provider config.
func ExtractRegions(ctx context.Context, model RegionalModel, meta *meta.Meta) ([]scw.Region, error) {
	regionsList := model.GetRegions()
	if regionsList.IsNull() {
		defaultRegion, exists := meta.ScwClient().GetDefaultRegion()
		if !exists {
			return nil, errors.New("no regions specified and no default region configured")
		}

		return []scw.Region{defaultRegion}, nil
	}

	var regionStrings []string

	diags := regionsList.ElementsAs(ctx, &regionStrings, false)
	if diags.HasError() {
		return nil, fmt.Errorf("converting regions: %s", diags.Errors()[0].Detail())
	}

	var res []scw.Region

	for _, region := range regionStrings {
		if region == "*" {
			return scw.AllRegions, nil
		}

		parsedRegion, err := scw.ParseRegion(region)
		if err != nil {
			return nil, err
		}

		res = append(res, parsedRegion)
	}

	return res, nil
}

type RegionalFetchTarget struct {
	Region    scw.Region
	ProjectID string
}
