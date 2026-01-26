### MongoDB instance restored from Snapshot

resource "scaleway_mongodb_instance" "restored_instance" {
  snapshot_id = "${scaleway_vpc_private_network.pn.id}scaleway_mongodb_snapshot.main_snapshot.id"
  name        = "restored-mongodb-from-snapshot"
  node_type   = "MGDB-PLAY2-NANO"
  node_number = 1
}
