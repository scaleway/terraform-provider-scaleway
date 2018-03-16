## 1.2.1 (Unreleased)

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
