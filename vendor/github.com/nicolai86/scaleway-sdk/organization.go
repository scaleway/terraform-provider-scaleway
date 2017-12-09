package api

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// OrganizationDefinition represents a  Organization
type OrganizationDefinition struct {
	ID    string           `json:"id"`
	Name  string           `json:"name"`
	Users []UserDefinition `json:"users"`
}

// OrganizationsDefinition represents a  Organizations
type OrganizationsDefinition struct {
	Organizations []OrganizationDefinition `json:"organizations"`
}

// GetOrganization returns Organization
func (s *API) GetOrganization() (*OrganizationsDefinition, error) {
	resp, err := s.GetResponsePaginate(AccountAPI, "organizations", url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var data OrganizationsDefinition

	if err = json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
