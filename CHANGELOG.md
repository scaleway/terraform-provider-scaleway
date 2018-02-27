## 1.0.2 (Unreleased)

* r/server: fix crash with stopped server and ipv6 enabled. ([#44](https://github.com/terraform-providers/terraform-provider-scaleway/pull/44))
* r/security_group: add `enable_default_security` attribute to manage Scaleway default security group rules ([#43](https://github.com/terraform-providers/terraform-provider-scaleway/pull/43))

## 1.0.1 (January 15, 2018)

* r/server: validate types against scaleway offerings ([#17](https://github.com/terraform-providers/terraform-provider-scaleway/issues/17))
* r/security_group_rule: fix error when using count ([#25](https://github.com/terraform-providers/terraform-provider-scaleway/issues/25))
* retry rate-limited API requests ([#35](https://github.com/terraform-providers/terraform-provider-scaleway/issues/35))

## 1.0.0 (October 25, 2017)

* d/bootscript: fix crash when no filter is specified ([#21](https://github.com/terraform-providers/terraform-provider-scaleway/issues/21))
* r/server: allow initial volumes without size to improve module support ([#19](https://github.com/terraform-providers/terraform-provider-scaleway/issues/19))

## 0.1.1 (August 04, 2017)

* r/server: improve public_ip attachment ([#14](https://github.com/terraform-providers/terraform-provider-scaleway/issues/14))
* r/volume_attachment: fix volume_attachment deletion ([#13](https://github.com/terraform-providers/terraform-provider-scaleway/issues/13))

## 0.1.0 (June 21, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
