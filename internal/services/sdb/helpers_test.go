package sdb_test

import (
	"sort"

	sdbSDK "github.com/scaleway/scaleway-sdk-go/api/serverless_sqldb/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func fetchAvailableVersions(tt *acctest.TestTools) []string {
	tt.T.Helper()

	api := sdbSDK.NewAPI(tt.Meta.ScwClient())

	versionsResp, err := api.ListVersions(&sdbSDK.ListVersionsRequest{
		Region: scw.RegionFrPar,
	}, scw.WithContext(tt.T.Context()), scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch sdb versions: %s", err)
	}

	if len(versionsResp.Versions) == 0 {
		tt.T.Fatal("no sdb versions available")
	}

	names := make([]string, 0, len(versionsResp.Versions))
	for _, version := range versionsResp.Versions {
		names = append(names, version.Name)
	}

	sort.Strings(names)

	return names
}
