### Get task_id for import

data "scaleway_domain_registration" "example" {
  domain_name = "example.com"
}

output "import_command" {
  value = "terraform import scaleway_domain_registration.example ${data.scaleway_domain_registration.example.project_id}/${data.scaleway_domain_registration.example.task_id}"
}
