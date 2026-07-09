resource "scaleway_container_namespace" "main" {
  name        = "main-container-namespace"
  description = "Main container namespace"
  tags        = ["tag1", "tag2"]
}
