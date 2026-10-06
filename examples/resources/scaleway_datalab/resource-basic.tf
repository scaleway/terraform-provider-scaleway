resource "scaleway_vpc" "main" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "main" {
  vpc_id = scaleway_vpc.main.id
  region = "fr-par"
}

resource "scaleway_datalab" "main" {
  name               = "my-datalab"
  spark_version      = "4.0.0"
  private_network_id = scaleway_vpc_private_network.main.id
  region             = "fr-par"

  main = {
    node_type = "DATALAB-SHARED-4C-8G"
  }

  worker = {
    node_type  = "DATALAB-DEDICATED2-2C-8G"
    node_count = 1
  }
}
