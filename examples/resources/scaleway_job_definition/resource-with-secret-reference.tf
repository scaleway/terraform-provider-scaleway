### With Secret Reference

resource "scaleway_job_definition" "main" {
  name         = "testjob"
  cpu_limit    = 140
  memory_limit = 256
  image_uri    = "docker.io/alpine:latest"
  command      = "ls"
  timeout      = "10m"

  cron {
    schedule = "5 4 1 * *" # cron at 04:05 on day-of-month 1
    timezone = "Europe/Paris"
  }

  secret_reference {
    secret_id = "11111111-1111-1111-1111-111111111111"
    file      = "/home/dev/secret_file"
  }

  secret_reference {
    secret_id      = scaleway_secret.job_secret.id
    secret_version = "1"
    environment    = "FOO"
  }
}
