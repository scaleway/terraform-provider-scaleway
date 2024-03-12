package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	serverless_sqldb "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
)

const (
	defaultSDBSQLTimeout = time.Minute * 15
)

// serverlessSQLdbAPIWithRegion returns a new serverless_sqldb API and the region for a Create request
func serverlessSQLdbAPIWithRegion(d *schema.ResourceData, m interface{}) (*serverless_sqldb.API, scw.Region, error) {
	sdbAPI := serverless_sqldb.NewAPI(m.(*meta.Meta).ScwClient())

	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return sdbAPI, region, nil
}

// serverlessSQLdbAPIWithRegionalAndID returns a new serverless_sqldb API with region and ID extracted from the state
func serverlessSQLdbAPIWithRegionAndID(m interface{}, regionalID string) (*serverless_sqldb.API, scw.Region, string, error) {
	sdbAPI := serverless_sqldb.NewAPI(m.(*meta.Meta).ScwClient())

	region, ID, err := regional.ParseID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return sdbAPI, region, ID, nil
}

func waitForServerlessSQLDBDatabase(ctx context.Context, sdbAPI *serverless_sqldb.API, region scw.Region, id string, timeout time.Duration) (*serverless_sqldb.Database, error) {
	retryInterval := defaultFunctionRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	database, err := sdbAPI.WaitForDatabase(&serverless_sqldb.WaitForDatabaseRequest{
		Region:        region,
		DatabaseID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return database, err
}
