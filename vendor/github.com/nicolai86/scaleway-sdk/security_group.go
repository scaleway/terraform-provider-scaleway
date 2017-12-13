package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SecurityGroups definition
type SecurityGroups struct {
	Description           string          `json:"description"`
	ID                    string          `json:"id"`
	Organization          string          `json:"organization"`
	Name                  string          `json:"name"`
	Servers               []SecurityGroup `json:"servers"`
	EnableDefaultSecurity bool            `json:"enable_default_security"`
	OrganizationDefault   bool            `json:"organization_default"`
}

// GetSecurityGroups represents the response of a GET /security_groups/
type GetSecurityGroups struct {
	SecurityGroups []SecurityGroups `json:"security_groups"`
}

// GetSecurityGroup represents the response of a GET /security_groups/{groupID}
type GetSecurityGroup struct {
	SecurityGroups SecurityGroups `json:"security_group"`
}

// SecurityGroup represents a  security group
type SecurityGroup struct {
	// Identifier is a unique identifier for the security group
	Identifier string `json:"id,omitempty"`

	// Name is the user-defined name of the security group
	Name string `json:"name,omitempty"`
}

// NewSecurityGroup definition POST request /security_groups
type NewSecurityGroup struct {
	Organization string `json:"organization"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}

// UpdateSecurityGroup definition PUT request /security_groups
type UpdateSecurityGroup struct {
	Organization        string `json:"organization"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	OrganizationDefault bool   `json:"organization_default"`
}

// DeleteSecurityGroup deletes a SecurityGroup
func (s *API) DeleteSecurityGroup(securityGroupID string) error {
	resp, err := s.DeleteResponse(s.computeAPI, fmt.Sprintf("security_groups/%s", securityGroupID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = s.handleHTTPError([]int{http.StatusNoContent}, resp)
	return err
}

// PutSecurityGroup updates a SecurityGroup
func (s *API) PutSecurityGroup(group UpdateSecurityGroup, securityGroupID string) error {
	resp, err := s.PutResponse(s.computeAPI, fmt.Sprintf("security_groups/%s", securityGroupID), group)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = s.handleHTTPError([]int{http.StatusOK}, resp)
	return err
}

// GetASecurityGroup returns a SecurityGroup
func (s *API) GetASecurityGroup(groupsID string) (*GetSecurityGroup, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, fmt.Sprintf("security_groups/%s", groupsID), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var securityGroups GetSecurityGroup

	if err = json.Unmarshal(body, &securityGroups); err != nil {
		return nil, err
	}
	return &securityGroups, nil
}

// PostSecurityGroup posts a group on a server
func (s *API) PostSecurityGroup(group NewSecurityGroup) error {
	resp, err := s.PostResponse(s.computeAPI, "security_groups", group)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = s.handleHTTPError([]int{http.StatusCreated}, resp)
	return err
}

// GetSecurityGroups returns a SecurityGroups
func (s *API) GetSecurityGroups() (*GetSecurityGroups, error) {
	resp, err := s.GetResponsePaginate(s.computeAPI, "security_groups", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var securityGroups GetSecurityGroups

	if err = json.Unmarshal(body, &securityGroups); err != nil {
		return nil, err
	}
	return &securityGroups, nil
}
