package sdb

import (
	"context"
	"time"

	sdbSDK "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/function"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultTimeout = 15 * time.Minute
)

// NewAPIWithRegionAndID returns a new serverless_sqldb API with region and ID extracted from the state
func NewAPIWithRegionAndID(m any, regionalID string) (*sdbSDK.API, scw.Region, string, error) {
	sdbAPI := sdbSDK.NewAPI(meta.ExtractScwClient(m))

	region, id, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return sdbAPI, region, id, nil
}

func waitForDatabase(ctx context.Context, sdbAPI *sdbSDK.API, region scw.Region, id string, timeout time.Duration) (*sdbSDK.Database, error) {
	retryInterval := function.DefaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	database, err := sdbAPI.WaitForDatabase(&sdbSDK.WaitForDatabaseRequest{
		Region:        region,
		DatabaseID:    id,
		RetryInterval: &retryInterval,
		Timeout:       &timeout,
	}, scw.WithContext(ctx))

	return database, err
}
