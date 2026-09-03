resource "scaleway_container_namespace" "main" {
}

resource "scaleway_container" "main" {
  name         = "my-container-with-cron-tf"
  namespace_id = scaleway_container_namespace.main.id
}

resource "scaleway_container_cron" "main" {
  container_id = scaleway_container.main.id
  name         = "my-cron-name"
  schedule     = "5 4 1 * *" #cron at 04:05 on day-of-month 1
  args = jsonencode(
    {
      address = {
        city    = "Paris"
        country = "FR"
      }
      age       = 23
      firstName = "John"
      isAlive   = true
      lastName  = "Smith"
      # minScale: 1
      # memoryLimit: 256
      # maxScale: 2
      # timeout: 20000
      # Local environment variables - used only in given function
    }
  )
}
