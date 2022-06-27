package scaleway

import (
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// instanceAPIWithZone returns a new iam API for a Create request
func iamAPI(m interface{}) *iam.API {
	meta := m.(*Meta)
	return iam.NewAPI(meta.scwClient)
}

func expandPolicyRuleSpecs(d interface{}) []*iam.RuleSpecs {
	rules := []*iam.RuleSpecs(nil)
	rawRules := d.([]interface{})
	for _, rawRule := range rawRules {
		mapRule := rawRule.(map[string]interface{})
		rule := &iam.RuleSpecs{
			OrganizationID: scw.StringPtr(mapRule["organization_id"].(string)),
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
			rawRule["organization_id"] = flattenStringPtr(rule.OrganizationID)
		}
		rawRules = append(rawRules, rawRule)
	}
	return rawRules
}
