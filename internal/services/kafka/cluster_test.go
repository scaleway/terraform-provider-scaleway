package kafka_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	kafkaSDK "github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/kafka"
)

func TestAccCluster_Basic(t *testing.T) {
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
  name   = "TestAccCluster_Basic"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "tf_test_kafka_basic_pn"
  region = "fr-par"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_kafka_cluster" "main" {
  name              = "tf-test-kafka-basic"
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
`, latestVersion, nodeType),
				Check: resource.ComposeTestCheckFunc(
					isClusterPresent(tt, "scaleway_kafka_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "name", "tf-test-kafka-basic"),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "version", latestVersion),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "node_amount", "1"),
					resource.TestCheckResourceAttrSet("scaleway_kafka_cluster.main", "node_type"),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "volume_type", "sbs_5k"),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "volume_size_in_gb", "10"),

					// Private network is present
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "private_network.#", "1"),
					resource.TestCheckResourceAttrSet("scaleway_kafka_cluster.main", "private_network.0.pn_id"),
					resource.TestCheckResourceAttrSet("scaleway_kafka_cluster.main", "private_network.0.id"),
				),
			},

			{
				// Update tags and name
				Config: fmt.Sprintf(`
resource "scaleway_vpc" "main" {
  region = "fr-par"
  name   = "TestAccCluster_Basic"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "tf_test_kafka_basic_pn"
  region = "fr-par"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_kafka_cluster" "main" {
  name              = "tf-test-kafka-updated"
  version           = "%s"
  node_amount       = 1
  node_type         = "%s"
  volume_type       = "sbs_5k"
  volume_size_in_gb = 10
  tags              = ["tag1", "tag2"]
  user_name         = "admin"
  password          = "password@1234567"

  private_network {
    pn_id = scaleway_vpc_private_network.pn01.id
  }
}
`, latestVersion, nodeType),
				Check: resource.ComposeTestCheckFunc(
					isClusterPresent(tt, "scaleway_kafka_cluster.main"),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "name", "tf-test-kafka-updated"),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "tags.#", "2"),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "tags.1", "tag2"),

					// Private network still present
					resource.TestCheckResourceAttr("scaleway_kafka_cluster.main", "private_network.#", "1"),
				),
			},
		},
	})
}

func isClusterDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_kafka_cluster" {
				continue
			}

			api, region, id, err := kafka.NewAPIWithRegionAndID(tt.Meta, rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = api.GetCluster(&kafkaSDK.GetClusterRequest{
				Region:    region,
				ClusterID: id,
			}, scw.WithContext(context.Background()))
			if err == nil {
				return fmt.Errorf("cluster %s still exists", id)
			}

			if !httperrors.Is404(err) {
				return err
			}
		}

		return nil
	}
}
