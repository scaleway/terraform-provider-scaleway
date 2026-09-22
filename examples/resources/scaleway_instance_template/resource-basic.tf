resource "scaleway_instance_template" "main" {
  name = "instance-template-basic"
  tags = ["terraform-examples", "scaleway_instance_template", "basic"]

  server_type       = "PRO2-M"
  server_tags       = ["created-from-template"]
  public_ipv4_count = 1
  public_ipv6_count = 3
}
