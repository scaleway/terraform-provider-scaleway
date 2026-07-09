## With ACLs

resource "scaleway_lb_frontend" "frontend01" {
  lb_id        = scaleway_lb.lb01.id
  backend_id   = scaleway_lb_backend.backend01.id
  name         = "frontend01"
  inbound_port = "80"

  # Allow downstream requests from: 192.168.0.1, 192.168.0.2 or 192.168.10.0/24
  acl {
    name = "blacklist wellknwon IPs"
    action {
      type = "allow"
    }
    match {
      ip_subnet = ["192.168.0.1", "192.168.0.2", "192.168.10.0/24"]
    }
  }

  # Deny downstream requests from: 51.51.51.51 that match "^foo*bar$"
  acl {
    action {
      type = "deny"
    }
    match {
      ip_subnet         = ["51.51.51.51"]
      http_filter       = "regex"
      http_filter_value = ["^foo*bar$"]
    }
  }

  # Allow downstream http requests that begins with "/foo" or "/bar"
  acl {
    action {
      type = "allow"
    }
    match {
      http_filter       = "path_begin"
      http_filter_value = ["foo", "bar"]
    }
  }

  # Allow upstream http requests that DO NOT begins with "/hi"
  acl {
    action {
      type = "allow"
    }
    match {
      http_filter       = "path_begin"
      http_filter_value = ["hi"]
      invert            = "true"
    }
  }

  # Allow upstream http requests that have an HTTP header "foo" that matches "bar"
  acl {
    action {
      type = "allow"
    }

    match {
      http_filter        = "http_header_match"
      http_filter_value  = "foo"
      http_filter_option = "bar"
    }
  }

  # Redirect requests from IP 10.0.0.10 and path beginning with "foo" or "bar" to "https://example.com" using a 307 redirect
  acl {
    action {
      type = "redirect"
      redirect {
        type   = "location"
        target = "https://example.com"
        code   = 307
      }
    }
    match {
      ip_subnet         = ["10.0.0.10"]
      http_filter       = "path_begin"
      http_filter_value = ["foo", "bar"]
    }
  }
}
