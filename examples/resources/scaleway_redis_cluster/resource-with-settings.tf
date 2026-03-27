### Redis cluster with settings

resource "scaleway_redis_cluster" "main" {
  name      = "test_redis_basic"
  version   = "6.2.7"
  node_type = "RED1-MICRO"
  user_name = "my_initial_user"
  password  = "thiZ_is_v&ry_s3cret"

  settings = {
    "maxclients"    = "1000"
    "tcp-keepalive" = "120"
  }
}
