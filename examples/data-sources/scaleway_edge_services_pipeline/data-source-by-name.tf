# Retrieve an Edge Services pipeline by name
data "scaleway_edge_services_pipeline" "by_name" {
  name = "my-pipeline"
}
