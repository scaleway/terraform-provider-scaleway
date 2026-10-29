package kafka_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceCluster_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	latestVersion := fetchLatestKafkaVersion(tt)
	nodeType := fetchAvailableKafkaNodeType(tt)

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isClusterDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "scaleway_vpc" "main" {
  region = "fr-par"
  name   = "TestAccDataSourceCluster_Basic"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "tf_test_kafka_datasource_pn"
  region = "fr-par"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_kafka_cluster" "main" {
  name              = "tf-test-kafka-datasource"
  version           = "%s"
  node_amount       = 1
  node_type         = "%s"
  volume_type       = "sbs_5k"
  volume_size_in_gb = 10
  user_name         = "admin"
  password          = "password@1234567"

  private_network {
    pn_id = scaleway_vpc_private_network.pn01.id
  }
}

data "scaleway_kafka_cluster" "by_name" {
  name = scaleway_kafka_cluster.main.name
}

data "scaleway_kafka_cluster" "by_id" {
  cluster_id = scaleway_kafka_cluster.main.id
}
`, latestVersion, nodeType),
				Check: resource.ComposeTestCheckFunc(
					isClusterPresent(tt, "scaleway_kafka_cluster.main"),

					resource.TestCheckResourceAttr("data.scaleway_kafka_cluster.by_name", "name", "tf-test-kafka-datasource"),
					resource.TestCheckResourceAttrSet("data.scaleway_kafka_cluster.by_name", "id"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_cluster.by_name", "node_amount", "1"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_cluster.by_name", "volume_type", "sbs_5k"),
					resource.TestCheckResourceAttr("data.scaleway_kafka_cluster.by_name", "volume_size_in_gb", "10"),

					resource.TestCheckResourceAttr("data.scaleway_kafka_cluster.by_id", "name", "tf-test-kafka-datasource"),
					resource.TestCheckResourceAttrSet("data.scaleway_kafka_cluster.by_id", "id"),
					resource.TestCheckResourceAttrPair("data.scaleway_kafka_cluster.by_id", "id", "scaleway_kafka_cluster.main", "id"),
				),
			},
		},
	})
}
