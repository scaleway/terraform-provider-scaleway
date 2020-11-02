---
page_title: "Scaleway Zones and Regions"
description: |-
  Scaleway resources can be created in availability zones and regions.
---

# Scaleway Zones and Regions

Scaleway's products are deployed across multiple datacenter in the world.

For technical and legal reasons, some products are splitted by Region or by Availability Zones.
When using such product, you can choose the location that better fits your need (country, latency, ...).

## Regions

A Region is represented as a Geographical area such as France (Paris: `fr-par`) or the Netherlands (Amsterdam: `nl-ams`).
It can contain multiple Availability Zones.


## Zones

In order to deploy highly available application, a region can be splitted in many Availability Zones (AZ).
Latency between multiple AZ of the same region are low as they have a common network layer.

List of availability zones by regions:

- France - Paris (`fr-par`)
    - `fr-par-1`
    - `fr-par-2`
- The Netherlands - Amsterdam (`nl-ams`)
    - `nl-ams-1`
- Poland - Warsaw (`pl-waw`)
    - `pl-waw-1`

## Resource IDs

To save this notion of regions and zones in the state, all the Terraform IDs of Scaleway contain the region or zone.
This is saved in the following format: `{zone|region}/{resource_id}`.
Where `zone` or `region` is the place where the resource is created and where `resource_id` is the ID that is used on Scaleway's console/API.

---

More information regarding zones and regions can be found [here](https://developers.scaleway.com/en/quickstart/#region-and-zone).
