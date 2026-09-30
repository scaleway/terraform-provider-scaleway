## Basic

data "scaleway_baremetal_partition_schema" "default" {
  offer_id         = "11111111-1111-1111-1111-111111111111"
  os_id            = "22222222-2222-2222-2222-222222222222"
  swap             = true
  extra_partition  = true
  ext_4_mountpoint = "/data"
}
