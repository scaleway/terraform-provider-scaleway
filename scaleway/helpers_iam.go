package scaleway

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// instanceAPIWithZone returns a new iam API for a Create request
func iamAPI(m interface{}) *iam.API {
	meta := m.(*Meta)
	return iam.NewAPI(meta.scwClient)
}

func getOrganizationID(m interface{}, d *schema.ResourceData) *string {
	meta := m.(*Meta)

	orgID, orgIDExist := d.GetOk("organization_id")

	if orgIDExist {
		return expandStringPtr(orgID)
	}

	defaultOrgID, defaultOrgIDExists := meta.scwClient.GetDefaultOrganizationID()
	if defaultOrgIDExists {
		return expandStringPtr(defaultOrgID)
	}

	return nil
}

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
		return StringHashcode(i.(string))
	}, rawPermissions)
}

func expandPolicyRuleSpecs(d interface{}) []*iam.RuleSpecs {
	rules := []*iam.RuleSpecs(nil)
	rawRules := d.(*schema.Set)
	for _, rawRule := range rawRules.List() {
		mapRule := rawRule.(map[string]interface{})
		rule := &iam.RuleSpecs{
			PermissionSetNames: expandPermissionSetNames(mapRule["permission_set_names"]),
		}
		if orgID, orgIDExists := mapRule["organization_id"]; orgIDExists && orgID.(string) != "" {
			rule.OrganizationID = scw.StringPtr(orgID.(string))
		}
		if projIDs, projIDsExists := mapRule["project_ids"]; projIDsExists {
			rule.ProjectIDs = expandStringsPtr(projIDs)
		}
		rules = append(rules, rule)
	}
	return rules
}

func iamPolicyRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if orgID, hasOrgID := m["organization_id"]; hasOrgID && orgID != nil {
		buf.WriteString(fmt.Sprintf("%s-", orgID.(string)))
	}
	if projIDs, hasProjIDs := m["project_ids"]; hasProjIDs && projIDs != nil {
		projIDList := projIDs.([]interface{})
		for _, projID := range projIDList {
			buf.WriteString(fmt.Sprintf("%s-", projID.(string)))
		}
	}
	if permSet, hasPermSet := m["permission_set_names"]; hasPermSet {
		permSetNames := permSet.(*schema.Set)
		for _, permName := range permSetNames.List() {
			buf.WriteString(fmt.Sprintf("%s-", permName.(string)))
		}
	}
	return StringHashcode(buf.String())
}

func flattenPolicyRules(rules []*iam.Rule) interface{} {
	rawRules := []interface{}(nil)
	for _, rule := range rules {
		rawRule := map[string]interface{}{}
		if rule.OrganizationID != nil {
			rawRule["organization_id"] = flattenStringPtr(rule.OrganizationID)
		} else {
			rawRule["organization_id"] = nil
		}
		if rule.ProjectIDs != nil {
			rawRule["project_ids"] = flattenSliceString(*rule.ProjectIDs)
		}
		if rule.PermissionSetNames != nil {
			rawRule["permission_set_names"] = flattenPermissionSetNames(*rule.PermissionSetNames)
		}
		rawRules = append(rawRules, rawRule)
	}
	return schema.NewSet(iamPolicyRuleHash, rawRules)
}
