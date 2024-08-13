---
subcategory: "Domains and DNS"
page_title: "Scaleway: scaleway_domain_record"
---

# Resource: scaleway_domain_record

The `scaleway_domain_record` resource allows you to create and manage DNS records for Scaleway domains.

Refer to the Domains and DNS [product documentation](https://www.scaleway.com/en/docs/network/domains-and-dns/) and [API documentation](https://www.scaleway.com/en/developers/api/domains-and-dns/) for more information.

## Example Usage

### Create basic DNS records

The folllowing commands allow you to:

- create an A record for the `www.domain.tld` domain, pointing to `1.2.3.4` and another one pointing to `1.2.3.5`

- create an MX record with the `mx.online.net.` mail server and a priority of 10, and another one with the `mx-cache.online.net.` mail server and a priority of 20

```terraform
resource "scaleway_domain_record" "www" {
  dns_zone = "domain.tld"
  name     = "www"
  type     = "A"
  data     = "1.2.3.4"
  ttl      = 3600
}

resource "scaleway_domain_record" "www2" {
  dns_zone = "domain.tld"
  name     = "www"
  type     = "A"
  data     = "1.2.3.5"
  ttl      = 3600
}

resource "scaleway_domain_record" "mx" {
  dns_zone = "domain.tld"
  name     = ""
  type     = "MX"
  data     = "mx.online.net."
  ttl      = 3600
  priority = 10
}

resource "scaleway_domain_record" "mx2" {
  dns_zone = "domain.tld"
  name     = ""
  type     = "MX"
  data     = "mx-cache.online.net."
  ttl      = 3600
  priority = 20
}
```

### Create dynamic records

The folllowing commands allow you to:

- create a Geo IP record for `images.domain.tld` that points to different IPs based on the user's location: `1.2.3.5` for users in France (EU), and `4.3.2.1` for users in North America (NA)

- create an HTTP service record for `app.domain.tld` that checks the health of specified IPs and responds based on their status.

- create view-based records for `db.domain.tld` that resolve differently based on the client's subnet.

- create a weighted record for `web.domain.tld` that directs traffic to different IPs based on their weights.

```terraform
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
```

### Create an Instance and add records with the new Instance IP

The following commands allow you to:

- create a Scaleway Instance
- assign The Instance's IP address to various DNS records for a specified DNS zone

```terraform
variable "project_id" {
  type        = string
  description = "Your project ID."
}

variable "dns_zone" {
  type        = string
  description = "The DNS Zone used for testing records."
}

resource "scaleway_instance_ip" "public_ip" {
  project_id = var.project_id
}

resource "scaleway_instance_server" "web" {
  project_id = var.project_id
  type       = "DEV1-S"
  image      = "ubuntu_jammy"
  tags       = ["front", "web"]
  ip_id      = scaleway_instance_ip.public_ip.id

  root_volume {
    size_in_gb = 20
  }
}

resource "scaleway_domain_record" "web_A" {
  dns_zone = var.dns_zone
  name     = "web"
  type     = "A"
  data     = scaleway_instance_server.web.public_ip
  ttl      = 3600
}

resource "scaleway_domain_record" "web_cname" {
  dns_zone = var.dns_zone
  name     = "www"
  type     = "CNAME"
  data     = "web.${var.dns_zone}."
  ttl      = 3600
}

resource "scaleway_domain_record" "web_alias" {
  dns_zone = var.dns_zone
  name     = ""
  type     = "ALIAS"
  data     = "web.${var.dns_zone}."
  ttl      = 3600
}
```

## Arguments reference

The following arguments are supported:

- `dns_zone` - (Required) The DNS zone of the domain. If the domain has no DNS zone, one will be automatically created.

- `keep_empty_zone` - (Optional, defaults to `false`) When destroying a resource, if only NS records remain and this is set to `false`, the zone will be deleted. Note that each zone not deleted will [be billed](https://www.scaleway.com/en/dns/).

- `name` - (Optional) The name of the record (can be an empty string for a root record).

- `type` - (Required) The type of the record (`A`, `AAAA`, `MX`, `CNAME`, `DNAME`, `ALIAS`, `NS`, `PTR`, `SRV`, `TXT`, `TLSA`, or `CAA`).

- `data` - (Required) The content of the record (an IPv4 for an `A` record, a string for a `TXT` record, etc.).

- `ttl` - (Optional, defaults to `3600`) Time To Live of the record in seconds.

- `priority` - (Optional, defaults to `0`) The priority of the record (mostly used with an `MX` record).

### Dynamic records

- `geo_ip` - (Optional) The Geo IP provides DNS resolution based on the user’s geographical location. You can define a default IP that resolves if no Geo IP rule matches, and specify IPs for each geographical zone. [Check the documentation for more information](https://www.scaleway.com/en/docs/network/domains-and-dns/how-to/manage-dns-records/#geo-ip-records).
    - `matches` - (Required) The list of matches. *(Can be more than one)*.
        - `countries` - (Optional) List of countries (eg: `FR` for France, `US` for the United States, `GB` for Great Britain, etc.). [Check the list of all country codes](https://api.scaleway.com/domain-private/v2beta1/countries).
        - `continents` - (Optional) List of continents (eg: `EU` for Europe, `NA` for North America, `AS` for Asia, etc.). [Check the list of all continent codes](https://api.scaleway.com/domain-private/v2beta1/continents).
        - `data` (Required) The data of the match result.


- `http_service` - (Optional) The DNS service checks the provided URL on the configured IPs and resolves the request to one of the IPs, by excluding the ones not responding to the given string to check. [Check the documentation for more information](https://www.scaleway.com/en/docs/network/domains-and-dns/how-to/manage-dns-records/#healthcheck-records).
    - `ips` - (Required) List of IPs to check.
    - `must_contain` - (Required) Text to search.
    - `url` - (Required) URL to match the `must_contain` text to validate an IP.
    - `user_agent` - (Optional) User-agent used when checking the URL.
    - `strategy` - (Required) Strategy to return an IP from the IPs list. Can be `random`, `hashed`, or `all`.


- `view` - (Optional) The answer to a DNS request is based on the client’s (resolver) subnet. *(Can be more than 1)* [Check the documentation for more information](https://www.scaleway.com/en/docs/network/domains-and-dns/how-to/manage-dns-records/#views-records).
    - `subnet` - (Required) The subnet of the view.
    - `data` - (Required) The data of the view record.


- `weighted` - (Optional) You provide a list of IPs with their corresponding weights. These weights are used to proportionally direct requests to each IP. Depending on the weight of a record more or fewer requests are answered with their related IP compared to the others in the list. *(Can be more than 1)* [Check the documentation for more information](https://www.scaleway.com/en/docs/network/domains-and-dns/how-to/manage-dns-records/#weight-records).
    - `ip` - (Required) The weighted IP.
    - `weight` - (Required) The weight of the IP as an integer UInt32.

## Attributes reference

This section lists the attributes that are exported when the `scaleway_domain_record` resource is created:

- `id` - The ID of the record.

- `fqdn` - The FQDN of the record.

~> **Important:** Domain records' IDs are in the `{dns_zone}/{id}` format. The ID of a record should look like the following: `subdomain.domain.tld/11111111-1111-1111-1111-111111111111`.

## Multiple records

Some record types can have multiple data with the same name (e.g., `A`, `AAAA`, `MX`, `NS`, etc.). You can duplicate a `scaleway_domain_record`  resource with the same `name`, and the records will be added.

Note however, that some records (e.g., CNAME, multiple dynamic records of different types) must be unique.

## Import

This section explains how to import a record using the `{dns_zone}/{id}` format.

```bash
terraform import scaleway_domain_record.www subdomain.domain.tld/11111111-1111-1111-1111-111111111111
```
