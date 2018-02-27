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

// GetSecurityGroupRules represents the response of a GET /_group/{groupID}/rules
type GetSecurityGroupRules struct {
	Rules []SecurityGroupRule `json:"rules"`
}

// GetSecurityGroupRule represents the response of a GET /_group/{groupID}/rules/{ruleID}
type GetSecurityGroupRule struct {
	Rules SecurityGroupRule `json:"rule"`
}

// NewSecurityGroupRule definition POST/PUT request /_group/{groupID}
type NewSecurityGroupRule struct {
	Action       string `json:"action"`
	Direction    string `json:"direction"`
	IPRange      string `json:"ip_range"`
	Protocol     string `json:"protocol"`
	DestPortFrom int    `json:"dest_port_from,omitempty"`
}

// UpdateSecurityGroupRule definition POST/PUT request /_group/{groupID}
type UpdateSecurityGroupRule struct {
	Action       string `json:"action"`
	Direction    string `json:"direction"`
	IPRange      string `json:"ip_range"`
	Protocol     string `json:"protocol"`
	Position     int    `json:"position"`
	DestPortFrom int    `json:"dest_port_from,omitempty"`
}

// GetSecurityGroupRules returns a GroupRules
func (s *API) GetSecurityGroupRules(groupID string) (*GetSecurityGroupRules, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, fmt.Sprintf("security_groups/%s/rules", groupID), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var GroupRules GetSecurityGroupRules

	if err = json.Unmarshal(body, &GroupRules); err != nil {
		return nil, err
	}
	return &GroupRules, nil
}

// GetAGroupRule returns a SecurityGroupRule
func (s *API) GetASecurityGroupRule(groupID string, rulesID string) (*GetSecurityGroupRule, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, fmt.Sprintf("security_groups/%s/rules/%s", groupID, rulesID), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var GroupRules GetSecurityGroupRule

	if err = json.Unmarshal(body, &GroupRules); err != nil {
		return nil, err
	}
	return &GroupRules, nil
}

type postGroupRuleResponse struct {
	SecurityGroupRule SecurityGroupRule `json:"rule"`
}

// PostGroupRule posts a rule on a server
func (s *API) PostSecurityGroupRule(GroupID string, rules NewSecurityGroupRule) (*SecurityGroupRule, error) {
	resp, err := s.PostResponse(s.computeAPI, fmt.Sprintf("security_groups/%s/rules", GroupID), rules)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := s.handleHTTPError([]int{http.StatusCreated}, resp)
	if err != nil {
		return nil, err
	}
	var res postGroupRuleResponse
	err = json.Unmarshal(data, &res)
	return &res.SecurityGroupRule, err
}

// PutGroupRule updates a SecurityGroupRule
func (s *API) PutSecurityGroupRule(rules UpdateSecurityGroupRule, GroupID, RuleID string) error {
	resp, err := s.PutResponse(s.computeAPI, fmt.Sprintf("security_groups/%s/rules/%s", GroupID, RuleID), rules)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = s.handleHTTPError([]int{http.StatusOK}, resp)
	return err
}

// DeleteGroupRule deletes a SecurityGroupRule
func (s *API) DeleteSecurityGroupRule(GroupID, RuleID string) error {
	resp, err := s.DeleteResponse(s.computeAPI, fmt.Sprintf("security_groups/%s/rules/%s", GroupID, RuleID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = s.handleHTTPError([]int{http.StatusNoContent}, resp)
	return err
}
