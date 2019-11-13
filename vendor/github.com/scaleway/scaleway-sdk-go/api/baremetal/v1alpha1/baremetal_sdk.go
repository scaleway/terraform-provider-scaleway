// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package baremetal provides methods and message types of the baremetal v1alpha1 API.
package baremetal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
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

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API this API allows to manage your Bare metal server.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type IPReverseStatus string

const (
	// IPReverseStatusUnknown is [insert doc].
	IPReverseStatusUnknown = IPReverseStatus("unknown")
	// IPReverseStatusPending is [insert doc].
	IPReverseStatusPending = IPReverseStatus("pending")
	// IPReverseStatusActive is [insert doc].
	IPReverseStatusActive = IPReverseStatus("active")
	// IPReverseStatusError is [insert doc].
	IPReverseStatusError = IPReverseStatus("error")
)

func (enum IPReverseStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum IPReverseStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *IPReverseStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = IPReverseStatus(IPReverseStatus(tmp).String())
	return nil
}

type IPVersion string

const (
	// IPVersionIPv4 is [insert doc].
	IPVersionIPv4 = IPVersion("Ipv4")
	// IPVersionIPv6 is [insert doc].
	IPVersionIPv6 = IPVersion("Ipv6")
)

func (enum IPVersion) String() string {
	if enum == "" {
		// return default value if empty
		return "Ipv4"
	}
	return string(enum)
}

func (enum IPVersion) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *IPVersion) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = IPVersion(IPVersion(tmp).String())
	return nil
}

type ListServerEventsRequestOrderBy string

const (
	// ListServerEventsRequestOrderByCreatedAtAsc is [insert doc].
	ListServerEventsRequestOrderByCreatedAtAsc = ListServerEventsRequestOrderBy("created_at_asc")
	// ListServerEventsRequestOrderByCreatedAtDesc is [insert doc].
	ListServerEventsRequestOrderByCreatedAtDesc = ListServerEventsRequestOrderBy("created_at_desc")
)

func (enum ListServerEventsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListServerEventsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListServerEventsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListServerEventsRequestOrderBy(ListServerEventsRequestOrderBy(tmp).String())
	return nil
}

type ListServersRequestOrderBy string

const (
	// ListServersRequestOrderByCreatedAtAsc is [insert doc].
	ListServersRequestOrderByCreatedAtAsc = ListServersRequestOrderBy("created_at_asc")
	// ListServersRequestOrderByCreatedAtDesc is [insert doc].
	ListServersRequestOrderByCreatedAtDesc = ListServersRequestOrderBy("created_at_desc")
)

func (enum ListServersRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListServersRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListServersRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListServersRequestOrderBy(ListServersRequestOrderBy(tmp).String())
	return nil
}

type RebootServerRequestBootType string

const (
	// RebootServerRequestBootTypeNormal is [insert doc].
	RebootServerRequestBootTypeNormal = RebootServerRequestBootType("normal")
	// RebootServerRequestBootTypeRescue is [insert doc].
	RebootServerRequestBootTypeRescue = RebootServerRequestBootType("rescue")
)

func (enum RebootServerRequestBootType) String() string {
	if enum == "" {
		// return default value if empty
		return "normal"
	}
	return string(enum)
}

func (enum RebootServerRequestBootType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *RebootServerRequestBootType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = RebootServerRequestBootType(RebootServerRequestBootType(tmp).String())
	return nil
}

type ServerBootType string

const (
	// ServerBootTypeNormal is [insert doc].
	ServerBootTypeNormal = ServerBootType("normal")
	// ServerBootTypeRescue is [insert doc].
	ServerBootTypeRescue = ServerBootType("rescue")
)

func (enum ServerBootType) String() string {
	if enum == "" {
		// return default value if empty
		return "normal"
	}
	return string(enum)
}

func (enum ServerBootType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerBootType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerBootType(ServerBootType(tmp).String())
	return nil
}

type ServerInstallStatus string

const (
	// ServerInstallStatusUnknown is [insert doc].
	ServerInstallStatusUnknown = ServerInstallStatus("unknown")
	// ServerInstallStatusCompleted is [insert doc].
	ServerInstallStatusCompleted = ServerInstallStatus("completed")
	// ServerInstallStatusInstalling is [insert doc].
	ServerInstallStatusInstalling = ServerInstallStatus("installing")
	// ServerInstallStatusToInstall is [insert doc].
	ServerInstallStatusToInstall = ServerInstallStatus("to_install")
	// ServerInstallStatusError is [insert doc].
	ServerInstallStatusError = ServerInstallStatus("error")
)

func (enum ServerInstallStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum ServerInstallStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerInstallStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerInstallStatus(ServerInstallStatus(tmp).String())
	return nil
}

type ServerStatus string

const (
	// ServerStatusUnknown is [insert doc].
	ServerStatusUnknown = ServerStatus("unknown")
	// ServerStatusUndelivered is [insert doc].
	ServerStatusUndelivered = ServerStatus("undelivered")
	// ServerStatusReady is [insert doc].
	ServerStatusReady = ServerStatus("ready")
	// ServerStatusStopping is [insert doc].
	ServerStatusStopping = ServerStatus("stopping")
	// ServerStatusStopped is [insert doc].
	ServerStatusStopped = ServerStatus("stopped")
	// ServerStatusStarting is [insert doc].
	ServerStatusStarting = ServerStatus("starting")
	// ServerStatusError is [insert doc].
	ServerStatusError = ServerStatus("error")
	// ServerStatusDeleting is [insert doc].
	ServerStatusDeleting = ServerStatus("deleting")
	// ServerStatusLocked is [insert doc].
	ServerStatusLocked = ServerStatus("locked")
)

func (enum ServerStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum ServerStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ServerStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ServerStatus(ServerStatus(tmp).String())
	return nil
}

// IP ip
type IP struct {
	// ID iD of the IP
	ID string `json:"id"`
	// Address address of the IP
	Address net.IP `json:"address"`
	// Reverse reverse IP value
	Reverse string `json:"reverse"`
	// Version version of IP (v4 or v6)
	//
	// Default value: Ipv4
	Version IPVersion `json:"version"`
	// ReverseStatus status of the reverse
	//
	// Default value: unknown
	ReverseStatus IPReverseStatus `json:"reverse_status"`
	// ReverseStatusMessage a message related to the reverse status, in case of an error for example
	ReverseStatusMessage *string `json:"reverse_status_message"`
}

// ListServerEventsResponse list server events response
type ListServerEventsResponse struct {
	// TotalCount total count of matching events
	TotalCount uint32 `json:"total_count"`
	// Event server events that match filters
	Event []*ServerEvent `json:"event"`
}

// ListServersResponse list servers response
type ListServersResponse struct {
	// TotalCount total count of matching servers
	TotalCount uint32 `json:"total_count"`
	// Servers servers that match filters
	Servers []*Server `json:"servers"`
}

// RemoteServerAccess remote server access
type RemoteServerAccess struct {
	// URL uRL to access to the server console
	URL string `json:"url"`
	// Login the login to use for the remote access authentification
	Login string `json:"login"`
	// Password the password to use for the remote access authentification
	Password string `json:"password"`
	// ExpiresAt the date after which the remote access will be closed
	ExpiresAt time.Time `json:"expires_at"`
}

// Server server
type Server struct {
	// ID iD of the server
	ID string `json:"id"`
	// OrganizationID organization ID the server is attached to
	OrganizationID string `json:"organization_id"`
	// Name name of the server
	Name string `json:"name"`
	// Description description of the server
	Description string `json:"description"`
	// UpdatedAt date of last modification of the server
	UpdatedAt time.Time `json:"updated_at"`
	// CreatedAt date of creation of the server
	CreatedAt time.Time `json:"created_at"`
	// Status status of the server
	//
	// Default value: unknown
	Status ServerStatus `json:"status"`
	// OfferID offer ID of the server
	OfferID string `json:"offer_id"`
	// Install information about the last installation of the server
	Install *ServerInstall `json:"install"`
	// Tags array of customs tags attached to the server
	Tags []string `json:"tags"`
	// IPs array of IPs attached to the server
	IPs []*IP `json:"ips"`
	// Domain domain of the server
	Domain string `json:"domain"`
	// BootType boot type of the server
	//
	// Default value: normal
	BootType ServerBootType `json:"boot_type"`
}

// ServerEvent server event
type ServerEvent struct {
	// ID iD of the server for whom the action will be applied
	ID string `json:"id"`
	// Action the action that will be applied to the server
	Action string `json:"action"`
	// UpdatedAt date of last modification of the action
	UpdatedAt time.Time `json:"updated_at"`
	// CreatedAt date of creation of the action
	CreatedAt time.Time `json:"created_at"`
}

// ServerInstall server install
type ServerInstall struct {
	// OsID iD of the OS
	OsID string `json:"os_id"`
	// Hostname host defined in the server install
	Hostname string `json:"hostname"`
	// SSHKeyIDs sSH public key IDs defined in the server install
	SSHKeyIDs []string `json:"ssh_key_ids"`
	// Status status of the server install
	//
	// Default value: unknown
	Status ServerInstallStatus `json:"status"`
}

// Service API

type ListServersRequest struct {
	Zone scw.Zone `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize number of server per page
	PageSize *uint32 `json:"-"`
	// OrderBy order of the servers
	//
	// Default value: created_at_asc
	OrderBy ListServersRequestOrderBy `json:"-"`
	// Tags filter servers by tags
	Tags []string `json:"-"`
	// Status filter servers by status
	Status []string `json:"-"`
	// Name filter servers by name
	Name *string `json:"-"`
}

// ListServers list servers
//
// List all created servers.
func (s *API) ListServers(req *ListServersRequest, opts ...scw.RequestOption) (*ListServersResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "tags", req.Tags)
	parameter.AddToQuery(query, "status", req.Status)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListServersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListServersResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListServersResponse) UnsafeAppend(res interface{}) (uint32, scw.SdkError) {
	results, ok := res.(*ListServersResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Servers = append(r.Servers, results.Servers...)
	r.TotalCount += uint32(len(results.Servers))
	return uint32(len(results.Servers)), nil
}

type GetServerRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server
	ServerID string `json:"-"`
}

// GetServer get server
//
// Get the server associated with the given ID.
func (s *API) GetServer(req *GetServerRequest, opts ...scw.RequestOption) (*Server, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
		Headers: http.Header{},
	}

	var resp Server

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateServerRequest struct {
	Zone scw.Zone `json:"-"`
	// OfferID offer ID of the new server
	OfferID string `json:"offer_id"`
	// OrganizationID organization ID with which the server will be created
	OrganizationID string `json:"organization_id"`
	// Name name of the server (≠hostname)
	Name string `json:"name"`
	// Description description associated to the server, max 255 characters
	Description string `json:"description"`
	// Tags tags to associate to the server
	Tags []string `json:"tags"`
}

// CreateServer create server
//
// Create a new server. Once the server is created, you probably want to install an OS.
func (s *API) CreateServer(req *CreateServerRequest, opts ...scw.RequestOption) (*Server, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Server

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateServerRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server to update
	ServerID string `json:"-"`
	// Name name of the server (≠hostname), not updated if null
	Name *string `json:"name"`
	// Description description associated to the server, max 255 characters, not updated if null
	Description *string `json:"description"`
	// Tags tags associated to the server, not updated if null
	Tags *[]string `json:"tags"`
}

// UpdateServer update server
//
// Update the server associated with the given ID.
func (s *API) UpdateServer(req *UpdateServerRequest, opts ...scw.RequestOption) (*Server, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Server

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type InstallServerRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID server ID to install
	ServerID string `json:"-"`
	// OsID iD of the OS to install on the server
	OsID string `json:"os_id"`
	// Hostname hostname of the server
	Hostname string `json:"hostname"`
	// SSHKeyIDs sSH key IDs authorized on the server
	SSHKeyIDs []string `json:"ssh_key_ids"`
}

// InstallServer install server
//
// Install an OS on the server associated with the given ID.
func (s *API) InstallServer(req *InstallServerRequest, opts ...scw.RequestOption) (*Server, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/install",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Server

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteServerRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server to delete
	ServerID string `json:"-"`
}

// DeleteServer delete server
//
// Delete the server associated with the given ID.
func (s *API) DeleteServer(req *DeleteServerRequest, opts ...scw.RequestOption) (*Server, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
		Headers: http.Header{},
	}

	var resp Server

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RebootServerRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server to reboot
	ServerID string `json:"-"`
	// BootType the type of boot
	//
	// Default value: normal
	BootType RebootServerRequestBootType `json:"boot_type"`
}

// RebootServer reboot server
//
// Reboot the server associated with the given ID, use boot param to reboot in rescue.
func (s *API) RebootServer(req *RebootServerRequest, opts ...scw.RequestOption) (*Server, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/reboot",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Server

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type StartServerRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server to start
	ServerID string `json:"-"`
}

// StartServer start server
//
// Start the server associated with the given ID.
func (s *API) StartServer(req *StartServerRequest, opts ...scw.RequestOption) (*Server, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/start",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Server

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type StopServerRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server to stop
	ServerID string `json:"-"`
}

// StopServer stop server
//
// Stop the server associated with the given ID.
func (s *API) StopServer(req *StopServerRequest, opts ...scw.RequestOption) (*Server, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/stop",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Server

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListServerEventsRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server events searched
	ServerID string `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize number of server events per page
	PageSize *uint32 `json:"-"`
	// OrderBy order of the server events
	//
	// Default value: created_at_asc
	OrderBy ListServerEventsRequestOrderBy `json:"-"`
}

// ListServerEvents list server events
//
// List events associated to the given server ID.
func (s *API) ListServerEvents(req *ListServerEventsRequest, opts ...scw.RequestOption) (*ListServerEventsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/events",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListServerEventsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListServerEventsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListServerEventsResponse) UnsafeAppend(res interface{}) (uint32, scw.SdkError) {
	results, ok := res.(*ListServerEventsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Event = append(r.Event, results.Event...)
	r.TotalCount += uint32(len(results.Event))
	return uint32(len(results.Event)), nil
}

type CreateRemoteServerAccessRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server
	ServerID string `json:"-"`
	// IP the IP authorized to connect to the given server
	IP string `json:"ip"`
}

// CreateRemoteServerAccess create remote server access
//
// Create remote server access associated with the given ID.
// The remote access is available one hour after the installation of the server.
//
func (s *API) CreateRemoteServerAccess(req *CreateRemoteServerAccessRequest, opts ...scw.RequestOption) (*RemoteServerAccess, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/remote-access",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp RemoteServerAccess

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetRemoteServerAccessRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server
	ServerID string `json:"-"`
}

// GetRemoteServerAccess get remote server access
//
// Get the remote server access associated with the given ID.
func (s *API) GetRemoteServerAccess(req *GetRemoteServerAccessRequest, opts ...scw.RequestOption) (*RemoteServerAccess, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/remote-access",
		Headers: http.Header{},
	}

	var resp RemoteServerAccess

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteRemoteServerAccessRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server
	ServerID string `json:"-"`
}

// DeleteRemoteServerAccess delete remote server access
//
// Delete remote server access associated with the given ID.
func (s *API) DeleteRemoteServerAccess(req *DeleteRemoteServerAccessRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return errors.New("field ServerID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/remote-access",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type UpdateIPRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID iD of the server
	ServerID string `json:"-"`
	// IPID iD of the IP to update
	IPID string `json:"-"`
	// Reverse new reverse IP to update, not updated if null
	Reverse *string `json:"reverse"`
}

// UpdateIP update IP
//
// Configure ip associated with the given server ID and ipID. You can use this method to set a reverse dns for an IP.
func (s *API) UpdateIP(req *UpdateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ServerID) == "" {
		return nil, errors.New("field ServerID cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
