package kafka_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	kafkaSDK "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/kafka"
)

// fetchLatestKafkaVersion returns the latest available Kafka version for testing purposes
func fetchLatestKafkaVersion(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := kafka.NewAPI(tt.Meta)

	versionsResp, err := api.ListVersions(&kafkaSDK.ListVersionsRequest{}, scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch kafka versions: %s", err)
	}

	if len(versionsResp.Versions) == 0 {
		tt.T.Fatal("no kafka versions available")
	}

	return versionsResp.Versions[0].Version
}

// fetchAvailableKafkaNodeType returns an available node type for testing purposes
func fetchAvailableKafkaNodeType(tt *acctest.TestTools) string {
	tt.T.Helper()

	api := kafka.NewAPI(tt.Meta)

	nodeTypesResp, err := api.ListNodeTypes(&kafkaSDK.ListNodeTypesRequest{
		Region: scw.RegionFrPar,
	}, scw.WithAllPages())
	if err != nil {
		tt.T.Fatalf("unable to fetch kafka node types: %s", err)
	}

	// Find first available non-disabled node type
	for _, nodeType := range nodeTypesResp.NodeTypes {
		if !nodeType.Disabled && nodeType.StockStatus == kafkaSDK.NodeTypeStockAvailable {
			return nodeType.Name
		}
	}

	// Fallback: return first node type found
	if len(nodeTypesResp.NodeTypes) > 0 {
		return nodeTypesResp.NodeTypes[0].Name
	}

	tt.T.Fatal("no kafka node types available")
	return ""
}

func isClusterPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		api, region, id, err := kafka.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = api.GetCluster(&kafkaSDK.GetClusterRequest{
			Region:    region,
			ClusterID: id,
		}, scw.WithContext(context.Background()))

		return err
	}
}
