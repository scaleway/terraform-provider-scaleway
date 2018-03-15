package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// TokenDefinition represents a  Token
type TokenDefinition struct {
	UserID             string         `json:"user_id"`
	Description        string         `json:"description,omitempty"`
	Roles              RoleDefinition `json:"roles"`
	Expires            string         `json:"expires"`
	InheritsUsersPerms bool           `json:"inherits_user_perms"`
	ID                 string         `json:"id"`
}

// TokensDefinition represents a  Tokens
type TokensDefinition struct {
	Token TokenDefinition `json:"token"`
}

// GetTokens represents a list of  Tokens
type GetTokens struct {
	Tokens []TokenDefinition `json:"tokens"`
}

// RoleDefinition represents a  Token UserId Role
type RoleDefinition struct {
	Organization OrganizationDefinition `json:"organization,omitempty"`
	Role         string                 `json:"role,omitempty"`
}

// UserDefinition represents a  User
type UserDefinition struct {
	Email         string                   `json:"email"`
	Firstname     string                   `json:"firstname"`
	Fullname      string                   `json:"fullname"`
	ID            string                   `json:"id"`
	Lastname      string                   `json:"lastname"`
	Organizations []OrganizationDefinition `json:"organizations"`
	Roles         []RoleDefinition         `json:"roles"`
	SSHPublicKeys []KeyDefinition          `json:"ssh_public_keys"`
}

// KeyDefinition represents a key
type KeyDefinition struct {
	Key         string `json:"key"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// UsersDefinition represents the response of a GET /user
type UsersDefinition struct {
	User UserDefinition `json:"user"`
}

// UserPatchSSHKeyDefinition represents a User Patch
type UserPatchSSHKeyDefinition struct {
	SSHPublicKeys []KeyDefinition `json:"ssh_public_keys"`
}

// PatchUserSSHKey updates a user
func (s *API) PatchUserSSHKey(UserID string, definition UserPatchSSHKeyDefinition) (*UserDefinition, error) {
	resp, err := s.PatchResponse(AccountAPI, fmt.Sprintf("users/%s", UserID), definition)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var user UsersDefinition

	if err = json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	return &user.User, nil
}

// GetUserID returns the userID
func (s *API) GetUserID() (string, error) {
	resp, err := s.GetResponsePaginate(AccountAPI, fmt.Sprintf("tokens/%s", s.Token), url.Values{})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return "", err
	}
	var token TokensDefinition

	if err = json.Unmarshal(body, &token); err != nil {
		return "", err
	}
	return token.Token.UserID, nil
}

// GetUser returns the user
func (s *API) GetUser() (*UserDefinition, error) {
	userID, err := s.GetUserID()
	if err != nil {
		return nil, err
	}
	resp, err := s.GetResponsePaginate(AccountAPI, fmt.Sprintf("users/%s", userID), url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := s.handleHTTPError([]int{http.StatusOK}, resp)
	if err != nil {
		return nil, err
	}
	var user UsersDefinition

	if err = json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	return &user.User, nil
}
