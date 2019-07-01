package instance

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/utils"
)

// UpdateSecurityGroupRequest contains the parameters to update a security group
type UpdateSecurityGroupRequest struct {
	Zone            utils.Zone `json:"-"`
	SecurityGroupID string     `json:"-"`

	Name                  *string              `json:"name,omitempty"`
	Description           *string              `json:"description,omitempty"`
	InboundDefaultPolicy  *SecurityGroupPolicy `json:"inbound_default_policy,omitempty"`
	OutboundDefaultPolicy *SecurityGroupPolicy `json:"outbound_default_policy,omitempty"`
	Stateful              *bool                `json:"stateful,omitempty"`
}

type UpdateSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup
}

// UpdateSecurityGroup updates a security group.
func (s *API) UpdateSecurityGroup(req *UpdateSecurityGroupRequest, opts ...scw.RequestOption) (*UpdateSecurityGroupResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	getSGResponse, err := s.GetSecurityGroup(&GetSecurityGroupRequest{
		Zone:            req.Zone,
		SecurityGroupID: req.SecurityGroupID,
	}, opts...)
	if err != nil {
		return nil, err
	}

	setRequest := &setSecurityGroupRequest{
		ID:                    getSGResponse.SecurityGroup.ID,
		Name:                  getSGResponse.SecurityGroup.Name,
		Description:           getSGResponse.SecurityGroup.Description,
		Organization:          getSGResponse.SecurityGroup.Organization,
		OrganizationDefault:   getSGResponse.SecurityGroup.OrganizationDefault,
		OutboundDefaultPolicy: getSGResponse.SecurityGroup.OutboundDefaultPolicy,
		InboundDefaultPolicy:  getSGResponse.SecurityGroup.InboundDefaultPolicy,
		Stateful:              getSGResponse.SecurityGroup.Stateful,
		Zone:                  req.Zone,
		EnableDefaultSecurity: getSGResponse.SecurityGroup.EnableDefaultSecurity,
		CreationDate:          getSGResponse.SecurityGroup.CreationDate,
		ModificationDate:      getSGResponse.SecurityGroup.ModificationDate,
		Servers:               getSGResponse.SecurityGroup.Servers,
	}

	// Override the values that need to be updated
	if req.Name != nil {
		setRequest.Name = *req.Name
	}
	if req.Description != nil {
		setRequest.Description = *req.Description
	}
	if req.InboundDefaultPolicy != nil {
		setRequest.InboundDefaultPolicy = *req.InboundDefaultPolicy
	}
	if req.OutboundDefaultPolicy != nil {
		setRequest.OutboundDefaultPolicy = *req.OutboundDefaultPolicy
	}
	if req.Stateful != nil {
		setRequest.Stateful = *req.Stateful
	}

	setRes, err := s.setSecurityGroup(setRequest, opts...)
	if err != nil {
		return nil, err
	}

	return &UpdateSecurityGroupResponse{
		SecurityGroup: setRes.SecurityGroup,
	}, nil
}
