### Basic

resource "scaleway_autoscaling_instance_template" "main" {
  name            = "asg-template"
  commercial_type = "PLAY2-MICRO"
  tags            = ["terraform-test", "basic"]
  volumes {
    name        = "as-volume"
    volume_type = "sbs"
    boot        = true
    from_snapshot {
      snapshot_id = scaleway_block_snapshot.main.id
    }
    perf_iops = 5000
  }
  public_ips_v4_count = 1
  private_network_ids = [scaleway_vpc_private_network.main.id]
}
