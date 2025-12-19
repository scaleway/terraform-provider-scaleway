Creates and manages Scaleway compute Instance security group rules. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/instance/#path-security-groups-list-security-groups).

This resource can be used to externalize rules from a `scaleway_instance_security_group` to solve circular dependency problems. When using this resource do not forget to set `external_rules = true` on the security group.

~> **Warning:** In order to guaranty rules order in a given security group only one scaleway_instance_security_group_rules is allowed per security group.