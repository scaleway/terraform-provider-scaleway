package list

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type RegionalModel interface {
	GetRegions() types.List
}

// ExtractRegions determines regions to query
func ExtractRegions(ctx context.Context, model RegionalModel) ([]scw.Region, error) {
	var regionStrings []string
	regionsList := model.GetRegions()
	if regionsList.IsNull() {
		return nil, nil
	}

	diags := regionsList.ElementsAs(ctx, &regionStrings, false)
	if diags.HasError() {
		return nil, fmt.Errorf("converting regions: %s", diags.Errors()[0].Detail())
	}

	var regionsToQuery []scw.Region
	for _, region := range regionStrings {
		if region == "*" {
			return scw.AllRegions, nil
		}

		parsedRegion, err := scw.ParseRegion(region)
		if err != nil {
			return nil, err
		}
		regionsToQuery = append(regionsToQuery, parsedRegion)
	}

	return regionsToQuery, nil
}
