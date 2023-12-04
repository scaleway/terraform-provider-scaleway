package scaleway

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	serverless_sqldb "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	defaultSDBSQLTimeout = time.Minute * 15
)

// serverlessSQLdbAPIWithRegion returns a new serverless_sqldb API and the region for a Create request
func serverlessSQLdbAPIWithRegion(d *schema.ResourceData, m interface{}) (*serverless_sqldb.API, scw.Region, error) {
	meta := m.(*Meta)
	sdbAPI := serverless_sqldb.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return sdbAPI, region, nil
}

// serverlessSQLdbAPIWithRegionalAndID returns a new serverless_sqldb API with region and ID extracted from the state
func serverlessSQLdbAPIWithRegionAndID(m interface{}, regionalID string) (*serverless_sqldb.API, scw.Region, string, error) {
	meta := m.(*Meta)
	sdbAPI := serverless_sqldb.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return sdbAPI, region, ID, nil
}

func waitForServerlessSQLDBDatabase(ctx context.Context, serverless_sqldbAPI *serverless_sqldb.API, region scw.Region, id string, timeout time.Duration) (*serverless_sqldb.Database, error) {
	retryInterval := defaultFunctionRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	database, err := serverless_sqldbAPI.WaitForDatabase(&serverless_sqldb.WaitForDatabaseRequest{
		Region:        region,
		DatabaseID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return database, err
}
