### Basic

resource "scaleway_vpc" "main" {
  region = "fr-par"
  name   = "kafka-datasource-vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  name   = "tf_test_kafka_datasource_pn"
  region = "fr-par"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_kafka_cluster" "main" {
  name              = "tf-test-kafka-datasource"
  version           = "4.1.1"
  node_amount       = 1
  node_type         = "kafka.ls.1"
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
