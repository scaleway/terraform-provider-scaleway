package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SecurityGroupRule definition
type SecurityGroupRule struct {
	Direction    string `json:"direction"`
	Protocol     string `json:"protocol"`
	IPRange      string `json:"ip_range"`
	DestPortFrom int    `json:"dest_port_from,omitempty"`
	Action       string `json:"action"`
	Position     int    `json:"position"`
	DestPortTo   string `json:"dest_port_to"`
	Editable     bool   `json:"editable"`
	ID           string `json:"id"`
}

// GetSecurityGroupRules represents the response of a GET /security_group/{groupID}/rules
type GetSecurityGroupRules struct {
	Rules []SecurityGroupRule `json:"rules"`
}

// GetSecurityGroupRule represents the response of a GET /security_group/{groupID}/rules/{ruleID}
type GetSecurityGroupRule struct {
	Rules SecurityGroupRule `json:"rule"`
}

// NewSecurityGroupRule definition POST/PUT request /security_group/{groupID}
type NewSecurityGroupRule struct {
	Action       string `json:"action"`
	Direction    string `json:"direction"`
	IPRange      string `json:"ip_range"`
	Protocol     string `json:"protocol"`
	DestPortFrom int    `json:"dest_port_from,omitempty"`
}

// GetSecurityGroupRules returns a SecurityGroupRules
func (s *ScalewayAPI) GetSecurityGroupRules(groupID string) (*GetSecurityGroupRules, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, fmt.Sprintf("security_groups/%s/rules", groupID), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var securityGroupRules GetSecurityGroupRules

	if err = json.Unmarshal(body, &securityGroupRules); err != nil {
		return nil, err
	}
	return &securityGroupRules, nil
}

// GetASecurityGroupRule returns a SecurityGroupRule
func (s *ScalewayAPI) GetASecurityGroupRule(groupID string, rulesID string) (*GetSecurityGroupRule, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, fmt.Sprintf("security_groups/%s/rules/%s", groupID, rulesID), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var securityGroupRules GetSecurityGroupRule

	if err = json.Unmarshal(body, &securityGroupRules); err != nil {
		return nil, err
	}
	return &securityGroupRules, nil
}

type postSecurityGroupRuleResponse struct {
	SecurityGroupRule SecurityGroupRule `json:"rule"`
}

// PostSecurityGroupRule posts a rule on a server
func (s *ScalewayAPI) PostSecurityGroupRule(SecurityGroupID string, rules NewSecurityGroupRule) (*SecurityGroupRule, error) {
	resp, err := s.PostResponse(s.computeAPI, fmt.Sprintf("security_groups/%s/rules", SecurityGroupID), rules)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := s.handleHTTPError([]int{http.StatusCreated}, resp)
	if err != nil {
		return nil, err
	}
	var res postSecurityGroupRuleResponse
	err = json.Unmarshal(data, &res)
	return &res.SecurityGroupRule, err
}

// PutSecurityGroupRule updates a SecurityGroupRule
func (s *ScalewayAPI) PutSecurityGroupRule(rules NewSecurityGroupRule, securityGroupID, RuleID string) error {
	resp, err := s.PutResponse(s.computeAPI, fmt.Sprintf("security_groups/%s/rules/%s", securityGroupID, RuleID), rules)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = s.handleHTTPError([]int{http.StatusOK}, resp)
	return err
}

// DeleteSecurityGroupRule deletes a SecurityGroupRule
func (s *ScalewayAPI) DeleteSecurityGroupRule(SecurityGroupID, RuleID string) error {
	resp, err := s.DeleteResponse(s.computeAPI, fmt.Sprintf("security_groups/%s/rules/%s", SecurityGroupID, RuleID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = s.handleHTTPError([]int{http.StatusNoContent}, resp)
	return err
}
