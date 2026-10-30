resource "scaleway_instance_template" "main" {
  name        = "instance-template-with-cloud-init"
  server_type = "PLAY2-PICO"

  volumes = [
    {
      volume_type = "sbs"
      image_label = "ubuntu_noble"
      size_in_gb  = 10
      perf_iops   = 5000
    }
  ]

  cloud_init = <<-EOT
    #cloud-config
    packages:
      - nginx
  EOT
}
