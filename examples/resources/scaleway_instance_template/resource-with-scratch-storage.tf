resource "scaleway_instance_template" "main" {
  name        = "instance-template-with-scratch-storage"
  server_type = "L40S-1-48G"
  zone        = "fr-par-2"

  volumes = [
    {
      volume_type = "sbs"
      image_label = "ubuntu_noble_gpu_os_13_nvidia"
      size_in_gb  = 30
      perf_iops   = 15000
    },
    {
      volume_type = "scratch"
      size_in_gb  = 300
    }
  ]
}
