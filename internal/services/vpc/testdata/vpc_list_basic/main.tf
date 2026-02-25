resource "scaleway_vpc" "main" {
  region = "fr-par"
  name   = "test-vpc-fr-par"
  tags   = ["environment=production", "team=devops"]
}

resource "scaleway_vpc" "alt" {
  region = "nl-ams"
  name   = "test-vpc-nl-ams"
  tags   = ["environment=staging", "team=devops"]
}
