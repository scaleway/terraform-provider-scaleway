## Example Usage

# Get info by name
data "scaleway_redis_cluster" "my_cluster" {
  name = "foobar"
}

# Get info by cluster ID
data "scaleway_redis_cluster" "my_cluster" {
  cluster_id = "11111111-1111-1111-1111-111111111111"
}
