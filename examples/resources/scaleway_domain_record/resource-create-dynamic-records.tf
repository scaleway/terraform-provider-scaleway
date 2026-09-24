### Create dynamic records

resource "scaleway_domain_record" "geo_ip" {
  dns_zone = "domain.tld"
  name     = "images"
  type     = "A"
  data     = "1.2.3.4"
  ttl      = 3600

  geo_ip {
    matches {
      continents = ["EU"]
      countries  = ["FR"]
      data       = "1.2.3.5"
    }

    matches {
      continents = ["NA"]
      data       = "4.3.2.1"
    }
  }
}

resource "scaleway_domain_record" "http_service" {
  dns_zone = "domain.tld"
  name     = "app"
  type     = "A"
  data     = "1.2.3.4"
  ttl      = 3600

  http_service {
    ips          = ["1.2.3.5", "1.2.3.6"]
    must_contain = "up"
    url          = "http://mywebsite.com/health"
    user_agent   = "scw_service_up"
    strategy     = "hashed"
  }
}

resource "scaleway_domain_record" "view" {
  dns_zone = "domain.tld"
  name     = "db"
  type     = "A"
  data     = "1.2.3.4"
  ttl      = 3600

  view {
    subnet = "100.0.0.0/16"
    data   = "1.2.3.5"
  }

  view {
    subnet = "100.1.0.0/16"
    data   = "1.2.3.6"
  }
}

resource "scaleway_domain_record" "weighted" {
  dns_zone = "domain.tld"
  name     = "web"
  type     = "A"
  data     = "1.2.3.4"
  ttl      = 3600

  weighted {
    ip     = "1.2.3.5"
    weight = 1
  }

  weighted {
    ip     = "1.2.3.6"
    weight = 2
  }
}
