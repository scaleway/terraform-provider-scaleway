### Basic

resource "scaleway_file_filesystem" "file" {
  name       = "my-nfs-filesystem"
  size_in_gb = 100
}
