---
page_title: "Scaleway: scaleway_webhosting_offer"
description: |-
Gets information about a webhosting offer.
---

# scaleway_webhosting_offer

Gets information about a webhosting offer.

## Example Usage

```hcl
# Get info by offer name
data "scaleway_webhosting_offer" "by_name" { 
  name = "performance"
}

# Get info by offer id
data "scaleway_webhosting_offer" "by_id" {
  offer_id = "de2426b4-a9e9-11ec-b909-0242ac120002"
}
```

## Argument Reference

- `name` - (Optional) The offer name. Only one of `name` and `offer_id` should be specified.

- `offer_id` - (Optional) The offer id. Only one of `name` and `offer_id` should be specified.

- `region` - (Defaults to [provider](../index.md#zone) `region`) The [region](../guides/regions_and_zones.md#zones) in which offer exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

