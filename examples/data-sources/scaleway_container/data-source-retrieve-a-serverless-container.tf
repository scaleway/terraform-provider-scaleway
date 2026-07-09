## Retrieve a Serverless Container

resource "scaleway_container_namespace" "main" {
}

resource "scaleway_container" "main" {
  name         = "test-container-data"
  namespace_id = scaleway_container_namespace.main.id
}

// Get info by container name
data "scaleway_container" "by_name" {
  namespace_id = scaleway_container_namespace.main.id
  name         = scaleway_container.main.name
}

// Get info by container ID
data "scaleway_container" "by_id" {
  namespace_id = scaleway_container_namespace.main.id
  container_id = scaleway_container.main.id
}
