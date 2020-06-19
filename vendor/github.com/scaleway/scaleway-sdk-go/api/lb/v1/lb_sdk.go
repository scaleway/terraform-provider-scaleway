// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package lb provides methods and message types of the lb v1 API.
package lb

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

// API: this API allows you to manage your Load Balancer service
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ACLActionType string

const (
	// ACLActionTypeAllow is [insert doc].
	ACLActionTypeAllow = ACLActionType("allow")
	// ACLActionTypeDeny is [insert doc].
	ACLActionTypeDeny = ACLActionType("deny")
)

func (enum ACLActionType) String() string {
	if enum == "" {
		// return default value if empty
		return "allow"
	}
	return string(enum)
}

func (enum ACLActionType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ACLActionType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ACLActionType(ACLActionType(tmp).String())
	return nil
}

type ACLHTTPFilter string

const (
	// ACLHTTPFilterACLHTTPFilterNone is [insert doc].
	ACLHTTPFilterACLHTTPFilterNone = ACLHTTPFilter("acl_http_filter_none")
	// ACLHTTPFilterPathBegin is [insert doc].
	ACLHTTPFilterPathBegin = ACLHTTPFilter("path_begin")
	// ACLHTTPFilterPathEnd is [insert doc].
	ACLHTTPFilterPathEnd = ACLHTTPFilter("path_end")
	// ACLHTTPFilterRegex is [insert doc].
	ACLHTTPFilterRegex = ACLHTTPFilter("regex")
)

func (enum ACLHTTPFilter) String() string {
	if enum == "" {
		// return default value if empty
		return "acl_http_filter_none"
	}
	return string(enum)
}

func (enum ACLHTTPFilter) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ACLHTTPFilter) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ACLHTTPFilter(ACLHTTPFilter(tmp).String())
	return nil
}

type BackendServerStatsHealthCheckStatus string

const (
	// BackendServerStatsHealthCheckStatusUnknown is [insert doc].
	BackendServerStatsHealthCheckStatusUnknown = BackendServerStatsHealthCheckStatus("unknown")
	// BackendServerStatsHealthCheckStatusNeutral is [insert doc].
	BackendServerStatsHealthCheckStatusNeutral = BackendServerStatsHealthCheckStatus("neutral")
	// BackendServerStatsHealthCheckStatusFailed is [insert doc].
	BackendServerStatsHealthCheckStatusFailed = BackendServerStatsHealthCheckStatus("failed")
	// BackendServerStatsHealthCheckStatusPassed is [insert doc].
	BackendServerStatsHealthCheckStatusPassed = BackendServerStatsHealthCheckStatus("passed")
	// BackendServerStatsHealthCheckStatusCondpass is [insert doc].
	BackendServerStatsHealthCheckStatusCondpass = BackendServerStatsHealthCheckStatus("condpass")
)

func (enum BackendServerStatsHealthCheckStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum BackendServerStatsHealthCheckStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *BackendServerStatsHealthCheckStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = BackendServerStatsHealthCheckStatus(BackendServerStatsHealthCheckStatus(tmp).String())
	return nil
}

type BackendServerStatsServerState string

const (
	// BackendServerStatsServerStateStopped is [insert doc].
	BackendServerStatsServerStateStopped = BackendServerStatsServerState("stopped")
	// BackendServerStatsServerStateStarting is [insert doc].
	BackendServerStatsServerStateStarting = BackendServerStatsServerState("starting")
	// BackendServerStatsServerStateRunning is [insert doc].
	BackendServerStatsServerStateRunning = BackendServerStatsServerState("running")
	// BackendServerStatsServerStateStopping is [insert doc].
	BackendServerStatsServerStateStopping = BackendServerStatsServerState("stopping")
)

func (enum BackendServerStatsServerState) String() string {
	if enum == "" {
		// return default value if empty
		return "stopped"
	}
	return string(enum)
}

func (enum BackendServerStatsServerState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *BackendServerStatsServerState) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = BackendServerStatsServerState(BackendServerStatsServerState(tmp).String())
	return nil
}

type CertificateStatus string

const (
	// CertificateStatusPending is [insert doc].
	CertificateStatusPending = CertificateStatus("pending")
	// CertificateStatusReady is [insert doc].
	CertificateStatusReady = CertificateStatus("ready")
	// CertificateStatusError is [insert doc].
	CertificateStatusError = CertificateStatus("error")
)

func (enum CertificateStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "pending"
	}
	return string(enum)
}

func (enum CertificateStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *CertificateStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = CertificateStatus(CertificateStatus(tmp).String())
	return nil
}

type CertificateType string

const (
	// CertificateTypeLetsencryt is [insert doc].
	CertificateTypeLetsencryt = CertificateType("letsencryt")
	// CertificateTypeCustom is [insert doc].
	CertificateTypeCustom = CertificateType("custom")
)

func (enum CertificateType) String() string {
	if enum == "" {
		// return default value if empty
		return "letsencryt"
	}
	return string(enum)
}

func (enum CertificateType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *CertificateType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = CertificateType(CertificateType(tmp).String())
	return nil
}

type ForwardPortAlgorithm string

const (
	// ForwardPortAlgorithmRoundrobin is [insert doc].
	ForwardPortAlgorithmRoundrobin = ForwardPortAlgorithm("roundrobin")
	// ForwardPortAlgorithmLeastconn is [insert doc].
	ForwardPortAlgorithmLeastconn = ForwardPortAlgorithm("leastconn")
)

func (enum ForwardPortAlgorithm) String() string {
	if enum == "" {
		// return default value if empty
		return "roundrobin"
	}
	return string(enum)
}

func (enum ForwardPortAlgorithm) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ForwardPortAlgorithm) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ForwardPortAlgorithm(ForwardPortAlgorithm(tmp).String())
	return nil
}

type InstanceStatus string

const (
	// InstanceStatusUnknown is [insert doc].
	InstanceStatusUnknown = InstanceStatus("unknown")
	// InstanceStatusReady is [insert doc].
	InstanceStatusReady = InstanceStatus("ready")
	// InstanceStatusPending is [insert doc].
	InstanceStatusPending = InstanceStatus("pending")
	// InstanceStatusStopped is [insert doc].
	InstanceStatusStopped = InstanceStatus("stopped")
	// InstanceStatusError is [insert doc].
	InstanceStatusError = InstanceStatus("error")
	// InstanceStatusLocked is [insert doc].
	InstanceStatusLocked = InstanceStatus("locked")
	// InstanceStatusMigrating is [insert doc].
	InstanceStatusMigrating = InstanceStatus("migrating")
)

func (enum InstanceStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum InstanceStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *InstanceStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = InstanceStatus(InstanceStatus(tmp).String())
	return nil
}

type LbStatus string

const (
	// LbStatusUnknown is [insert doc].
	LbStatusUnknown = LbStatus("unknown")
	// LbStatusReady is [insert doc].
	LbStatusReady = LbStatus("ready")
	// LbStatusPending is [insert doc].
	LbStatusPending = LbStatus("pending")
	// LbStatusStopped is [insert doc].
	LbStatusStopped = LbStatus("stopped")
	// LbStatusError is [insert doc].
	LbStatusError = LbStatus("error")
	// LbStatusLocked is [insert doc].
	LbStatusLocked = LbStatus("locked")
	// LbStatusMigrating is [insert doc].
	LbStatusMigrating = LbStatus("migrating")
)

func (enum LbStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum LbStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LbStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LbStatus(LbStatus(tmp).String())
	return nil
}

type LbTypeStock string

const (
	// LbTypeStockUnknown is [insert doc].
	LbTypeStockUnknown = LbTypeStock("unknown")
	// LbTypeStockLowStock is [insert doc].
	LbTypeStockLowStock = LbTypeStock("low_stock")
	// LbTypeStockOutOfStock is [insert doc].
	LbTypeStockOutOfStock = LbTypeStock("out_of_stock")
	// LbTypeStockAvailable is [insert doc].
	LbTypeStockAvailable = LbTypeStock("available")
)

func (enum LbTypeStock) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum LbTypeStock) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *LbTypeStock) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = LbTypeStock(LbTypeStock(tmp).String())
	return nil
}

type ListACLRequestOrderBy string

const (
	// ListACLRequestOrderByCreatedAtAsc is [insert doc].
	ListACLRequestOrderByCreatedAtAsc = ListACLRequestOrderBy("created_at_asc")
	// ListACLRequestOrderByCreatedAtDesc is [insert doc].
	ListACLRequestOrderByCreatedAtDesc = ListACLRequestOrderBy("created_at_desc")
	// ListACLRequestOrderByNameAsc is [insert doc].
	ListACLRequestOrderByNameAsc = ListACLRequestOrderBy("name_asc")
	// ListACLRequestOrderByNameDesc is [insert doc].
	ListACLRequestOrderByNameDesc = ListACLRequestOrderBy("name_desc")
)

func (enum ListACLRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListACLRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListACLRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListACLRequestOrderBy(ListACLRequestOrderBy(tmp).String())
	return nil
}

type ListBackendsRequestOrderBy string

const (
	// ListBackendsRequestOrderByCreatedAtAsc is [insert doc].
	ListBackendsRequestOrderByCreatedAtAsc = ListBackendsRequestOrderBy("created_at_asc")
	// ListBackendsRequestOrderByCreatedAtDesc is [insert doc].
	ListBackendsRequestOrderByCreatedAtDesc = ListBackendsRequestOrderBy("created_at_desc")
	// ListBackendsRequestOrderByNameAsc is [insert doc].
	ListBackendsRequestOrderByNameAsc = ListBackendsRequestOrderBy("name_asc")
	// ListBackendsRequestOrderByNameDesc is [insert doc].
	ListBackendsRequestOrderByNameDesc = ListBackendsRequestOrderBy("name_desc")
)

func (enum ListBackendsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListBackendsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListBackendsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListBackendsRequestOrderBy(ListBackendsRequestOrderBy(tmp).String())
	return nil
}

type ListCertificatesRequestOrderBy string

const (
	// ListCertificatesRequestOrderByCreatedAtAsc is [insert doc].
	ListCertificatesRequestOrderByCreatedAtAsc = ListCertificatesRequestOrderBy("created_at_asc")
	// ListCertificatesRequestOrderByCreatedAtDesc is [insert doc].
	ListCertificatesRequestOrderByCreatedAtDesc = ListCertificatesRequestOrderBy("created_at_desc")
	// ListCertificatesRequestOrderByNameAsc is [insert doc].
	ListCertificatesRequestOrderByNameAsc = ListCertificatesRequestOrderBy("name_asc")
	// ListCertificatesRequestOrderByNameDesc is [insert doc].
	ListCertificatesRequestOrderByNameDesc = ListCertificatesRequestOrderBy("name_desc")
)

func (enum ListCertificatesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListCertificatesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListCertificatesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListCertificatesRequestOrderBy(ListCertificatesRequestOrderBy(tmp).String())
	return nil
}

type ListFrontendsRequestOrderBy string

const (
	// ListFrontendsRequestOrderByCreatedAtAsc is [insert doc].
	ListFrontendsRequestOrderByCreatedAtAsc = ListFrontendsRequestOrderBy("created_at_asc")
	// ListFrontendsRequestOrderByCreatedAtDesc is [insert doc].
	ListFrontendsRequestOrderByCreatedAtDesc = ListFrontendsRequestOrderBy("created_at_desc")
	// ListFrontendsRequestOrderByNameAsc is [insert doc].
	ListFrontendsRequestOrderByNameAsc = ListFrontendsRequestOrderBy("name_asc")
	// ListFrontendsRequestOrderByNameDesc is [insert doc].
	ListFrontendsRequestOrderByNameDesc = ListFrontendsRequestOrderBy("name_desc")
)

func (enum ListFrontendsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListFrontendsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListFrontendsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListFrontendsRequestOrderBy(ListFrontendsRequestOrderBy(tmp).String())
	return nil
}

type ListLbsRequestOrderBy string

const (
	// ListLbsRequestOrderByCreatedAtAsc is [insert doc].
	ListLbsRequestOrderByCreatedAtAsc = ListLbsRequestOrderBy("created_at_asc")
	// ListLbsRequestOrderByCreatedAtDesc is [insert doc].
	ListLbsRequestOrderByCreatedAtDesc = ListLbsRequestOrderBy("created_at_desc")
	// ListLbsRequestOrderByNameAsc is [insert doc].
	ListLbsRequestOrderByNameAsc = ListLbsRequestOrderBy("name_asc")
	// ListLbsRequestOrderByNameDesc is [insert doc].
	ListLbsRequestOrderByNameDesc = ListLbsRequestOrderBy("name_desc")
)

func (enum ListLbsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListLbsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListLbsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListLbsRequestOrderBy(ListLbsRequestOrderBy(tmp).String())
	return nil
}

type ListPrivateNetworksRequestOrderBy string

const (
	// ListPrivateNetworksRequestOrderByCreatedAtAsc is [insert doc].
	ListPrivateNetworksRequestOrderByCreatedAtAsc = ListPrivateNetworksRequestOrderBy("created_at_asc")
	// ListPrivateNetworksRequestOrderByCreatedAtDesc is [insert doc].
	ListPrivateNetworksRequestOrderByCreatedAtDesc = ListPrivateNetworksRequestOrderBy("created_at_desc")
)

func (enum ListPrivateNetworksRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListPrivateNetworksRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListPrivateNetworksRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListPrivateNetworksRequestOrderBy(ListPrivateNetworksRequestOrderBy(tmp).String())
	return nil
}

type ListSubscriberRequestOrderBy string

const (
	// ListSubscriberRequestOrderByCreatedAtAsc is [insert doc].
	ListSubscriberRequestOrderByCreatedAtAsc = ListSubscriberRequestOrderBy("created_at_asc")
	// ListSubscriberRequestOrderByCreatedAtDesc is [insert doc].
	ListSubscriberRequestOrderByCreatedAtDesc = ListSubscriberRequestOrderBy("created_at_desc")
	// ListSubscriberRequestOrderByNameAsc is [insert doc].
	ListSubscriberRequestOrderByNameAsc = ListSubscriberRequestOrderBy("name_asc")
	// ListSubscriberRequestOrderByNameDesc is [insert doc].
	ListSubscriberRequestOrderByNameDesc = ListSubscriberRequestOrderBy("name_desc")
)

func (enum ListSubscriberRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListSubscriberRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListSubscriberRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListSubscriberRequestOrderBy(ListSubscriberRequestOrderBy(tmp).String())
	return nil
}

type OnMarkedDownAction string

const (
	// OnMarkedDownActionOnMarkedDownActionNone is [insert doc].
	OnMarkedDownActionOnMarkedDownActionNone = OnMarkedDownAction("on_marked_down_action_none")
	// OnMarkedDownActionShutdownSessions is [insert doc].
	OnMarkedDownActionShutdownSessions = OnMarkedDownAction("shutdown_sessions")
)

func (enum OnMarkedDownAction) String() string {
	if enum == "" {
		// return default value if empty
		return "on_marked_down_action_none"
	}
	return string(enum)
}

func (enum OnMarkedDownAction) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *OnMarkedDownAction) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = OnMarkedDownAction(OnMarkedDownAction(tmp).String())
	return nil
}

type PrivateNetworkStatus string

const (
	// PrivateNetworkStatusUnknown is [insert doc].
	PrivateNetworkStatusUnknown = PrivateNetworkStatus("unknown")
	// PrivateNetworkStatusReady is [insert doc].
	PrivateNetworkStatusReady = PrivateNetworkStatus("ready")
	// PrivateNetworkStatusPending is [insert doc].
	PrivateNetworkStatusPending = PrivateNetworkStatus("pending")
	// PrivateNetworkStatusError is [insert doc].
	PrivateNetworkStatusError = PrivateNetworkStatus("error")
)

func (enum PrivateNetworkStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum PrivateNetworkStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PrivateNetworkStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PrivateNetworkStatus(PrivateNetworkStatus(tmp).String())
	return nil
}

type Protocol string

const (
	// ProtocolTCP is [insert doc].
	ProtocolTCP = Protocol("tcp")
	// ProtocolHTTP is [insert doc].
	ProtocolHTTP = Protocol("http")
)

func (enum Protocol) String() string {
	if enum == "" {
		// return default value if empty
		return "tcp"
	}
	return string(enum)
}

func (enum Protocol) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *Protocol) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = Protocol(Protocol(tmp).String())
	return nil
}

// ProxyProtocol: pROXY protocol, forward client's address (must be supported by backend servers software)
//
// The PROXY protocol informs the other end about the incoming connection, so that it can know the client's address or the public address it accessed to, whatever the upper layer protocol.
//
// * `proxy_protocol_none` Disable proxy protocol.
// * `proxy_protocol_v1` Version one (text format).
// * `proxy_protocol_v2` Version two (binary format).
// * `proxy_protocol_v2_ssl` Version two with SSL connection.
// * `proxy_protocol_v2_ssl_cn` Version two with SSL connection and common name information.
//
type ProxyProtocol string

const (
	// ProxyProtocolProxyProtocolUnknown is [insert doc].
	ProxyProtocolProxyProtocolUnknown = ProxyProtocol("proxy_protocol_unknown")
	// ProxyProtocolProxyProtocolNone is [insert doc].
	ProxyProtocolProxyProtocolNone = ProxyProtocol("proxy_protocol_none")
	// ProxyProtocolProxyProtocolV1 is [insert doc].
	ProxyProtocolProxyProtocolV1 = ProxyProtocol("proxy_protocol_v1")
	// ProxyProtocolProxyProtocolV2 is [insert doc].
	ProxyProtocolProxyProtocolV2 = ProxyProtocol("proxy_protocol_v2")
	// ProxyProtocolProxyProtocolV2Ssl is [insert doc].
	ProxyProtocolProxyProtocolV2Ssl = ProxyProtocol("proxy_protocol_v2_ssl")
	// ProxyProtocolProxyProtocolV2SslCn is [insert doc].
	ProxyProtocolProxyProtocolV2SslCn = ProxyProtocol("proxy_protocol_v2_ssl_cn")
)

func (enum ProxyProtocol) String() string {
	if enum == "" {
		// return default value if empty
		return "proxy_protocol_unknown"
	}
	return string(enum)
}

func (enum ProxyProtocol) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ProxyProtocol) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ProxyProtocol(ProxyProtocol(tmp).String())
	return nil
}

type SSLCompatibilityLevel string

const (
	// SSLCompatibilityLevelSslCompatibilityLevelUnknown is [insert doc].
	SSLCompatibilityLevelSslCompatibilityLevelUnknown = SSLCompatibilityLevel("ssl_compatibility_level_unknown")
	// SSLCompatibilityLevelSslCompatibilityLevelIntermediate is [insert doc].
	SSLCompatibilityLevelSslCompatibilityLevelIntermediate = SSLCompatibilityLevel("ssl_compatibility_level_intermediate")
	// SSLCompatibilityLevelSslCompatibilityLevelModern is [insert doc].
	SSLCompatibilityLevelSslCompatibilityLevelModern = SSLCompatibilityLevel("ssl_compatibility_level_modern")
	// SSLCompatibilityLevelSslCompatibilityLevelOld is [insert doc].
	SSLCompatibilityLevelSslCompatibilityLevelOld = SSLCompatibilityLevel("ssl_compatibility_level_old")
)

func (enum SSLCompatibilityLevel) String() string {
	if enum == "" {
		// return default value if empty
		return "ssl_compatibility_level_unknown"
	}
	return string(enum)
}

func (enum SSLCompatibilityLevel) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *SSLCompatibilityLevel) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = SSLCompatibilityLevel(SSLCompatibilityLevel(tmp).String())
	return nil
}

type StickySessionsType string

const (
	// StickySessionsTypeNone is [insert doc].
	StickySessionsTypeNone = StickySessionsType("none")
	// StickySessionsTypeCookie is [insert doc].
	StickySessionsTypeCookie = StickySessionsType("cookie")
	// StickySessionsTypeTable is [insert doc].
	StickySessionsTypeTable = StickySessionsType("table")
)

func (enum StickySessionsType) String() string {
	if enum == "" {
		// return default value if empty
		return "none"
	}
	return string(enum)
}

func (enum StickySessionsType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *StickySessionsType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = StickySessionsType(StickySessionsType(tmp).String())
	return nil
}

// ACL: the use of Access Control Lists (ACL) provide a flexible solution to perform a action generally consist in blocking or allow a request based on ip (and URL on HTTP)
type ACL struct {
	// ID: ID of your ACL ressource
	ID string `json:"id"`
	// Name: name of you ACL ressource
	Name string `json:"name"`
	// Match: the ACL match rule. At least `ip_subnet` or `http_filter` and `http_filter_value` are required
	Match *ACLMatch `json:"match"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Frontend: see the Frontend object description
	Frontend *Frontend `json:"frontend"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
}

// ACLAction: acl action
type ACLAction struct {
	// Type: the action type
	//
	// Default value: allow
	Type ACLActionType `json:"type"`
}

// ACLMatch: acl match
type ACLMatch struct {
	// IPSubnet: a list of IPs or CIDR v4/v6 addresses of the client of the session to match
	IPSubnet []*string `json:"ip_subnet"`
	// HTTPFilter: the HTTP filter to match
	//
	// The HTTP filter to match. This filter is supported only if your backend supports HTTP forwarding.
	// It extracts the request's URL path, which starts at the first slash and ends before the question mark (without the host part).
	//
	// Default value: acl_http_filter_none
	HTTPFilter ACLHTTPFilter `json:"http_filter"`
	// HTTPFilterValue: a list of possible values to match for the given HTTP filter
	HTTPFilterValue []*string `json:"http_filter_value"`
	// Invert: if set to `true`, the ACL matching condition will be of type "UNLESS"
	Invert bool `json:"invert"`
}

// Backend: backend
type Backend struct {
	ID string `json:"id"`

	Name string `json:"name"`
	// ForwardProtocol:
	//
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`

	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm:
	//
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions:
	//
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`

	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`

	HealthCheck *HealthCheck `json:"health_check"`

	Pool []string `json:"pool"`

	Lb *Lb `json:"lb"`

	SendProxyV2 bool `json:"send_proxy_v2"`

	TimeoutServer *time.Duration `json:"timeout_server"`

	TimeoutConnect *time.Duration `json:"timeout_connect"`

	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction:
	//
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol:
	//
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
}

func (m *Backend) UnmarshalJSON(b []byte) error {
	type tmpType Backend
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = Backend(tmp.tmpType)

	m.TimeoutServer = tmp.TmpTimeoutServer.Standard()
	m.TimeoutConnect = tmp.TmpTimeoutConnect.Standard()
	m.TimeoutTunnel = tmp.TmpTimeoutTunnel.Standard()
	return nil
}

func (m Backend) MarshalJSON() ([]byte, error) {
	type tmpType Backend
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

// BackendServerStats: state and statistics of your backend server like last healthcheck status, server uptime, result state of your backend server
type BackendServerStats struct {
	// InstanceID: ID of your loadbalancer cluster server
	InstanceID string `json:"instance_id"`
	// BackendID: ID of your Backend
	BackendID string `json:"backend_id"`
	// IP: iPv4 or IPv6 address of the server backend
	IP string `json:"ip"`
	// ServerState: server operational state (stopped/starting/running/stopping)
	//
	// Default value: stopped
	ServerState BackendServerStatsServerState `json:"server_state"`
	// ServerStateChangedAt: time since last operational change
	ServerStateChangedAt time.Time `json:"server_state_changed_at"`
	// LastHealthCheckStatus: last health check status (unknown/neutral/failed/passed/condpass)
	//
	// Default value: unknown
	LastHealthCheckStatus BackendServerStatsHealthCheckStatus `json:"last_health_check_status"`
}

// Certificate: sSL certificate
type Certificate struct {
	// Type: type of certificate (Let's encrypt or custom)
	//
	// Default value: letsencryt
	Type CertificateType `json:"type"`
	// ID: certificate ID
	ID string `json:"id"`
	// CommonName: main domain name of certificate
	CommonName string `json:"common_name"`
	// SubjectAlternativeName: alternative domain names
	SubjectAlternativeName []string `json:"subject_alternative_name"`
	// Fingerprint: identifier (SHA-1) of the certificate
	Fingerprint string `json:"fingerprint"`
	// NotValidBefore: validity bounds
	NotValidBefore time.Time `json:"not_valid_before"`
	// NotValidAfter: validity bounds
	NotValidAfter time.Time `json:"not_valid_after"`
	// Status: status of certificate
	//
	// Default value: pending
	Status CertificateStatus `json:"status"`
	// Lb: load Balancer object
	Lb *Lb `json:"lb"`
	// Name: certificate name
	Name string `json:"name"`
}

// CreateCertificateRequestCustomCertificate: import a custom SSL certificate
type CreateCertificateRequestCustomCertificate struct {
	// CertificateChain: the full PEM-formatted include an entire certificate chain including public key, private key, and optionally certificate authorities.
	CertificateChain string `json:"certificate_chain"`
}

// CreateCertificateRequestLetsencryptConfig: generate a new SSL certificate using Let's Encrypt.
type CreateCertificateRequestLetsencryptConfig struct {
	// CommonName: main domain name of certificate (make sure this domain exists and resolves to your Load Balancer HA IP)
	CommonName string `json:"common_name"`
	// SubjectAlternativeName: alternative domain names (make sure all domain names exists and resolves to your Load Balancer HA IP)
	SubjectAlternativeName []string `json:"subject_alternative_name"`
}

// Frontend: frontend
type Frontend struct {
	ID string `json:"id"`

	Name string `json:"name"`

	InboundPort int32 `json:"inbound_port"`

	Backend *Backend `json:"backend"`

	Lb *Lb `json:"lb"`

	TimeoutClient *time.Duration `json:"timeout_client"`

	Certificate *Certificate `json:"certificate"`
}

func (m *Frontend) UnmarshalJSON(b []byte) error {
	type tmpType Frontend
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = Frontend(tmp.tmpType)

	m.TimeoutClient = tmp.TmpTimeoutClient.Standard()
	return nil
}

func (m Frontend) MarshalJSON() ([]byte, error) {
	type tmpType Frontend
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

// HealthCheck: health check
type HealthCheck struct {
	// MysqlConfig: the check requires MySQL >=3.22, for older versions, use TCP check
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`
	// LdapConfig: the response is analyzed to find an LDAPv3 response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`
	// RedisConfig: the response is analyzed to find the +PONG response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`

	CheckMaxRetries int32 `json:"check_max_retries"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`

	Port int32 `json:"port"`

	CheckTimeout *time.Duration `json:"check_timeout"`

	CheckDelay *time.Duration `json:"check_delay"`
}

func (m *HealthCheck) UnmarshalJSON(b []byte) error {
	type tmpType HealthCheck
	tmp := struct {
		tmpType

		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = HealthCheck(tmp.tmpType)

	m.CheckTimeout = tmp.TmpCheckTimeout.Standard()
	m.CheckDelay = tmp.TmpCheckDelay.Standard()
	return nil
}

func (m HealthCheck) MarshalJSON() ([]byte, error) {
	type tmpType HealthCheck
	tmp := struct {
		tmpType

		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
	}{
		tmpType: tmpType(m),

		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
	}
	return json.Marshal(tmp)
}

type HealthCheckHTTPConfig struct {
	URI string `json:"uri"`

	Method string `json:"method"`

	Code *int32 `json:"code"`
}

type HealthCheckHTTPSConfig struct {
	URI string `json:"uri"`

	Method string `json:"method"`

	Code *int32 `json:"code"`
}

type HealthCheckLdapConfig struct {
}

type HealthCheckMysqlConfig struct {
	User string `json:"user"`
}

type HealthCheckPgsqlConfig struct {
	User string `json:"user"`
}

type HealthCheckRedisConfig struct {
}

type HealthCheckTCPConfig struct {
}

// IP: ip
type IP struct {
	ID string `json:"id"`

	IPAddress string `json:"ip_address"`

	OrganizationID string `json:"organization_id"`

	LbID *string `json:"lb_id"`

	Reverse string `json:"reverse"`

	Region scw.Region `json:"region"`
}

type Instance struct {
	ID string `json:"id"`
	// Status:
	//
	// Default value: unknown
	Status InstanceStatus `json:"status"`

	IPAddress string `json:"ip_address"`

	Region scw.Region `json:"region"`
}

// Lb: lb
type Lb struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Description string `json:"description"`
	// Status:
	//
	// Default value: unknown
	Status LbStatus `json:"status"`

	Instances []*Instance `json:"instances"`

	OrganizationID string `json:"organization_id"`

	IP []*IP `json:"ip"`

	Tags []string `json:"tags"`

	FrontendCount int32 `json:"frontend_count"`

	BackendCount int32 `json:"backend_count"`

	Type string `json:"type"`

	Subscriber *Subscriber `json:"subscriber"`
	// SslCompatibilityLevel:
	//
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`

	Region scw.Region `json:"region"`
}

// LbStats: lb stats
type LbStats struct {
	// BackendServersStats: list stats object of your loadbalancer
	BackendServersStats []*BackendServerStats `json:"backend_servers_stats"`
}

type LbType struct {
	Name string `json:"name"`
	// StockStatus:
	//
	// Default value: unknown
	StockStatus LbTypeStock `json:"stock_status"`

	Description string `json:"description"`

	Region scw.Region `json:"region"`
}

// ListACLResponse: list acl response
type ListACLResponse struct {
	// ACLs: list of Acl object (see Acl object description)
	ACLs []*ACL `json:"acls"`
	// TotalCount: the total number of items
	TotalCount uint32 `json:"total_count"`
}

// ListBackendStatsResponse: list backend stats response
type ListBackendStatsResponse struct {
	// BackendServersStats: list backend stats object of your loadbalancer
	BackendServersStats []*BackendServerStats `json:"backend_servers_stats"`
	// TotalCount: the total number of items
	TotalCount uint32 `json:"total_count"`
}

// ListBackendsResponse: list backends response
type ListBackendsResponse struct {
	// Backends: list Backend objects of a Load Balancer
	Backends []*Backend `json:"backends"`
	// TotalCount: total count, wihtout pagination
	TotalCount uint32 `json:"total_count"`
}

type ListCertificatesResponse struct {
	Certificates []*Certificate `json:"certificates"`

	TotalCount uint32 `json:"total_count"`
}

// ListFrontendsResponse: list frontends response
type ListFrontendsResponse struct {
	// Frontends: list frontends object of your loadbalancer
	Frontends []*Frontend `json:"frontends"`
	// TotalCount: total count, wihtout pagination
	TotalCount uint32 `json:"total_count"`
}

// ListIPsResponse: list ips response
type ListIPsResponse struct {
	// IPs: list IP address object
	IPs []*IP `json:"ips"`
	// TotalCount: total count, wihtout pagination
	TotalCount uint32 `json:"total_count"`
}

type ListLbPrivateNetworksResponse struct {
	PrivateNetwork []*PrivateNetwork `json:"private_network"`

	TotalCount uint32 `json:"total_count"`
}

type ListLbTypesResponse struct {
	LbTypes []*LbType `json:"lb_types"`

	TotalCount uint32 `json:"total_count"`
}

// ListLbsResponse: list lbs response
type ListLbsResponse struct {
	Lbs []*Lb `json:"lbs"`

	TotalCount uint32 `json:"total_count"`
}

// ListSubscriberResponse: list subscriber response
type ListSubscriberResponse struct {
	// Subscribers: list of Subscribers object
	Subscribers []*Subscriber `json:"subscribers"`
	// TotalCount: the total number of items
	TotalCount uint32 `json:"total_count"`
}

// PrivateNetwork: private network
type PrivateNetwork struct {
	// Lb: loadBalancer object
	Lb *Lb `json:"lb"`
	// IPAddress: local ip address of Load Balancer instance
	IPAddress []string `json:"ip_address"`
	// PrivateNetworkID: instance private network id
	PrivateNetworkID string `json:"private_network_id"`
	// Status: status (running, to create...) of private network connection
	//
	// Default value: unknown
	Status PrivateNetworkStatus `json:"status"`
}

// Subscriber: subscriber
type Subscriber struct {
	// ID: subscriber ID
	ID string `json:"id"`
	// Name: subscriber name
	Name string `json:"name"`
	// EmailConfig: email address of subscriber
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI of subscriber
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
}

// SubscriberEmailConfig: email alert of subscriber
type SubscriberEmailConfig struct {
	// Email: email who receive alert
	Email string `json:"email"`
}

// SubscriberWebhookConfig: webhook alert of subscriber
type SubscriberWebhookConfig struct {
	// URI: URI who receive POST request
	URI string `json:"uri"`
}

// Service API

type GetServiceInfoRequest struct {
	Region scw.Region `json:"-"`
}

func (s *API) GetServiceInfo(req *GetServiceInfoRequest, opts ...scw.RequestOption) (*scw.ServiceInfo, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "",
		Headers: http.Header{},
	}

	var resp scw.ServiceInfo

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListLbsRequest struct {
	Region scw.Region `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListLbsRequestOrderBy `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`

	OrganizationID *string `json:"-"`
}

func (s *API) ListLbs(req *ListLbsRequest, opts ...scw.RequestOption) (*ListLbsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLbsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLbsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLbsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListLbsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Lbs = append(r.Lbs, results.Lbs...)
	r.TotalCount += uint32(len(results.Lbs))
	return uint32(len(results.Lbs)), nil
}

type CreateLbRequest struct {
	Region scw.Region `json:"-"`
	// OrganizationID: owner of resources
	OrganizationID string `json:"organization_id"`
	// Name: resource names
	Name string `json:"name"`
	// Description: resource description
	Description string `json:"description"`
	// IPID: just like for compute instances, when you destroy a Load Balancer, you can keep its highly available IP address and reuse it for another Load Balancer later
	IPID *string `json:"ip_id"`
	// Tags: list of keyword
	Tags []string `json:"tags"`
	// Type: load Balancer offer type
	Type string `json:"type"`
	// SslCompatibilityLevel:
	//
	// Enforces minimal SSL version (in SSL/TLS offloading context).
	// - `intermediate` General-purpose servers with a variety of clients, recommended for almost all systems (Supports Firefox 27, Android 4.4.2, Chrome 31, Edge, IE 11 on Windows 7, Java 8u31, OpenSSL 1.0.1, Opera 20, and Safari 9).
	// - `modern` Services with clients that support TLS 1.3 and don't need backward compatibility (Firefox 63, Android 10.0, Chrome 70, Edge 75, Java 11, OpenSSL 1.1.1, Opera 57, and Safari 12.1).
	// - `old` Compatible with a number of very old clients, and should be used only as a last resort (Firefox 1, Android 2.3, Chrome 1, Edge 12, IE8 on Windows XP, Java 6, OpenSSL 0.9.8, Opera 5, and Safari 1).
	//
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

func (s *API) CreateLb(req *CreateLbRequest, opts ...scw.RequestOption) (*Lb, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Lb

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetLbRequest struct {
	Region scw.Region `json:"-"`

	LbID string `json:"-"`
}

func (s *API) GetLb(req *GetLbRequest, opts ...scw.RequestOption) (*Lb, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "",
		Headers: http.Header{},
	}

	var resp Lb

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateLbRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// Description: resource description
	Description string `json:"description"`
	// Tags: list of keywords
	Tags []string `json:"tags"`
	// SslCompatibilityLevel:
	//
	// Enforces minimal SSL version (in SSL/TLS offloading context).
	// - `intermediate` General-purpose servers with a variety of clients, recommended for almost all systems (Supports Firefox 27, Android 4.4.2, Chrome 31, Edge, IE 11 on Windows 7, Java 8u31, OpenSSL 1.0.1, Opera 20, and Safari 9).
	// - `modern` Services with clients that support TLS 1.3 and don't need backward compatibility (Firefox 63, Android 10.0, Chrome 70, Edge 75, Java 11, OpenSSL 1.1.1, Opera 57, and Safari 12.1).
	// - `old` Compatible with a number of very old clients, and should be used only as a last resort (Firefox 1, Android 2.3, Chrome 1, Edge 12, IE8 on Windows XP, Java 6, OpenSSL 0.9.8, Opera 5, and Safari 1).
	//
	// Default value: ssl_compatibility_level_unknown
	SslCompatibilityLevel SSLCompatibilityLevel `json:"ssl_compatibility_level"`
}

func (s *API) UpdateLb(req *UpdateLbRequest, opts ...scw.RequestOption) (*Lb, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Lb

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteLbRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// ReleaseIP: set true if you don't want to keep this IP address
	ReleaseIP bool `json:"-"`
}

func (s *API) DeleteLb(req *DeleteLbRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	query := url.Values{}
	parameter.AddToQuery(query, "release_ip", req.ReleaseIP)

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type MigrateLbRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// Type: load Balancer type (check /lb-types to list all type)
	Type string `json:"type"`
}

// MigrateLb: migrate Load Balancer
func (s *API) MigrateLb(req *MigrateLbRequest, opts ...scw.RequestOption) (*Lb, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/migrate",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Lb

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListIPsRequest struct {
	Region scw.Region `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// IPAddress: use this to search by IP address
	IPAddress *string `json:"-"`

	OrganizationID *string `json:"-"`
}

// ListIPs: list IPs
func (s *API) ListIPs(req *ListIPsRequest, opts ...scw.RequestOption) (*ListIPsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "ip_address", req.IPAddress)
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListIPsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListIPsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListIPsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.IPs = append(r.IPs, results.IPs...)
	r.TotalCount += uint32(len(results.IPs))
	return uint32(len(results.IPs)), nil
}

type CreateIPRequest struct {
	Region scw.Region `json:"-"`
	// OrganizationID: owner of resources
	OrganizationID string `json:"organization_id"`
	// Reverse: reverse domain name
	Reverse *string `json:"reverse"`
}

// CreateIP: create IP
func (s *API) CreateIP(req *CreateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips",
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

type GetIPRequest struct {
	Region scw.Region `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
}

// GetIP: get IP
func (s *API) GetIP(req *GetIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	var resp IP

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ReleaseIPRequest struct {
	Region scw.Region `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
}

// ReleaseIP: release IP
func (s *API) ReleaseIP(req *ReleaseIPRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type UpdateIPRequest struct {
	Region scw.Region `json:"-"`
	// IPID: IP address ID
	IPID string `json:"-"`
	// Reverse: reverse DNS
	Reverse *string `json:"reverse"`
}

// UpdateIP: update IP
func (s *API) UpdateIP(req *UpdateIPRequest, opts ...scw.RequestOption) (*IP, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/ips/" + fmt.Sprint(req.IPID) + "",
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

type ListBackendsRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy: choose order of response
	//
	// Default value: created_at_asc
	OrderBy ListBackendsRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to returns
	PageSize *uint32 `json:"-"`
}

func (s *API) ListBackends(req *ListBackendsRequest, opts ...scw.RequestOption) (*ListBackendsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/backends",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListBackendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListBackendsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListBackendsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListBackendsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Backends = append(r.Backends, results.Backends...)
	r.TotalCount += uint32(len(results.Backends))
	return uint32(len(results.Backends)), nil
}

type CreateBackendRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// ForwardProtocol: backend protocol. TCP or HTTP
	//
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`
	// ForwardPort: user sessions will be forwarded to this port of backend servers
	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm: load balancing algorithm
	//
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions: enables cookie-based session persistence
	//
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`
	// StickySessionsCookieName: cookie name for for sticky sessions
	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`
	// HealthCheck: see the Healthcheck object description
	HealthCheck *HealthCheck `json:"health_check"`
	// ServerIP: backend server IP addresses list (IPv4 or IPv6)
	ServerIP []string `json:"server_ip"`
	// SendProxyV2: deprecated in favor of proxy_protocol field !
	SendProxyV2 bool `json:"send_proxy_v2"`
	// TimeoutServer: maximum server connection inactivity time
	TimeoutServer *time.Duration `json:"timeout_server"`
	// TimeoutConnect: maximum initical server connection establishment time
	TimeoutConnect *time.Duration `json:"timeout_connect"`
	// TimeoutTunnel: maximum tunnel inactivity time
	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction: modify what occurs when a backend server is marked down
	//
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol: pROXY protocol, forward client's address (must be supported by backend servers software)
	//
	// The PROXY protocol informs the other end about the incoming connection, so that it can know the client's address or the public address it accessed to, whatever the upper layer protocol.
	//
	// * `proxy_protocol_none` Disable proxy protocol.
	// * `proxy_protocol_v1` Version one (text format).
	// * `proxy_protocol_v2` Version two (binary format).
	// * `proxy_protocol_v2_ssl` Version two with SSL connection.
	// * `proxy_protocol_v2_ssl_cn` Version two with SSL connection and common name information.
	//
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
}

func (m *CreateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType CreateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = CreateBackendRequest(tmp.tmpType)

	m.TimeoutServer = tmp.TmpTimeoutServer.Standard()
	m.TimeoutConnect = tmp.TmpTimeoutConnect.Standard()
	m.TimeoutTunnel = tmp.TmpTimeoutTunnel.Standard()
	return nil
}

func (m CreateBackendRequest) MarshalJSON() ([]byte, error) {
	type tmpType CreateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

func (s *API) CreateBackend(req *CreateBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/backends",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetBackendRequest struct {
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
}

func (s *API) GetBackend(req *GetBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateBackendRequest struct {
	Region scw.Region `json:"-"`

	BackendID string `json:"-"`

	Name string `json:"name"`
	// ForwardProtocol:
	//
	// Default value: tcp
	ForwardProtocol Protocol `json:"forward_protocol"`

	ForwardPort int32 `json:"forward_port"`
	// ForwardPortAlgorithm:
	//
	// Default value: roundrobin
	ForwardPortAlgorithm ForwardPortAlgorithm `json:"forward_port_algorithm"`
	// StickySessions:
	//
	// Default value: none
	StickySessions StickySessionsType `json:"sticky_sessions"`

	StickySessionsCookieName string `json:"sticky_sessions_cookie_name"`

	SendProxyV2 bool `json:"send_proxy_v2"`

	TimeoutServer *time.Duration `json:"timeout_server"`

	TimeoutConnect *time.Duration `json:"timeout_connect"`

	TimeoutTunnel *time.Duration `json:"timeout_tunnel"`
	// OnMarkedDownAction:
	//
	// Default value: on_marked_down_action_none
	OnMarkedDownAction OnMarkedDownAction `json:"on_marked_down_action"`
	// ProxyProtocol:
	//
	// Default value: proxy_protocol_unknown
	ProxyProtocol ProxyProtocol `json:"proxy_protocol"`
}

func (m *UpdateBackendRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = UpdateBackendRequest(tmp.tmpType)

	m.TimeoutServer = tmp.TmpTimeoutServer.Standard()
	m.TimeoutConnect = tmp.TmpTimeoutConnect.Standard()
	m.TimeoutTunnel = tmp.TmpTimeoutTunnel.Standard()
	return nil
}

func (m UpdateBackendRequest) MarshalJSON() ([]byte, error) {
	type tmpType UpdateBackendRequest
	tmp := struct {
		tmpType

		TmpTimeoutServer  *marshaler.Duration `json:"timeout_server"`
		TmpTimeoutConnect *marshaler.Duration `json:"timeout_connect"`
		TmpTimeoutTunnel  *marshaler.Duration `json:"timeout_tunnel"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutServer:  marshaler.NewDuration(m.TimeoutServer),
		TmpTimeoutConnect: marshaler.NewDuration(m.TimeoutConnect),
		TmpTimeoutTunnel:  marshaler.NewDuration(m.TimeoutTunnel),
	}
	return json.Marshal(tmp)
}

func (s *API) UpdateBackend(req *UpdateBackendRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteBackendRequest struct {
	Region scw.Region `json:"-"`
	// BackendID: ID of the backend to delete
	BackendID string `json:"-"`
}

func (s *API) DeleteBackend(req *DeleteBackendRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type AddBackendServersRequest struct {
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to add on your backend
	ServerIP []string `json:"server_ip"`
}

func (s *API) AddBackendServers(req *AddBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RemoveBackendServersRequest struct {
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to remove of your backend
	ServerIP []string `json:"server_ip"`
}

func (s *API) RemoveBackendServers(req *RemoveBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type SetBackendServersRequest struct {
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// ServerIP: set all IPs to add on your backend and remove all other
	ServerIP []string `json:"server_ip"`
}

func (s *API) SetBackendServers(req *SetBackendServersRequest, opts ...scw.RequestOption) (*Backend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Backend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateHealthCheckRequest struct {
	Region scw.Region `json:"-"`
	// BackendID: backend ID
	BackendID string `json:"-"`
	// Port: specify the port used to health check
	Port int32 `json:"port"`
	// CheckDelay: time between two consecutive health checks
	CheckDelay *time.Duration `json:"check_delay"`
	// CheckTimeout: additional check timeout, after the connection has been already established
	CheckTimeout *time.Duration `json:"check_timeout"`
	// CheckMaxRetries: number of consecutive unsuccessful health checks, after wich the server will be considered dead
	CheckMaxRetries int32 `json:"check_max_retries"`
	// MysqlConfig: the check requires MySQL >=3.22, for older version, please use TCP check
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	MysqlConfig *HealthCheckMysqlConfig `json:"mysql_config,omitempty"`
	// LdapConfig: the response is analyzed to find an LDAPv3 response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	LdapConfig *HealthCheckLdapConfig `json:"ldap_config,omitempty"`
	// RedisConfig: the response is analyzed to find the +PONG response message
	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	RedisConfig *HealthCheckRedisConfig `json:"redis_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	PgsqlConfig *HealthCheckPgsqlConfig `json:"pgsql_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	TCPConfig *HealthCheckTCPConfig `json:"tcp_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPConfig *HealthCheckHTTPConfig `json:"http_config,omitempty"`

	// Precisely one of HTTPConfig, HTTPSConfig, LdapConfig, MysqlConfig, PgsqlConfig, RedisConfig, TCPConfig must be set.
	HTTPSConfig *HealthCheckHTTPSConfig `json:"https_config,omitempty"`
}

func (m *UpdateHealthCheckRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateHealthCheckRequest
	tmp := struct {
		tmpType

		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = UpdateHealthCheckRequest(tmp.tmpType)

	m.CheckDelay = tmp.TmpCheckDelay.Standard()
	m.CheckTimeout = tmp.TmpCheckTimeout.Standard()
	return nil
}

func (m UpdateHealthCheckRequest) MarshalJSON() ([]byte, error) {
	type tmpType UpdateHealthCheckRequest
	tmp := struct {
		tmpType

		TmpCheckDelay   *marshaler.Duration `json:"check_delay"`
		TmpCheckTimeout *marshaler.Duration `json:"check_timeout"`
	}{
		tmpType: tmpType(m),

		TmpCheckDelay:   marshaler.NewDuration(m.CheckDelay),
		TmpCheckTimeout: marshaler.NewDuration(m.CheckTimeout),
	}
	return json.Marshal(tmp)
}

func (s *API) UpdateHealthCheck(req *UpdateHealthCheckRequest, opts ...scw.RequestOption) (*HealthCheck, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.BackendID) == "" {
		return nil, errors.New("field BackendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/backends/" + fmt.Sprint(req.BackendID) + "/healthcheck",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp HealthCheck

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListFrontendsRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrderBy: response order
	//
	// Default value: created_at_asc
	OrderBy ListFrontendsRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to returns
	PageSize *uint32 `json:"-"`
}

func (s *API) ListFrontends(req *ListFrontendsRequest, opts ...scw.RequestOption) (*ListFrontendsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/frontends",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListFrontendsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListFrontendsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListFrontendsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListFrontendsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Frontends = append(r.Frontends, results.Frontends...)
	r.TotalCount += uint32(len(results.Frontends))
	return uint32(len(results.Frontends)), nil
}

type CreateFrontendRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// InboundPort: TCP port to listen on the front side
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID
	BackendID string `json:"backend_id"`
	// TimeoutClient: set the maximum inactivity time on the client side
	TimeoutClient *time.Duration `json:"timeout_client"`
	// CertificateID: certificate ID
	CertificateID *string `json:"certificate_id"`
}

func (m *CreateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType CreateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = CreateFrontendRequest(tmp.tmpType)

	m.TimeoutClient = tmp.TmpTimeoutClient.Standard()
	return nil
}

func (m CreateFrontendRequest) MarshalJSON() ([]byte, error) {
	type tmpType CreateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

func (s *API) CreateFrontend(req *CreateFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/frontends",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetFrontendRequest struct {
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID
	FrontendID string `json:"-"`
}

func (s *API) GetFrontend(req *GetFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateFrontendRequest struct {
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID
	FrontendID string `json:"-"`
	// Name: resource name
	Name string `json:"name"`
	// InboundPort: TCP port to listen on the front side
	InboundPort int32 `json:"inbound_port"`
	// BackendID: backend ID
	BackendID string `json:"backend_id"`
	// TimeoutClient: client session maximum inactivity time
	TimeoutClient *time.Duration `json:"timeout_client"`
	// CertificateID: certificate ID
	CertificateID *string `json:"certificate_id"`
}

func (m *UpdateFrontendRequest) UnmarshalJSON(b []byte) error {
	type tmpType UpdateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	*m = UpdateFrontendRequest(tmp.tmpType)

	m.TimeoutClient = tmp.TmpTimeoutClient.Standard()
	return nil
}

func (m UpdateFrontendRequest) MarshalJSON() ([]byte, error) {
	type tmpType UpdateFrontendRequest
	tmp := struct {
		tmpType

		TmpTimeoutClient *marshaler.Duration `json:"timeout_client"`
	}{
		tmpType: tmpType(m),

		TmpTimeoutClient: marshaler.NewDuration(m.TimeoutClient),
	}
	return json.Marshal(tmp)
}

func (s *API) UpdateFrontend(req *UpdateFrontendRequest, opts ...scw.RequestOption) (*Frontend, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Frontend

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteFrontendRequest struct {
	Region scw.Region `json:"-"`
	// FrontendID: frontend ID to delete
	FrontendID string `json:"-"`
}

func (s *API) DeleteFrontend(req *DeleteFrontendRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type GetLbStatsRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
}

func (s *API) GetLbStats(req *GetLbStatsRequest, opts ...scw.RequestOption) (*LbStats, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/stats",
		Headers: http.Header{},
	}

	var resp LbStats

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListBackendStatsRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
}

func (s *API) ListBackendStats(req *ListBackendStatsRequest, opts ...scw.RequestOption) (*ListBackendStatsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/backend-stats",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListBackendStatsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListBackendStatsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListBackendStatsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListBackendStatsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.BackendServersStats = append(r.BackendServersStats, results.BackendServersStats...)
	r.TotalCount += uint32(len(results.BackendServersStats))
	return uint32(len(results.BackendServersStats)), nil
}

type ListACLsRequest struct {
	Region scw.Region `json:"-"`
	// FrontendID: ID of your frontend
	FrontendID string `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListACLRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: filter acl per name
	Name *string `json:"-"`
}

func (s *API) ListACLs(req *ListACLsRequest, opts ...scw.RequestOption) (*ListACLResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListACLResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListACLResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListACLResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListACLResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.ACLs = append(r.ACLs, results.ACLs...)
	r.TotalCount += uint32(len(results.ACLs))
	return uint32(len(results.ACLs)), nil
}

type CreateACLRequest struct {
	Region scw.Region `json:"-"`
	// FrontendID: ID of your frontend
	FrontendID string `json:"-"`
	// Name: name of your ACL ressource
	Name string `json:"name"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Match: the ACL match rule
	//
	// The ACL match rule. You can have one of those three cases:
	//
	//   - `ip_subnet` is defined
	//   - `http_filter` and `http_filter_value` are defined
	//   - `ip_subnet`, `http_filter` and `http_filter_value` are defined
	//
	Match *ACLMatch `json:"match"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
}

func (s *API) CreateACL(req *CreateACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.FrontendID) == "" {
		return nil, errors.New("field FrontendID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/frontends/" + fmt.Sprint(req.FrontendID) + "/acls",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetACLRequest struct {
	Region scw.Region `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
}

func (s *API) GetACL(req *GetACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return nil, errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateACLRequest struct {
	Region scw.Region `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
	// Name: name of your ACL ressource
	Name string `json:"name"`
	// Action: action to undertake when an ACL filter matches
	Action *ACLAction `json:"action"`
	// Match: the ACL match rule. At least `ip_subnet` or `http_filter` and `http_filter_value` are required
	Match *ACLMatch `json:"match"`
	// Index: order between your Acls (ascending order, 0 is first acl executed)
	Index int32 `json:"index"`
}

func (s *API) UpdateACL(req *UpdateACLRequest, opts ...scw.RequestOption) (*ACL, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return nil, errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ACL

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteACLRequest struct {
	Region scw.Region `json:"-"`
	// ACLID: ID of your ACL ressource
	ACLID string `json:"-"`
}

func (s *API) DeleteACL(req *DeleteACLRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ACLID) == "" {
		return errors.New("field ACLID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/acls/" + fmt.Sprint(req.ACLID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type CreateCertificateRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// Name: certificate name
	Name string `json:"name"`
	// Letsencrypt: let's Encrypt type
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	Letsencrypt *CreateCertificateRequestLetsencryptConfig `json:"letsencrypt,omitempty"`
	// CustomCertificate: custom import certificate
	// Precisely one of CustomCertificate, Letsencrypt must be set.
	CustomCertificate *CreateCertificateRequestCustomCertificate `json:"custom_certificate,omitempty"`
}

// CreateCertificate: create Certificate
//
// Generate a new SSL certificate using Let's Encrypt or import your certificate.
func (s *API) CreateCertificate(req *CreateCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/certificates",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListCertificatesRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListCertificatesRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
}

// ListCertificates: list Certificates
func (s *API) ListCertificates(req *ListCertificatesRequest, opts ...scw.RequestOption) (*ListCertificatesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/certificates",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListCertificatesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListCertificatesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListCertificatesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListCertificatesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Certificates = append(r.Certificates, results.Certificates...)
	r.TotalCount += uint32(len(results.Certificates))
	return uint32(len(results.Certificates)), nil
}

type GetCertificateRequest struct {
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
}

// GetCertificate: get Certificate
func (s *API) GetCertificate(req *GetCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return nil, errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateCertificateRequest struct {
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
	// Name: certificate name
	Name string `json:"name"`
}

// UpdateCertificate: update Certificate
func (s *API) UpdateCertificate(req *UpdateCertificateRequest, opts ...scw.RequestOption) (*Certificate, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return nil, errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Certificate

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteCertificateRequest struct {
	Region scw.Region `json:"-"`
	// CertificateID: certificate ID
	CertificateID string `json:"-"`
}

// DeleteCertificate: delete Certificate
func (s *API) DeleteCertificate(req *DeleteCertificateRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.CertificateID) == "" {
		return errors.New("field CertificateID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/certificates/" + fmt.Sprint(req.CertificateID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListLbTypesRequest struct {
	Region scw.Region `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
}

// ListLbTypes: list all Load Balancer offer type
func (s *API) ListLbTypes(req *ListLbTypesRequest, opts ...scw.RequestOption) (*ListLbTypesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb-types",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLbTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLbTypesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLbTypesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListLbTypesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.LbTypes = append(r.LbTypes, results.LbTypes...)
	r.TotalCount += uint32(len(results.LbTypes))
	return uint32(len(results.LbTypes)), nil
}

type CreateSubscriberRequest struct {
	Region scw.Region `json:"-"`
	// Name: subscriber name
	Name string `json:"name"`
	// EmailConfig: email address configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
	// OrganizationID: owner of resources
	OrganizationID string `json:"organization_id"`
}

// CreateSubscriber: create a subscriber, webhook or email
func (s *API) CreateSubscriber(req *CreateSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetSubscriberRequest struct {
	Region scw.Region `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"-"`
}

// GetSubscriber: get a subscriber
func (s *API) GetSubscriber(req *GetSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return nil, errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListSubscriberRequest struct {
	Region scw.Region `json:"-"`
	// OrderBy: you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListSubscriberRequestOrderBy `json:"-"`
	// Page: page number
	Page *int32 `json:"-"`
	// PageSize: the number of items to return
	PageSize *uint32 `json:"-"`
	// Name: use this to search by name
	Name *string `json:"-"`
	// OrganizationID: owner of resources
	OrganizationID *string `json:"-"`
}

// ListSubscriber: list all subscriber
func (s *API) ListSubscriber(req *ListSubscriberRequest, opts ...scw.RequestOption) (*ListSubscriberResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

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

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSubscriberResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSubscriberResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSubscriberResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListSubscriberResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Subscribers = append(r.Subscribers, results.Subscribers...)
	r.TotalCount += uint32(len(results.Subscribers))
	return uint32(len(results.Subscribers)), nil
}

type UpdateSubscriberRequest struct {
	Region scw.Region `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"-"`
	// Name: subscriber name
	Name string `json:"name"`
	// EmailConfig: email address configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	EmailConfig *SubscriberEmailConfig `json:"email_config,omitempty"`
	// WebhookConfig: webHook URI configuration
	// Precisely one of EmailConfig, WebhookConfig must be set.
	WebhookConfig *SubscriberWebhookConfig `json:"webhook_config,omitempty"`
}

// UpdateSubscriber: update a subscriber
func (s *API) UpdateSubscriber(req *UpdateSubscriberRequest, opts ...scw.RequestOption) (*Subscriber, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return nil, errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/subscribers/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Subscriber

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteSubscriberRequest struct {
	Region scw.Region `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"-"`
}

// DeleteSubscriber: delete a subscriber
func (s *API) DeleteSubscriber(req *DeleteSubscriberRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.SubscriberID) == "" {
		return errors.New("field SubscriberID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/subscription/" + fmt.Sprint(req.SubscriberID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type SubscribeToLbRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// SubscriberID: subscriber ID
	SubscriberID string `json:"subscriber_id"`
}

// SubscribeToLb: link Load Balancer to a subscriber
func (s *API) SubscribeToLb(req *SubscribeToLbRequest, opts ...scw.RequestOption) (*Lb, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/" + fmt.Sprint(req.LbID) + "/subscribe",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Lb

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UnsubscribeFromLbRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
}

// UnsubscribeFromLb: remove link between Load Balancer and subscriber
func (s *API) UnsubscribeFromLb(req *UnsubscribeFromLbRequest, opts ...scw.RequestOption) (*Lb, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lb/" + fmt.Sprint(req.LbID) + "/unsubscribe",
		Headers: http.Header{},
	}

	var resp Lb

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListLbPrivateNetworksRequest struct {
	Region scw.Region `json:"-"`

	LbID string `json:"-"`
	// OrderBy:
	//
	// Default value: created_at_asc
	OrderBy ListPrivateNetworksRequestOrderBy `json:"-"`

	PageSize *uint32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListLbPrivateNetworks: bETA - List attached private network of Load Balancer
func (s *API) ListLbPrivateNetworks(req *ListLbPrivateNetworksRequest, opts ...scw.RequestOption) (*ListLbPrivateNetworksResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/private-networks",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListLbPrivateNetworksResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListLbPrivateNetworksResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListLbPrivateNetworksResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListLbPrivateNetworksResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.PrivateNetwork = append(r.PrivateNetwork, results.PrivateNetwork...)
	r.TotalCount += uint32(len(results.PrivateNetwork))
	return uint32(len(results.PrivateNetwork)), nil
}

type AttachPrivateNetworkRequest struct {
	Region scw.Region `json:"-"`
	// LbID: load Balancer ID
	LbID string `json:"-"`
	// PrivateNetworkID: set your instance private network id
	PrivateNetworkID string `json:"-"`
	// IPAddress: define two local ip address of your choice for each Load Balancer instance
	IPAddress []string `json:"ip_address"`
}

// AttachPrivateNetwork: bETA - Add Load Balancer on instance private network
func (s *API) AttachPrivateNetwork(req *AttachPrivateNetworkRequest, opts ...scw.RequestOption) (*PrivateNetwork, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return nil, errors.New("field LbID cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return nil, errors.New("field PrivateNetworkID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "/attach",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp PrivateNetwork

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DetachPrivateNetworkRequest struct {
	Region scw.Region `json:"-"`

	PrivateNetworkID string `json:"-"`

	LbID string `json:"-"`
}

// DetachPrivateNetwork: bETA - Remove Load Balancer of private network
func (s *API) DetachPrivateNetwork(req *DetachPrivateNetworkRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PrivateNetworkID) == "" {
		return errors.New("field PrivateNetworkID cannot be empty in request")
	}

	if fmt.Sprint(req.LbID) == "" {
		return errors.New("field LbID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/lb/v1/regions/" + fmt.Sprint(req.Region) + "/lbs/" + fmt.Sprint(req.LbID) + "/private-networks/" + fmt.Sprint(req.PrivateNetworkID) + "/detach",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}
