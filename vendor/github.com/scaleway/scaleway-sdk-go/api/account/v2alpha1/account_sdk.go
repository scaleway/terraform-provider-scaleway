// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package account provides methods and message types of the account v2alpha1 API.
package account

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API: this API allows to manage your scaleway account
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ListSSHKeysRequestOrderBy string

const (
	// ListSSHKeysRequestOrderByCreatedAtAsc is [insert doc].
	ListSSHKeysRequestOrderByCreatedAtAsc = ListSSHKeysRequestOrderBy("created_at_asc")
	// ListSSHKeysRequestOrderByCreatedAtDesc is [insert doc].
	ListSSHKeysRequestOrderByCreatedAtDesc = ListSSHKeysRequestOrderBy("created_at_desc")
	// ListSSHKeysRequestOrderByUpdatedAtAsc is [insert doc].
	ListSSHKeysRequestOrderByUpdatedAtAsc = ListSSHKeysRequestOrderBy("updated_at_asc")
	// ListSSHKeysRequestOrderByUpdatedAtDesc is [insert doc].
	ListSSHKeysRequestOrderByUpdatedAtDesc = ListSSHKeysRequestOrderBy("updated_at_desc")
	// ListSSHKeysRequestOrderByNameAsc is [insert doc].
	ListSSHKeysRequestOrderByNameAsc = ListSSHKeysRequestOrderBy("name_asc")
	// ListSSHKeysRequestOrderByNameDesc is [insert doc].
	ListSSHKeysRequestOrderByNameDesc = ListSSHKeysRequestOrderBy("name_desc")
)

func (enum ListSSHKeysRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListSSHKeysRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListSSHKeysRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListSSHKeysRequestOrderBy(ListSSHKeysRequestOrderBy(tmp).String())
	return nil
}

// ListSSHKeysResponse: list ssh keys response
type ListSSHKeysResponse struct {
	SSHKeys []*SSHKey `json:"ssh_keys"`

	TotalCount uint32 `json:"total_count"`
}

// SSHKey: ssh key
type SSHKey struct {
	ID string `json:"id"`

	Name string `json:"name"`

	PublicKey string `json:"public_key"`

	Fingerprint string `json:"fingerprint"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`

	CreationInfo *SSHKeyCreationInfo `json:"creation_info"`

	OrganizationID string `json:"organization_id"`
}

type SSHKeyCreationInfo struct {
	Address string `json:"address"`

	UserAgent string `json:"user_agent"`

	CountryCode string `json:"country_code"`
}

// Service API

type ListSSHKeysRequest struct {
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListSSHKeysRequestOrderBy `json:"-"`

	Page *int32 `json:"-"`

	PageSize *uint32 `json:"-"`

	Name *string `json:"-"`

	OrganizationID *string `json:"-"`
}

// ListSSHKeys: list all SSH keys
func (s *API) ListSSHKeys(req *ListSSHKeysRequest, opts ...scw.RequestOption) (*ListSSHKeysResponse, error) {
	var err error

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/account/v2alpha1/ssh-keys",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSSHKeysResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSSHKeysResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSSHKeysResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListSSHKeysResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.SSHKeys = append(r.SSHKeys, results.SSHKeys...)
	r.TotalCount += uint32(len(results.SSHKeys))
	return uint32(len(results.SSHKeys)), nil
}

type CreateSSHKeyRequest struct {
	// Name: the name of the SSH key
	Name string `json:"name"`
	// PublicKey: SSH public key. Currently ssh-rsa, ssh-dss (DSA), ssh-ed25519 and ecdsa keys with NIST curves are supported
	PublicKey string `json:"public_key"`
	// OrganizationID: organization owning the resource
	OrganizationID string `json:"organization_id"`
}

// CreateSSHKey: add a SSH key to your Scaleway account
//
// Add a SSH key to your Scaleway account.
func (s *API) CreateSSHKey(req *CreateSSHKeyRequest, opts ...scw.RequestOption) (*SSHKey, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/account/v2alpha1/ssh-keys",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SSHKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetSSHKeyRequest struct {
	// SSHKeyID: the ID of the SSH key
	SSHKeyID string `json:"-"`
}

// GetSSHKey: get SSH key details
func (s *API) GetSSHKey(req *GetSSHKeyRequest, opts ...scw.RequestOption) (*SSHKey, error) {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return nil, errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/account/v2alpha1/ssh-key/" + fmt.Sprint(req.SSHKeyID) + "",
		Headers: http.Header{},
	}

	var resp SSHKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateSSHKeyRequest struct {
	SSHKeyID string `json:"-"`

	Name *string `json:"name"`
}

// UpdateSSHKey: update an SSH key
func (s *API) UpdateSSHKey(req *UpdateSSHKeyRequest, opts ...scw.RequestOption) (*SSHKey, error) {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return nil, errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/account/v2alpha1/ssh-key/" + fmt.Sprint(req.SSHKeyID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SSHKey

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteSSHKeyRequest struct {
	SSHKeyID string `json:"-"`
}

// DeleteSSHKey: remove a SSH key from your Scaleway account
//
// Remove a SSH key from your Scaleway account.
func (s *API) DeleteSSHKey(req *DeleteSSHKeyRequest, opts ...scw.RequestOption) error {
	var err error

	if fmt.Sprint(req.SSHKeyID) == "" {
		return errors.New("field SSHKeyID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/account/v2alpha1/ssh-key/" + fmt.Sprint(req.SSHKeyID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}
