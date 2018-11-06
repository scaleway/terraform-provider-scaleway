## 1.8.0 (Unreleased)

FEATURES:

* **New Resource:**: `scaleway_ip_reverse_dns` ([#96](https://github.com/terraform-providers/terraform-provider-scaleway/pull/96))
* resource/scaleway_server: support cloudinit ([#97](https://github.com/terraform-providers/terraform-provider-scaleway/pull/97))
* resource/scaleway_security_group: support stateful security groups as well as default inbound and outbount policies ([#97](https://github.com/terraform-providers/terraform-provider-scaleway/pull/97))

## 1.7.0 (October 05, 2018)

FEATURES:

* **New Resource:** `scaleway_bucket` ([#94](https://github.com/terraform-providers/terraform-provider-scaleway/issues/94))

## 1.6.0 (August 28, 2018)

FEATURES:

* **New Data Source:** `scaleway_security_group` ([#78](https://github.com/terraform-providers/terraform-provider-scaleway/issues/78))
* **New Data Source:** `scaleway_volume` ([#77](https://github.com/terraform-providers/terraform-provider-scaleway/issues/77))
* resource/scaleway_image: support filtering by most recently created image ([#82](https://github.com/terraform-providers/terraform-provider-scaleway/pull/82))

BUG FIXES:

* resource/scaleway_token: fix compatability to changes in Scaleway API ([#86](https://github.com/terraform-providers/terraform-provider-scaleway/pull/86))
* resource/server: fix issue identifying restarts properly ([#87](https://github.com/terraform-providers/terraform-provider-scaleway/pull/87))

## 1.5.1 (July 11, 2018)

IMPROVEMENTS:

* provider: update documentation ([#75](https://github.com/terraform-providers/terraform-provider-scaleway/pull/75))

BUG FIXES:

* resource/scaleway_server & resource/scaleway_volume_attachment: race condition between startup & shutdown of servers ([#74](https://github.com/terraform-providers/terraform-provider-scaleway/pull/74))

## 1.5.0 (June 29, 2018)

IMPROVEMENTS:

* provider: update documentation ([#68](https://github.com/terraform-providers/terraform-provider-scaleway/pull/68), [#70](https://github.com/terraform-providers/terraform-provider-scaleway/pull/70))
* resource/scaleway_server: validate instance type availability ([#69](https://github.com/terraform-providers/terraform-provider-scaleway/pull/69))
* provider: update scaleway sdk ([#71](https://github.com/terraform-providers/terraform-provider-scaleway/pull/71))
* provider: allow concurrent creation of server resources ([#72](https://github.com/terraform-providers/terraform-provider-scaleway/pull/72))

## 1.4.1 (May 18, 2018)

BUG FIXES:

* resource/scaleway_server: fix server type validation ([#63](https://github.com/terraform-providers/terraform-provider-scaleway/pull/63))

## 1.4.0 (May 07, 2018)

IMPROVEMENTS:

* resource/scaleway_server: Update public_ip documentation ([#58](https://github.com/terraform-providers/terraform-provider-scaleway/pull/58))
* resource/scaleway_server: Add boot_type ([#59](https://github.com/terraform-providers/terraform-provider-scaleway/pull/59))

## 1.3.0 (April 11, 2018)

FEATURES:

* **New Resource:** `scaleway_token` ([#56](https://github.com/terraform-providers/terraform-provider-scaleway/pull/56))
* **New Resource:** `scaleway_user_data` ([#57](https://github.com/terraform-providers/terraform-provider-scaleway/pull/57))

IMPROVEMENTS:

* provider: update documentation ([#51](https://github.com/terraform-providers/terraform-provider-scaleway/pull/51),[#52](https://github.com/terraform-providers/terraform-provider-scaleway/pull/52))
* provider: update scaleway sdk ([#53](https://github.com/terraform-providers/terraform-provider-scaleway/pull/53), [#54](https://github.com/terraform-providers/terraform-provider-scaleway/pull/54), [#55](https://github.com/terraform-providers/terraform-provider-scaleway/pull/55))

BUG FIXES:

* provider: fix crash when working over slow and unreliable network connection ([#49](https://github.com/terraform-providers/terraform-provider-scaleway/pull/49))

## 1.2.0 (March 15, 2018)

IMPROVEMENTS:

* resource/scaleway_ip: Add support for setting reverse DNS field ([#48](https://github.com/terraform-providers/terraform-provider-scaleway/pull/48))
* resource/scaleway_ssh_key: Add new resource to manage ssh keys ([#47](https://github.com/terraform-providers/terraform-provider-scaleway/pull/47))

## 1.1.0 (February 27, 2018)

BUG FIXES:

* resource/scaleway_server: Fix crash with stopped server and ipv6 enabled ([#44](https://github.com/terraform-providers/terraform-provider-scaleway/issues/44))

IMPROVEMENTS:

* resource/scaleway_security_group: Add `enable_default_security` attribute to manage Scaleway default security group rules ([#43](https://github.com/terraform-providers/terraform-provider-scaleway/issues/43))

## 1.0.1 (January 15, 2018)

BUG FIXES:

* resource/scaleway_security_group_rule: Fix error when using count ([#25](https://github.com/terraform-providers/terraform-provider-scaleway/issues/25))
* provider: Retry rate-limited API requests ([#35](https://github.com/terraform-providers/terraform-provider-scaleway/issues/35))

IMPROVEMENTS:

* resource/scaleway_server: Validate types against scaleway offerings ([#17](https://github.com/terraform-providers/terraform-provider-scaleway/issues/17))

## 1.0.0 (October 25, 2017)

BUG FIXES:

* data-source/scaleway_bootscript: Fix crash when no filter is specified ([#21](https://github.com/terraform-providers/terraform-provider-scaleway/issues/21))

IMPROVEMENTS:

* resource/scaleway_server: Allow initial volumes without size to improve module support ([#19](https://github.com/terraform-providers/terraform-provider-scaleway/issues/19))

## 0.1.1 (August 04, 2017)

BUG FIXES:

* resource/scaleway_volume_attachment: Fix volume_attachment deletion ([#13](https://github.com/terraform-providers/terraform-provider-scaleway/issues/13))

IMPROVEMENTS:

* resource/scaleway_server: Improve public_ip attachment ([#14](https://github.com/terraform-providers/terraform-provider-scaleway/issues/14))

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
