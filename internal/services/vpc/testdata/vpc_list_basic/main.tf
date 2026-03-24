resource "scaleway_account_project" "main" {

}

resource "scaleway_vpc" "main" {
  project_id= scaleway_account_project.main.id
  region = "fr-par"
  name   = "test-vpc-fr-par"
}

resource "scaleway_vpc" "alt" {
  project_id= scaleway_account_project.main.id
  region = "nl-ams"
  name   = "test-vpc-nl-ams"
}
