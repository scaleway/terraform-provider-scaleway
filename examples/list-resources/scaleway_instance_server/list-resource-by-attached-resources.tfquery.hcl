# List servers of a specific zone matching the filters (on the default project)
list "scaleway_instance_server" "by_attached_resources" {
  provider = scaleway

  config {
    zones = ["nl-ams-2"]
    security_group_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
    placement_group_ids = [scaleway_instance_placement_group.pg.id]
    private_network_ids = [
      scaleway_vpc_private_network.pn1.id,
      scaleway_vpc_private_network.pn2.id,
    ]
  }
}