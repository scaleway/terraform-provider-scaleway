### With project_id filter

data "scaleway_domain_registration" "example" {
  domain_name = "example.com"
  project_id  = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
