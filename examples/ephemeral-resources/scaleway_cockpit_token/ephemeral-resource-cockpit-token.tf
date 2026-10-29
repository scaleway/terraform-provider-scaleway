ephemeral "scaleway_cockpit_token" "main" {
  name = "my-ephemeral-cockpit-token"
  scopes {
    write_metrics = true
    write_logs    = true
  }
}
