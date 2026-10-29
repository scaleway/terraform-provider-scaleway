# Retrieve a dedicated connection by its ID
data "scaleway_interlink_dedicated_connection" "by_id" {
  connection_id = "11111111-1111-1111-1111-111111111111"
}
