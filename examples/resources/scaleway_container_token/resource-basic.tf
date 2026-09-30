resource "scaleway_container_namespace" "main" {
  name = "test-container-token-ns"
}

resource "scaleway_container" "main" {
  namespace_id = scaleway_container_namespace.main.id
}

// Namespace Token
resource "scaleway_container_token" "namespace" {
  namespace_id = scaleway_container_namespace.main.id
  expires_at   = "2022-10-18T11:35:15+02:00"
}

// Container Token
resource "scaleway_container_token" "container" {
  container_id = scaleway_container.main.id
}
