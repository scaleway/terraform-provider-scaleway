package acctest

import (
	"context"
	"fmt"

	cockpitSDK "github.com/scaleway/scaleway-sdk-go/api/cockpit/v1"
	rdbSDK "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

// CockpitReadProbe checks that cockpit read permissions are available on projectID.
func CockpitReadProbe(tt *TestTools, projectID string, region scw.Region) func(context.Context) error {
	return func(ctx context.Context) error {
		api := cockpitSDK.NewRegionalAPI(tt.Meta.ScwClient())

		_, err := api.ListDataSources(&cockpitSDK.RegionalAPIListDataSourcesRequest{
			Region:    region,
			ProjectID: projectID,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}

		return nil
	}
}

// RDBReadProjectProbe checks that RDB list permissions are available on projectID.
func RDBReadProjectProbe(tt *TestTools, projectID string, region scw.Region) func(context.Context) error {
	return func(ctx context.Context) error {
		api := rdbSDK.NewAPI(tt.Meta.ScwClient())

		_, err := api.ListInstances(&rdbSDK.ListInstancesRequest{
			Region:    region,
			ProjectID: &projectID,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}

		return nil
	}
}

// RDBReadInstanceProbe checks that RDB read permissions are available on a created instance.
func RDBReadInstanceProbe(tt *TestTools, regionalInstanceID string) func(context.Context) error {
	return func(ctx context.Context) error {
		region, instanceID, err := regional.ParseID(regionalInstanceID)
		if err != nil {
			return fmt.Errorf("parse instance id: %w", err)
		}

		api := rdbSDK.NewAPI(tt.Meta.ScwClient())

		_, err = api.GetInstance(&rdbSDK.GetInstanceRequest{
			Region:     region,
			InstanceID: instanceID,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}

		return nil
	}
}

// PreCheckWaitForCockpitIAM waits for cockpit read permissions before creating cockpit resources.
func PreCheckWaitForCockpitIAM(tt *TestTools, projectID string) func() {
	return func() {
		if projectID == "" {
			tt.T.Fatal("projectID is empty: use StoreResourceID in the previous test step")
		}

		err := WaitForProjectIAM(tt.T.Context(), CockpitReadProbe(tt, projectID, scw.RegionFrPar))
		if err != nil {
			tt.T.Fatalf("wait for cockpit IAM on project %s: %v", projectID, err)
		}
	}
}

// PreCheckWaitForRDBProjectIAM waits for RDB list permissions on a new project.
func PreCheckWaitForRDBProjectIAM(tt *TestTools, projectID string) func() {
	return func() {
		if projectID == "" {
			tt.T.Fatal("projectID is empty: use StoreResourceID in the previous test step")
		}

		err := WaitForProjectIAM(tt.T.Context(), RDBReadProjectProbe(tt, projectID, scw.RegionFrPar))
		if err != nil {
			tt.T.Fatalf("wait for RDB IAM on project %s: %v", projectID, err)
		}
	}
}

// PreCheckWaitForRDBInstanceIAM waits for RDB read permissions on a created instance.
func PreCheckWaitForRDBInstanceIAM(tt *TestTools, regionalInstanceID string) func() {
	return func() {
		if regionalInstanceID == "" {
			tt.T.Fatal("instanceID is empty: use StoreResourceID in the previous test step")
		}

		err := WaitForProjectIAM(tt.T.Context(), RDBReadInstanceProbe(tt, regionalInstanceID))
		if err != nil {
			tt.T.Fatalf("wait for RDB IAM on instance %s: %v", regionalInstanceID, err)
		}
	}
}
