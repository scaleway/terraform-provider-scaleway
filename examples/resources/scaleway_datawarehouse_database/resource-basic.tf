### Basic

resource "scaleway_datawarehouse_deployment" "main" {
  name          = "my-datawarehouse"
  version       = "v25"
  replica_count = 1
  cpu_min       = 2
  cpu_max       = 4
  ram_per_cpu   = 4
  password      = "thiZ_is_v&ry_s3cret"
}

resource "scaleway_datawarehouse_database" "main" {
  deployment_id = scaleway_datawarehouse_deployment.main.id
  name          = "my_database"
}
