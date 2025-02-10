package iam

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func expandPermissionSetNames(rawPermissions interface{}) *[]string {
	permissions := []string{}
	permissionSet := rawPermissions.(*schema.Set)

	for _, rawPermission := range permissionSet.List() {
		permissions = append(permissions, rawPermission.(string))
	}

	return &permissions
}

func flattenPermissionSetNames(permissions []string) *schema.Set {
	rawPermissions := []interface{}(nil)
	for _, perm := range permissions {
		rawPermissions = append(rawPermissions, perm)
	}

	return schema.NewSet(func(i interface{}) int {
		return types.StringHashcode(i.(string))
	}, rawPermissions)
}

func expandPolicyRuleSpecs(d interface{}) []*iam.RuleSpecs {
	rules := []*iam.RuleSpecs(nil)

	rawRules := d.([]interface{})
	for _, rawRule := range rawRules {
		mapRule := rawRule.(map[string]interface{})
		rule := &iam.RuleSpecs{
			PermissionSetNames: expandPermissionSetNames(mapRule["permission_set_names"]),
			Condition:          mapRule["condition"].(string),
		}

		if orgID, orgIDExists := mapRule["organization_id"]; orgIDExists && orgID.(string) != "" {
			rule.OrganizationID = scw.StringPtr(orgID.(string))
		}

		if projIDs, projIDsExists := mapRule["project_ids"]; projIDsExists {
			rule.ProjectIDs = types.ExpandStringsPtr(projIDs)
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenPolicyRules(rules []*iam.Rule) interface{} {
	rawRules := []interface{}(nil)

	for _, rule := range rules {
		rawRule := map[string]interface{}{}
		if rule.OrganizationID != nil {
			rawRule["organization_id"] = types.FlattenStringPtr(rule.OrganizationID)
		} else {
			rawRule["organization_id"] = nil
		}

		if rule.ProjectIDs != nil {
			rawRule["project_ids"] = types.FlattenSliceString(*rule.ProjectIDs)
		}

		if rule.PermissionSetNames != nil {
			rawRule["permission_set_names"] = flattenPermissionSetNames(*rule.PermissionSetNames)
		}

		if rule.Condition != "" {
			rawRule["condition"] = rule.Condition
		}

		rawRules = append(rawRules, rawRule)
	}

	return rawRules
}
