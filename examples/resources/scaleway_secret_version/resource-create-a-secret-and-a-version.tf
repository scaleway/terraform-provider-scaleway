### Create a secret and a version

resource "scaleway_secret" "main" {
  name        = "foo"
  description = "barr"
  tags        = ["foo", "terraform"]
}

resource "scaleway_secret_version" "v1" {
  description = "version1"
  secret_id   = scaleway_secret.main.id
  data        = "my_new_secret"
}
