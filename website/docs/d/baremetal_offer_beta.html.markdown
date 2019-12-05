---
layout: "scaleway"
page_title: "Scaleway: scaleway_baremetal_offer_beta"
description: |-
  Gets information about an Baremetal offer.
---

# scaleway_baremetal_offer_beta

Gets information about a baremetal offer. For more information, see [the documentation](https://developers.scaleway.com/en/products/baremetal/api).

## Example Usage

```hcl
// Get info by offer name
data "scaleway_baremetal_offer_beta" "my_offer" {
  zone = "fr-par-2"
  name = "HC-BM1-L"
}

// Get info by offer id
data "scaleway_baremetal_offer_beta" "my_offer" {
  zone     = "fr-par-2"
  offer_id = "3ab0dc29-2fd4-486e-88bf-d08fbf49214b"
}
```

## Argument Reference

- `name` - (Optional) The offer name. Only one of `name` and `offer_id` should be specified.

- `offer_id` - (Optional) The offer id. Only one of `name` and `offer_id` should be specified.

- `allow_disabled` - (Optional, default `false`) Include disabled offers.

- `zone` - (Defaults to [provider](../index.html#zone) `zone`) The [zone](../guides/regions_and_zones.html#zones) in which the offer should be created.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the offer.

- `bandwidth` - Available Bandwidth with the offer.

- `commercial_range` - Commercial range of the offer.

- `price_per_sixty_minutes` - Price of the offer for the next 60 minutes (a server order at 11h32 will be payed until 12h32).

- `price_per_month` - Price of the offer per months.

- `quota_name` - Quota name of this offer.

- `stock` - Stock status for this offer. Possible values are: `empty`, `low` or `available`.
