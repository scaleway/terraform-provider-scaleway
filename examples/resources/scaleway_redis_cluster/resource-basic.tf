### Basic Redis cluster creation

resource "scaleway_redis_cluster" "main" {
  name         = "test_redis_basic"
  version      = "6.2.7"
  node_type    = "RED1-MICRO"
  user_name    = "my_initial_user"
  password     = "thiZ_is_v&ry_s3cret"
  tags         = ["test", "redis"]
  cluster_size = 1
  tls_enabled  = "true"

  acl {
    ip          = "0.0.0.0/0"
    description = "Allow all"
  }
}
