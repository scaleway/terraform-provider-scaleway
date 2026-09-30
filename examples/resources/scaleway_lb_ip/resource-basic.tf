### Basic

resource "scaleway_lb_ip" "ip" {
  reverse = "my-reverse.com"
}
