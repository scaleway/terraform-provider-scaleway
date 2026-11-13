data "scaleway_messageq_version" "latest" {
  name = "latest"
}

data "scaleway_messageq_version" "specific" {
  name = "4.0"
}
