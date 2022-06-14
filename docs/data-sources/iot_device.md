---
layout: "scaleway"
page_title: "Scaleway: scaleway_iot_device"
description: |-
Gets information about an IOT Device.
---

# scaleway_iot_device

Gets information about an IOT Device.

## Example Usage

```hcl
# Get info by name 
data "scaleway_iot_device" "my_device" {
  name = "foobar"
}

# Get info by name and hub_id
data "scaleway_iot_device" "my_device" {
  name = "foobar"
  hub_id = "11111111-1111-1111-1111-111111111111"
}

# Get info by device ID
data "scaleway_iot_device" "my_device" {
  device_id = "11111111-1111-1111-1111-111111111111"
}

```

## Argument Reference

- `name` - (Optional) The name of the Hub.
  Only one of the `name` and `device_id` should be specified.

- `hub_id` - (Optional) The hub ID.

- `device_id` - (Optional) The device ID.
  Only one of the `name` and `device_id` should be specified.

- `region` - (Default to [provider](../index.md) `region`) The [region](../guides/regions_and_zones.md#zones) in which the hub exists.
