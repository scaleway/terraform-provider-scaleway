---
subcategory: "Domains and DNS"
page_title: "Scaleway: scaleway_domain_record"
---

# scaleway_domain_record

The `scaleway_domain_record` data source is used to get information about an existing domain record.

Refer to the Domains and DNS [product documentation](https://www.scaleway.com/en/docs/network/domains-and-dns/) and [API documentation](https://www.scaleway.com/en/developers/api/domains-and-dns/) for more information.


## Query domain records

The following commands allow you to:

- query a domain record specified by the DNS zone (`domain.tld`), the record name (`www`), the record type (`A`), and the record content (`1.2.3.4`).
- query a domain record specified by the DNS zone (`domain.tld`) and the unique record ID (`11111111-1111-1111-1111-111111111111`).

```hcl
# Query record by DNS zone, record name, type and content
data "scaleway_domain_record" "by_content" {
  dns_zone = "domain.tld"
  name     = "www"
  type     = "A"
  data     = "1.2.3.4"
}

# Query record by DNS zone and record ID
data "scaleway_domain_record" "by_id" {
  dns_zone  = "domain.tld"
  record_id = "11111111-1111-1111-1111-111111111111"
}
```

## Arguments reference

This section lists the arguments that you can provide to the `scaleway_domain_record` data source to filter and retrieve the desired record:

- `dns_zone` - (Optional) The DNS zone (domain) to which the record belongs. This is a required field in both examples above but is optional in the context of defining the data source.

- `name` - (Required when not using `record_id`) The name of the record, which can be an empty string for a root record. Cannot be used with `record_id`.

- `type` - (Required when not using `record_id`) The type of the record (`A`, `AAAA`, `MX`, `CNAME`, etc.). Cannot be used with `record_id`.

- `data` - (Required when not using `record_id`) The content of the record (e.g., an IPv4 address for an `A` record or a string for a `TXT` record). Cannot be used with `record_id`.

- `record_id` - (Optional) The unique identifier of the record. Cannot be used with `name`, `type`, and `data`.

- `project_id` - (Defaults to the Project ID specified in the [provider configuration](../index.md#project_id)). The ID of the Project associated with the domain.

## Attributes reference

This section lists the attributes that are exported when the `scaleway_domain_record` data source is created. These attributes can be referenced in other parts of your Terraform configuration:

- `id` - The unique identifier of the record.

~> **Important:** Domain records' IDs are formatted as `{dns_zone}/{id}` (e.g. `subdomain.domain.tld/11111111-1111-1111-1111-111111111111`).

- `ttl` - The Time To Live (TTL) of the record in seconds.
- `priority` - The priority of the record, mainly used with `MX` records.
- `geo_ip` - Information about dynamic records based on user geolocation. [Find out more about dynamic records](../resources/domain_record.md#dynamic-records).
- `http_service` - Information about dynamic records based on URL resolution. [Find out more about dynamic records](../resources/domain_record.md#dynamic-records).
- `weighted` - Information about dynamic records based on IP weights. [Find out more about dynamic records](../resources/domain_record.md#dynamic-records).
- `view` - Information about dynamic records based on the client’s (resolver) subnet. [Find out more about dynamic records](../resources/domain_record.md#dynamic-records).
