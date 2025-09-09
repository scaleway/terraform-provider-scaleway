package vpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	validator "github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const defaultVPCPrivateNetworkRetryInterval = 30 * time.Second

// vpcAPIWithRegion returns a new VPC API and the region for a Create request
func vpcAPIWithRegion(d *schema.ResourceData, m any) (*vpc.API, scw.Region, error) {
	vpcAPI := vpc.NewAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return vpcAPI, region, err
}

// NewAPIWithRegionAndID returns a new VPC API with locality and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*vpc.API, scw.Region, string, error) {
	vpcAPI := vpc.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return vpcAPI, region, ID, err
}

func NewAPIWithRegionAndIDFromState(m interface{}, d *schema.ResourceData) (
	*vpc.API,
	scw.Region,
	string,
	error) {
	vpcAPI := vpc.NewAPI(meta.ExtractScwClient(m))

	region, ID, err := regional.ResolveRegionAndID(d, func(d *schema.ResourceData) (scw.Region, error) {
		_, provRegion, err := vpcAPIWithRegion(d, m)

		return provRegion, err
	})
	if err != nil {
		return nil, "", "", err
	}

	return vpcAPI, region, ID, nil
}

func NewAPI(m any) (*vpc.API, error) {
	return vpc.NewAPI(meta.ExtractScwClient(m)), nil
}

// routesAPIWithRegion returns a new VPC API and the region for a Create request
func routesAPIWithRegion(d *schema.ResourceData, m any) (*vpc.RoutesWithNexthopAPI, scw.Region, error) {
	routesAPI := vpc.NewRoutesWithNexthopAPI(meta.ExtractScwClient(m))

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return routesAPI, region, err
}

func vpcPrivateNetworkUpgradeV1SchemaType() cty.Type {
	return cty.Object(map[string]cty.Type{
		"id": cty.String,
	})
}

func vpcPrivateNetworkV1SUpgradeFunc(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	var err error

	ID, exist := rawState["id"]
	if !exist {
		return nil, errors.New("upgrade: id not exist")
	}

	rawState["id"], err = vpcPrivateNetworkUpgradeV1ZonalToRegionalID(ID.(string))
	if err != nil {
		return nil, err
	}

	return rawState, nil
}

func vpcPrivateNetworkUpgradeV1ZonalToRegionalID(element string) (string, error) {
	l, id, err := locality.ParseLocalizedID(element)
	// return error if l cannot be parsed
	if err != nil {
		return "", fmt.Errorf("upgrade: could not retrieve the locality from `%s`", element)
	}
	// if locality is already regional return
	if validator.IsRegion(l) {
		return element, nil
	}

	fetchRegion, err := scw.Zone(l).Region()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", fetchRegion.String(), id), nil
}

func vpcRouteExpandResourceID(id string) (string, error) {
	parts := strings.Split(id, "/")
	partCount := len(parts)

	switch partCount {
	case 1:
		return id, nil
	case 2:
		_, ID, err := locality.ParseLocalizedID(id)
		if err != nil {
			return "", fmt.Errorf("failed to parse localized ID: %w", err)
		}

		return ID, nil
	case 3:
		// Parse as a nested ID and return the outerID
		_, _, ID, err := locality.ParseLocalizedNestedID(id)
		if err != nil {
			return "", fmt.Errorf("failed to parse nested ID: %w", err)
		}

		return ID, nil
	default:
		return "", fmt.Errorf("unrecognized ID format: %s", id)
	}
}

func diffSuppressFuncRouteResourceID(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	oldResourceID, err := vpcRouteExpandResourceID(oldValue)
	if err != nil {
		return false
	}

	newResourceID, err := vpcRouteExpandResourceID(newValue)
	if err != nil {
		return false
	}

	return oldResourceID == newResourceID
}
