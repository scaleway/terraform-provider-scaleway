data "scaleway_datalabs" "filtered" {
  region = "fr-par"
  name   = "my-datalab"
  tags   = ["production"]
}
