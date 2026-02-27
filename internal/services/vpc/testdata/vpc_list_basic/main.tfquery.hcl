list "scaleway_vpc" "all" {
  provider = scaleway

  config {
    region = "all"
  }
}

list "scaleway_vpc" "fr-par" {
  provider = scaleway

  config {
    region = "fr-par"
    tags = ["environment=production"]
  }
}

list "scaleway_vpc" "by_name" {
  provider = scaleway

  config {
    region = "all"
    name = "test-vpc*"
  }
}