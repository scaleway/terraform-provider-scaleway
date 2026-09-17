# Retrieve a dedicated connection by name
data "scaleway_interlink_dedicated_connection" "by_name" {
  name = "my-dedicated-connection"
}
