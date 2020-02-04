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

// API this API allows to manage your Bare metal server
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type IPFailoverEventAction string

const (
	// IPFailoverEventActionUnknown is [insert doc].
	IPFailoverEventActionUnknown = IPFailoverEventAction("unknown")
	// IPFailoverEventActionBillingStart is [insert doc].
	IPFailoverEventActionBillingStart = IPFailoverEventAction("billing_start")
	// IPFailoverEventActionBillingStop is [insert doc].
	IPFailoverEventActionBillingStop = IPFailoverEventAction("billing_stop")
	// IPFailoverEventActionOrderFail is [insert doc].
	IPFailoverEventActionOrderFail = IPFailoverEventAction("order_fail")
	// IPFailoverEventActionUpdateIP is [insert doc].
	IPFailoverEventActionUpdateIP = IPFailoverEventAction("update_ip")
	// IPFailoverEventActionUpdateIPFail is [insert doc].
	IPFailoverEventActionUpdateIPFail = IPFailoverEventAction("update_ip_fail")
)

func (enum IPFailoverEventAction) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum IPFailoverEventAction) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *IPFailoverEventAction) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = IPFailoverEventAction(IPFailoverEventAction(tmp).String())
	return nil
}

type IPFailoverMACType string

const (
	// IPFailoverMACTypeUnknownMacType is [insert doc].
	IPFailoverMACTypeUnknownMacType = IPFailoverMACType("unknown_mac_type")
	// IPFailoverMACTypeNone is [insert doc].
	IPFailoverMACTypeNone = IPFailoverMACType("none")
	// IPFailoverMACTypeDuplicate is [insert doc].
	IPFailoverMACTypeDuplicate = IPFailoverMACType("duplicate")
	// IPFailoverMACTypeVmware is [insert doc].
	IPFailoverMACTypeVmware = IPFailoverMACType("vmware")
	// IPFailoverMACTypeXen is [insert doc].
	IPFailoverMACTypeXen = IPFailoverMACType("xen")
	// IPFailoverMACTypeKvm is [insert doc].
	IPFailoverMACTypeKvm = IPFailoverMACType("kvm")
)

func (enum IPFailoverMACType) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_mac_type"
	}
	return string(enum)
}

func (enum IPFailoverMACType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *IPFailoverMACType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = IPFailoverMACType(IPFailoverMACType(tmp).String())
	return nil
}

type IPFailoverStatus string

const (
	// IPFailoverStatusUnknown is [insert doc].
	IPFailoverStatusUnknown = IPFailoverStatus("unknown")
	// IPFailoverStatusDelivering is [insert doc].
	IPFailoverStatusDelivering = IPFailoverStatus("delivering")
	// IPFailoverStatusReady is [insert doc].
	IPFailoverStatusReady = IPFailoverStatus("ready")
	// IPFailoverStatusUpdating is [insert doc].
	IPFailoverStatusUpdating = IPFailoverStatus("updating")
	// IPFailoverStatusError is [insert doc].
	IPFailoverStatusError = IPFailoverStatus("error")
	// IPFailoverStatusDeleting is [insert doc].
	IPFailoverStatusDeleting = IPFailoverStatus("deleting")
	// IPFailoverStatusLocked is [insert doc].
	IPFailoverStatusLocked = IPFailoverStatus("locked")
)

func (enum IPFailoverStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum IPFailoverStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *IPFailoverStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = IPFailoverStatus(IPFailoverStatus(tmp).String())
	return nil
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

type ListIPFailoverEventsRequestOrderBy string

const (
	// ListIPFailoverEventsRequestOrderByCreatedAtAsc is [insert doc].
	ListIPFailoverEventsRequestOrderByCreatedAtAsc = ListIPFailoverEventsRequestOrderBy("created_at_asc")
	// ListIPFailoverEventsRequestOrderByCreatedAtDesc is [insert doc].
	ListIPFailoverEventsRequestOrderByCreatedAtDesc = ListIPFailoverEventsRequestOrderBy("created_at_desc")
)

func (enum ListIPFailoverEventsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListIPFailoverEventsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListIPFailoverEventsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListIPFailoverEventsRequestOrderBy(ListIPFailoverEventsRequestOrderBy(tmp).String())
	return nil
}

type ListIPFailoversRequestOrderBy string

const (
	// ListIPFailoversRequestOrderByCreatedAtAsc is [insert doc].
	ListIPFailoversRequestOrderByCreatedAtAsc = ListIPFailoversRequestOrderBy("created_at_asc")
	// ListIPFailoversRequestOrderByCreatedAtDesc is [insert doc].
	ListIPFailoversRequestOrderByCreatedAtDesc = ListIPFailoversRequestOrderBy("created_at_desc")
)

func (enum ListIPFailoversRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListIPFailoversRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListIPFailoversRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListIPFailoversRequestOrderBy(ListIPFailoversRequestOrderBy(tmp).String())
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

type OfferStock string

const (
	// OfferStockEmpty is [insert doc].
	OfferStockEmpty = OfferStock("empty")
	// OfferStockLow is [insert doc].
	OfferStockLow = OfferStock("low")
	// OfferStockAvailable is [insert doc].
	OfferStockAvailable = OfferStock("available")
)

func (enum OfferStock) String() string {
	if enum == "" {
		// return default value if empty
		return "empty"
	}
	return string(enum)
}

func (enum OfferStock) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *OfferStock) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = OfferStock(OfferStock(tmp).String())
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

// AttachIPFailoversResponse attach ip failovers response
type AttachIPFailoversResponse struct {
	// Failovers list of the attached IP failovers
	Failovers []*IPFailover `json:"failovers"`
}

// CPU cpu
type CPU struct {
	// Name name of the CPU
	Name string `json:"name"`
	// Cores number of cores of the CPU
	Cores uint32 `json:"cores"`
	// Threads number of threads of the CPU
	Threads uint32 `json:"threads"`

	Frequency uint32 `json:"frequency"`
}

// DetachIPFailoversResponse detach ip failovers response
type DetachIPFailoversResponse struct {
	// Failovers list of the detached IP failovers
	Failovers []*IPFailover `json:"failovers"`
}

// Disk disk
type Disk struct {
	// Capacity capacity of the disk in GB
	Capacity uint64 `json:"capacity"`
	// Type type of the disk
	Type string `json:"type"`
}

// GetServerMetricsResponse get server metrics response
type GetServerMetricsResponse struct {
	// Pings timeseries of ping on the server
	Pings *scw.TimeSeries `json:"pings"`
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

// IPFailover ip failover
type IPFailover struct {
	// ID iD of the IP failover
	ID string `json:"id"`
	// OrganizationID organization ID the IP failover is attached to
	OrganizationID string `json:"organization_id"`
	// Description description of the IP failover
	Description string `json:"description"`
	// Tags tags associated to the IP failover
	Tags []string `json:"tags"`
	// UpdatedAt date of last update of the IP failover
	UpdatedAt time.Time `json:"updated_at"`
	// CreatedAt date of creation of the IP failover
	CreatedAt time.Time `json:"created_at"`
	// Status status of the IP failover
	//
	// Default value: unknown
	Status IPFailoverStatus `json:"status"`
	// IPAddress iP of the IP failover
	IPAddress net.IP `json:"ip_address"`
	// MacAddress mac address of the IP failover
	MacAddress string `json:"mac_address"`
	// ServerID serverID linked to the IP failover
	ServerID string `json:"server_id"`
	// MacType type of the MAC generated of the IP failover
	//
	// Default value: unknown_mac_type
	MacType IPFailoverMACType `json:"mac_type"`
	// Reverse reverse IP value
	Reverse string `json:"reverse"`
	// ReverseStatus status of the reverse
	//
	// Default value: unknown
	ReverseStatus IPReverseStatus `json:"reverse_status"`
	// ReverseStatusMessage a message related to the reverse status, in case of an error for example
	ReverseStatusMessage *string `json:"reverse_status_message"`
	// Zone the zone in which is the ip
	Zone scw.Zone `json:"zone"`
}

// IPFailoverEvent ip failover event
type IPFailoverEvent struct {
	// ID iD of the IP failover for whom the action will be applied
	ID string `json:"id"`
	// Action the action that will be applied to the IP failover
	//
	// Default value: unknown
	Action IPFailoverEventAction `json:"action"`
	// UpdatedAt date of last modification of the action
	UpdatedAt time.Time `json:"updated_at"`
	// CreatedAt date of creation of the action
	CreatedAt time.Time `json:"created_at"`
}

// ListIPFailoverEventsResponse list ip failover events response
type ListIPFailoverEventsResponse struct {
	// TotalCount total count of matching IP failover events
	TotalCount uint32 `json:"total_count"`
	// Event iP failover events that match filters
	Event []*IPFailoverEvent `json:"event"`
}

// ListIPFailoversResponse list ip failovers response
type ListIPFailoversResponse struct {
	// TotalCount total count of matching IP failovers
	TotalCount uint32 `json:"total_count"`
	// Failovers listing of failovers
	Failovers []*IPFailover `json:"failovers"`
}

// ListOffersResponse list offers response
type ListOffersResponse struct {
	// TotalCount total count of matching offers
	TotalCount uint32 `json:"total_count"`
	// Offers offers that match filters
	Offers []*Offer `json:"offers"`
}

// ListOsResponse list os response
type ListOsResponse struct {
	// TotalCount total count of matching OS
	TotalCount uint32 `json:"total_count"`
	// Os oS that match filters
	Os []*Os `json:"os"`
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

// Memory memory
type Memory struct {
	Capacity uint64 `json:"capacity"`

	Type string `json:"type"`

	Frequency uint32 `json:"frequency"`

	Ecc bool `json:"ecc"`
}

// Offer offer
type Offer struct {
	// ID iD of the offer
	ID string `json:"id"`
	// Name name of the offer
	Name string `json:"name"`
	// Stock stock level
	//
	// Default value: empty
	Stock OfferStock `json:"stock"`
	// Bandwidth bandwidth available with the offer
	Bandwidth uint32 `json:"bandwidth"`
	// CommercialRange commercial range of the offer
	CommercialRange string `json:"commercial_range"`
	// PriceByMinute price of the offer by minutes, this field is deprecated, please use `price_per_sixty_minutes` instead
	PriceByMinute *scw.Money `json:"price_by_minute"`
	// PriceByMonth price of the offer by months, this field is deprecated, please use `price_per_month` instead
	PriceByMonth *scw.Money `json:"price_by_month"`
	// PricePerSixtyMinutes price of the offer for the next 60 minutes (a server order at 11h32 will be payed until 12h32)
	PricePerSixtyMinutes *scw.Money `json:"price_per_sixty_minutes"`
	// PricePerMonth price of the offer per months
	PricePerMonth *scw.Money `json:"price_per_month"`
	// Disk disks specifications of the offer
	Disk []*Disk `json:"disk"`
	// Enable true if the offer is currently available
	Enable bool `json:"enable"`
	// CPU cPU specifications of the offer
	CPU []*CPU `json:"cpu"`
	// Memory memory specifications of the offer
	Memory []*Memory `json:"memory"`
	// QuotaName name of the quota associated to the offer
	QuotaName string `json:"quota_name"`
}

// Os os
type Os struct {
	// ID iD of the OS
	ID string `json:"id"`
	// Name name of the OS
	Name string `json:"name"`
	// Version version of the OS
	Version string `json:"version"`
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
	// Zone the zone in which is the server
	Zone scw.Zone `json:"zone"`
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
	// OrganizationID filter servers by organization ID
	OrganizationID *string `json:"-"`
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
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

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
func (r *ListServersResponse) UnsafeAppend(res interface{}) (uint32, error) {
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

type GetServerMetricsRequest struct {
	Zone scw.Zone `json:"-"`
	// ServerID server ID to get the metrics
	ServerID string `json:"-"`
}

// GetServerMetrics return server metrics
//
// Give the ping status on the server associated with the given ID.
func (s *API) GetServerMetrics(req *GetServerMetricsRequest, opts ...scw.RequestOption) (*GetServerMetricsResponse, error) {
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
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/metrics",
		Headers: http.Header{},
	}

	var resp GetServerMetricsResponse

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
func (r *ListServerEventsResponse) UnsafeAppend(res interface{}) (uint32, error) {
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

type CreateIPFailoverRequest struct {
	Zone scw.Zone `json:"-"`
	// OrganizationID iD of the organization to associate to the IP failover
	OrganizationID string `json:"organization_id"`
	// Description description to associate to the IP failover, max 255 characters
	Description string `json:"description"`
	// Tags tags to associate to the IP failover
	Tags []string `json:"tags"`
	// MacType mAC type to use for the IP failover
	//
	// Default value: unknown_mac_type
	MacType IPFailoverMACType `json:"mac_type"`
	// DuplicateMacFrom iD of the IP failover which must be duplicate
	DuplicateMacFrom *string `json:"duplicate_mac_from"`
}

// CreateIPFailover create IP failover
//
// Create an IP failover. Once the IP failover is created, you probably want to attach it to a server.
func (s *API) CreateIPFailover(req *CreateIPFailoverRequest, opts ...scw.RequestOption) (*IPFailover, error) {
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
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/ip-failovers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IPFailover

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetIPFailoverRequest struct {
	Zone scw.Zone `json:"-"`
	// IPFailoverID iD of the IP failover
	IPFailoverID string `json:"-"`
}

// GetIPFailover get IP failover
//
// Get the IP failover associated with the given ID.
func (s *API) GetIPFailover(req *GetIPFailoverRequest, opts ...scw.RequestOption) (*IPFailover, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPFailoverID) == "" {
		return nil, errors.New("field IPFailoverID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/ip-failovers/" + fmt.Sprint(req.IPFailoverID) + "",
		Headers: http.Header{},
	}

	var resp IPFailover

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListIPFailoversRequest struct {
	Zone scw.Zone `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize number of IP failover per page
	PageSize *uint32 `json:"-"`
	// OrderBy order of the IP failovers
	//
	// Default value: created_at_asc
	OrderBy ListIPFailoversRequestOrderBy `json:"-"`
	// Tags filter IP failovers by tags
	Tags []string `json:"-"`
	// Status filter IP failovers by status
	Status []string `json:"-"`
	// ServerIDs filter IP failovers by server IDs
	ServerIDs []string `json:"-"`
	// OrganizationID filter servers by organization ID
	OrganizationID *string `json:"-"`
}

// ListIPFailovers list IP failovers
//
// List all created IP failovers.
func (s *API) ListIPFailovers(req *ListIPFailoversRequest, opts ...scw.RequestOption) (*ListIPFailoversResponse, error) {
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
	parameter.AddToQuery(query, "server_ids", req.ServerIDs)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/ip-failovers",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListIPFailoversResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListIPFailoversResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListIPFailoversResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListIPFailoversResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Failovers = append(r.Failovers, results.Failovers...)
	r.TotalCount += uint32(len(results.Failovers))
	return uint32(len(results.Failovers)), nil
}

type DeleteIPFailoverRequest struct {
	Zone scw.Zone `json:"-"`
	// IPFailoverID iD of the IP failover to delete
	IPFailoverID string `json:"-"`
}

// DeleteIPFailover delete IP failover
//
// Delete the IP failover associated with the given IP.
func (s *API) DeleteIPFailover(req *DeleteIPFailoverRequest, opts ...scw.RequestOption) (*IPFailover, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPFailoverID) == "" {
		return nil, errors.New("field IPFailoverID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/ip-failovers/" + fmt.Sprint(req.IPFailoverID) + "",
		Headers: http.Header{},
	}

	var resp IPFailover

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateIPFailoverRequest struct {
	Zone scw.Zone `json:"-"`
	// IPFailoverID iD of the IP failover to update
	IPFailoverID string `json:"-"`
	// Description description to associate to the IP failover, max 255 characters, not updated if null
	Description *string `json:"description"`
	// Tags tags to associate to the IP failover, not updated if null
	Tags *[]string `json:"tags"`
	// MacType mAC type to use for the IP failover, not updated if null
	//
	// Default value: unknown_mac_type
	MacType IPFailoverMACType `json:"mac_type"`
	// DuplicateMacFrom iD of the IP failover which must be duplicate, not updated if null
	DuplicateMacFrom *string `json:"duplicate_mac_from"`
	// Reverse new reverse IP to update, not updated if null
	Reverse *string `json:"reverse"`
}

// UpdateIPFailover update IP failover
//
// Update the IP failover associated with the given IP.
func (s *API) UpdateIPFailover(req *UpdateIPFailoverRequest, opts ...scw.RequestOption) (*IPFailover, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPFailoverID) == "" {
		return nil, errors.New("field IPFailoverID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/ip-failovers/" + fmt.Sprint(req.IPFailoverID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp IPFailover

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type AttachIPFailoversRequest struct {
	Zone scw.Zone `json:"-"`
	// IPFailoverIDs iP failover IDs to attach to the server
	IPFailoverIDs []string `json:"ip_failover_ids"`
	// ServerID iD of the server to attach to the IP failovers
	ServerID string `json:"server_id"`
}

// AttachIPFailovers attach IP failovers
//
// Attach IP failovers to the given server ID.
func (s *API) AttachIPFailovers(req *AttachIPFailoversRequest, opts ...scw.RequestOption) (*AttachIPFailoversResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/ip-failovers/attach",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp AttachIPFailoversResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DetachIPFailoversRequest struct {
	Zone scw.Zone `json:"-"`
	// IPFailoverIDs iP failover IDs to detach to the server
	IPFailoverIDs []string `json:"ip_failover_ids"`
}

// DetachIPFailovers detach IP failovers
//
// Detach IP failovers to the given server ID.
func (s *API) DetachIPFailovers(req *DetachIPFailoversRequest, opts ...scw.RequestOption) (*DetachIPFailoversResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/ip-failovers/detach",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp DetachIPFailoversResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListIPFailoverEventsRequest struct {
	Zone scw.Zone `json:"-"`
	// IPFailoverID iD of the IP failover events searched
	IPFailoverID string `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize number of IP failover events per page
	PageSize *uint32 `json:"-"`
	// OrderBy order of the IP failover events
	//
	// Default value: created_at_asc
	OrderBy ListIPFailoverEventsRequestOrderBy `json:"-"`
}

// ListIPFailoverEvents list IP failover events
//
// List IP failover events associated with the given ID.
func (s *API) ListIPFailoverEvents(req *ListIPFailoverEventsRequest, opts ...scw.RequestOption) (*ListIPFailoverEventsResponse, error) {
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

	if fmt.Sprint(req.IPFailoverID) == "" {
		return nil, errors.New("field IPFailoverID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/ip-failovers/" + fmt.Sprint(req.IPFailoverID) + "/events",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListIPFailoverEventsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListIPFailoverEventsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListIPFailoverEventsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListIPFailoverEventsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Event = append(r.Event, results.Event...)
	r.TotalCount += uint32(len(results.Event))
	return uint32(len(results.Event)), nil
}

type ListOffersRequest struct {
	Zone scw.Zone `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize number of offers per page
	PageSize *uint32 `json:"-"`
}

// ListOffers list offers
//
// List all available server offers.
func (s *API) ListOffers(req *ListOffersRequest, opts ...scw.RequestOption) (*ListOffersResponse, error) {
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

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/offers",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListOffersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListOffersResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListOffersResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListOffersResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Offers = append(r.Offers, results.Offers...)
	r.TotalCount += uint32(len(results.Offers))
	return uint32(len(results.Offers)), nil
}

type GetOfferRequest struct {
	Zone scw.Zone `json:"-"`
	// OfferID iD of the researched Offer
	OfferID string `json:"-"`
}

// GetOffer get offer
//
// Return specific offer for the given ID.
func (s *API) GetOffer(req *GetOfferRequest, opts ...scw.RequestOption) (*Offer, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.OfferID) == "" {
		return nil, errors.New("field OfferID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/offers/" + fmt.Sprint(req.OfferID) + "",
		Headers: http.Header{},
	}

	var resp Offer

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListOsRequest struct {
	Zone scw.Zone `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize number of OS per page
	PageSize *uint32 `json:"-"`
}

// ListOs list OS
//
// List all available OS that can be install on a baremetal server.
func (s *API) ListOs(req *ListOsRequest, opts ...scw.RequestOption) (*ListOsResponse, error) {
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

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/os",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListOsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListOsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListOsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListOsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Os = append(r.Os, results.Os...)
	r.TotalCount += uint32(len(results.Os))
	return uint32(len(results.Os)), nil
}

type GetOsRequest struct {
	Zone scw.Zone `json:"-"`
	// OsID iD of the researched OS
	OsID string `json:"-"`
}

// GetOs get OS
//
// Return specific OS for the given ID.
func (s *API) GetOs(req *GetOsRequest, opts ...scw.RequestOption) (*Os, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.OsID) == "" {
		return nil, errors.New("field OsID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/baremetal/v1alpha1/zones/" + fmt.Sprint(req.Zone) + "/os/" + fmt.Sprint(req.OsID) + "",
		Headers: http.Header{},
	}

	var resp Os

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
