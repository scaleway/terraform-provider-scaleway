// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

package instance

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
)

// API instance API
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type Arch string

const (
	// ArchX86_64 is [insert doc].
	ArchX86_64 = Arch("x86_64")
	// ArchArm is [insert doc].
	ArchArm = Arch("arm")
)

func (enum Arch) String() string {
	if enum == "" {
		// return default value if empty
		return "x86_64"
	}
	return string(enum)
}

type ComputeClusterPolicyMode string

const (
	// ComputeClusterPolicyModeOptional is [insert doc].
	ComputeClusterPolicyModeOptional = ComputeClusterPolicyMode("optional")
	// ComputeClusterPolicyModeEnforced is [insert doc].
	ComputeClusterPolicyModeEnforced = ComputeClusterPolicyMode("enforced")
)

func (enum ComputeClusterPolicyMode) String() string {
	if enum == "" {
		// return default value if empty
		return "optional"
	}
	return string(enum)
}

type ComputeClusterPolicyType string

const (
	// ComputeClusterPolicyTypeLowLatency is [insert doc].
	ComputeClusterPolicyTypeLowLatency = ComputeClusterPolicyType("low_latency")
	// ComputeClusterPolicyTypeMaxAvailability is [insert doc].
	ComputeClusterPolicyTypeMaxAvailability = ComputeClusterPolicyType("max_availability")
)

func (enum ComputeClusterPolicyType) String() string {
	if enum == "" {
		// return default value if empty
		return "low_latency"
	}
	return string(enum)
}

type GetServerTypesAvailabilityResponseAvailability string

const (
	// GetServerTypesAvailabilityResponseAvailabilityAvailable is [insert doc].
	GetServerTypesAvailabilityResponseAvailabilityAvailable = GetServerTypesAvailabilityResponseAvailability("available")
	// GetServerTypesAvailabilityResponseAvailabilityScarce is [insert doc].
	GetServerTypesAvailabilityResponseAvailabilityScarce = GetServerTypesAvailabilityResponseAvailability("scarce")
	// GetServerTypesAvailabilityResponseAvailabilityShortage is [insert doc].
	GetServerTypesAvailabilityResponseAvailabilityShortage = GetServerTypesAvailabilityResponseAvailability("shortage")
)

func (enum GetServerTypesAvailabilityResponseAvailability) String() string {
	if enum == "" {
		// return default value if empty
		return "available"
	}
	return string(enum)
}

type ImageState string

const (
	// ImageStateAvailable is [insert doc].
	ImageStateAvailable = ImageState("available")
	// ImageStateCreating is [insert doc].
	ImageStateCreating = ImageState("creating")
	// ImageStateError is [insert doc].
	ImageStateError = ImageState("error")
)

func (enum ImageState) String() string {
	if enum == "" {
		// return default value if empty
		return "available"
	}
	return string(enum)
}

type SecurityGroupPolicy string

const (
	// SecurityGroupPolicyAccept is [insert doc].
	SecurityGroupPolicyAccept = SecurityGroupPolicy("accept")
	// SecurityGroupPolicyDrop is [insert doc].
	SecurityGroupPolicyDrop = SecurityGroupPolicy("drop")
)

func (enum SecurityGroupPolicy) String() string {
	if enum == "" {
		// return default value if empty
		return "accept"
	}
	return string(enum)
}

type SecurityGroupRuleAction string

const (
	// SecurityGroupRuleActionAccept is [insert doc].
	SecurityGroupRuleActionAccept = SecurityGroupRuleAction("accept")
	// SecurityGroupRuleActionDrop is [insert doc].
	SecurityGroupRuleActionDrop = SecurityGroupRuleAction("drop")
)

func (enum SecurityGroupRuleAction) String() string {
	if enum == "" {
		// return default value if empty
		return "accept"
	}
	return string(enum)
}

type SecurityGroupRuleDirection string

const (
	// SecurityGroupRuleDirectionInbound is [insert doc].
	SecurityGroupRuleDirectionInbound = SecurityGroupRuleDirection("inbound")
	// SecurityGroupRuleDirectionOutbound is [insert doc].
	SecurityGroupRuleDirectionOutbound = SecurityGroupRuleDirection("outbound")
)

func (enum SecurityGroupRuleDirection) String() string {
	if enum == "" {
		// return default value if empty
		return "inbound"
	}
	return string(enum)
}

type SecurityGroupRuleProtocol string

const (
	// SecurityGroupRuleProtocolTCP is [insert doc].
	SecurityGroupRuleProtocolTCP = SecurityGroupRuleProtocol("TCP")
	// SecurityGroupRuleProtocolUDP is [insert doc].
	SecurityGroupRuleProtocolUDP = SecurityGroupRuleProtocol("UDP")
	// SecurityGroupRuleProtocolICMP is [insert doc].
	SecurityGroupRuleProtocolICMP = SecurityGroupRuleProtocol("ICMP")
)

func (enum SecurityGroupRuleProtocol) String() string {
	if enum == "" {
		// return default value if empty
		return "TCP"
	}
	return string(enum)
}

type ServerAction string

const (
	// ServerActionPoweron is [insert doc].
	ServerActionPoweron = ServerAction("poweron")
	// ServerActionBackup is [insert doc].
	ServerActionBackup = ServerAction("backup")
	// ServerActionStopInPlace is [insert doc].
	ServerActionStopInPlace = ServerAction("stop_in_place")
	// ServerActionPoweroff is [insert doc].
	ServerActionPoweroff = ServerAction("poweroff")
	// ServerActionTerminate is [insert doc].
	ServerActionTerminate = ServerAction("terminate")
	// ServerActionReboot is [insert doc].
	ServerActionReboot = ServerAction("reboot")
)

func (enum ServerAction) String() string {
	if enum == "" {
		// return default value if empty
		return "poweron"
	}
	return string(enum)
}

type ServerBootType string

const (
	// ServerBootTypeLocal is [insert doc].
	ServerBootTypeLocal = ServerBootType("local")
)

func (enum ServerBootType) String() string {
	if enum == "" {
		// return default value if empty
		return "local"
	}
	return string(enum)
}

type ServerState string

const (
	// ServerStateRunning is [insert doc].
	ServerStateRunning = ServerState("running")
	// ServerStateStopped is [insert doc].
	ServerStateStopped = ServerState("stopped")
	// ServerStateStoppedInPlace is [insert doc].
	ServerStateStoppedInPlace = ServerState("stopped in place")
	// ServerStateStarting is [insert doc].
	ServerStateStarting = ServerState("starting")
	// ServerStateStopping is [insert doc].
	ServerStateStopping = ServerState("stopping")
	// ServerStateLocked is [insert doc].
	ServerStateLocked = ServerState("locked")
)

func (enum ServerState) String() string {
	if enum == "" {
		// return default value if empty
		return "running"
	}
	return string(enum)
}

type SnapshotState string

const (
	// SnapshotStateAvailable is [insert doc].
	SnapshotStateAvailable = SnapshotState("available")
	// SnapshotStateSnapshotting is [insert doc].
	SnapshotStateSnapshotting = SnapshotState("snapshotting")
	// SnapshotStateError is [insert doc].
	SnapshotStateError = SnapshotState("error")
)

func (enum SnapshotState) String() string {
	if enum == "" {
		// return default value if empty
		return "available"
	}
	return string(enum)
}

type TaskStatus string

const (
	// TaskStatusPending is [insert doc].
	TaskStatusPending = TaskStatus("pending")
	// TaskStatusStarted is [insert doc].
	TaskStatusStarted = TaskStatus("started")
	// TaskStatusSuccess is [insert doc].
	TaskStatusSuccess = TaskStatus("success")
	// TaskStatusFailure is [insert doc].
	TaskStatusFailure = TaskStatus("failure")
	// TaskStatusRetry is [insert doc].
	TaskStatusRetry = TaskStatus("retry")
)

func (enum TaskStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "pending"
	}
	return string(enum)
}

type VolumeState string

const (
	// VolumeStateAvailable is [insert doc].
	VolumeStateAvailable = VolumeState("available")
	// VolumeStateSnapshotting is [insert doc].
	VolumeStateSnapshotting = VolumeState("snapshotting")
	// VolumeStateError is [insert doc].
	VolumeStateError = VolumeState("error")
)

func (enum VolumeState) String() string {
	if enum == "" {
		// return default value if empty
		return "available"
	}
	return string(enum)
}

type VolumeType string

const (
	// VolumeTypeLSSD is [insert doc].
	VolumeTypeLSSD = VolumeType("l_ssd")
	// VolumeTypeLHdd is [insert doc].
	VolumeTypeLHdd = VolumeType("l_hdd")
	// VolumeTypeBSSD is [insert doc].
	VolumeTypeBSSD = VolumeType("b_ssd")
)

func (enum VolumeType) String() string {
	if enum == "" {
		// return default value if empty
		return "l_ssd"
	}
	return string(enum)
}

type Bootscript struct {
	// Arch display the bootscripts arch
	//
	// Default value: x86_64
	Arch Arch `json:"arch,omitempty"`
	// Bootcmdargs display the bootscript parameters
	Bootcmdargs string `json:"bootcmdargs,omitempty"`
	// Default dispmay if the bootscript is the default bootscript if no other boot option is configured
	Default bool `json:"default,omitempty"`
	// Dtb provide information regarding a Device Tree Binary (dtb) for use with C1 servers
	Dtb string `json:"dtb,omitempty"`
	// ID display the bootscripts ID
	ID string `json:"id,omitempty"`
	// Initrd display the initrd (initial ramdisk) configuration
	Initrd string `json:"initrd,omitempty"`
	// Kernel display the server kernel version
	Kernel string `json:"kernel,omitempty"`
	// Organization display the bootscripts organization
	Organization string `json:"organization,omitempty"`
	// Public provide information if the bootscript is public
	Public bool `json:"public,omitempty"`
	// Title display the bootscripts title
	Title string `json:"title,omitempty"`
}

type ComputeCluster struct {
	// ID display compute-cluster unique ID
	ID string `json:"id,omitempty"`
	// Name display compute-cluster name
	Name string `json:"name,omitempty"`
	// Organization display compute-cluster organization
	Organization string `json:"organization,omitempty"`
	// PolicyMode select the failling mode when the placement cannot be  respected, either optional or enforced
	//
	// Default value: optional
	PolicyMode ComputeClusterPolicyMode `json:"policy_mode,omitempty"`
	// PolicyType select the behavior of the compute-cluster, either low_latency (group) or max_availability (spread)
	//
	// Default value: low_latency
	PolicyType ComputeClusterPolicyType `json:"policy_type,omitempty"`
	// PolicyRespected indicate if the selected policy is respected. Returns true if the policy is respected, false otherwise
	PolicyRespected bool `json:"policy_respected,omitempty"`
}

type ComputeClusterServer struct {
	ID string `json:"id,omitempty"`

	Name string `json:"name,omitempty"`

	PolicyRespected bool `json:"policy_respected,omitempty"`
}

type CreateComputeClusterResponse struct {
	ComputeCluster *ComputeCluster `json:"compute_cluster,omitempty"`
}

type CreateIPResponse struct {
	IP *IP `json:"ip,omitempty"`

	Location string `json:"Location,omitempty"`
}

type CreateImageResponse struct {
	Image *Image `json:"image,omitempty"`

	Location string `json:"Location,omitempty"`
}

type CreateSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup `json:"security_group,omitempty"`
}

type CreateSecurityGroupRuleResponse struct {
	Rule *SecurityGroupRule `json:"rule,omitempty"`
}

type CreateServerResponse struct {
	Server *Server `json:"server,omitempty"`
}

type CreateSnapshotResponse struct {
	Snapshot *Snapshot `json:"snapshot,omitempty"`
}

type CreateVolumeResponse struct {
	Volume *Volume `json:"volume,omitempty"`

	Location string `json:"Location,omitempty"`
}

type Dashboard struct {
	VolumesCount uint32 `json:"volumes_count,omitempty"`

	RunningServersCount uint32 `json:"running_servers_count,omitempty"`

	ServersByTypes map[string]uint32 `json:"servers_by_types,omitempty"`

	ImagesCount uint32 `json:"images_count,omitempty"`

	SnapshotsCount uint32 `json:"snapshots_count,omitempty"`

	ServersCount uint32 `json:"servers_count,omitempty"`

	IpsCount uint32 `json:"ips_count,omitempty"`

	SecurityGroupsCount uint32 `json:"security_groups_count,omitempty"`

	IpsUnused uint32 `json:"ips_unused,omitempty"`
}

type GetBootscriptResponse struct {
	Bootscript *Bootscript `json:"bootscript,omitempty"`
}

type GetComputeClusterResponse struct {
	ComputeCluster *ComputeCluster `json:"compute_cluster,omitempty"`
}

type GetComputeClusterServersResponse struct {
	Servers []*ComputeClusterServer `json:"servers,omitempty"`
}

type GetDashboardResponse struct {
	Dashboard *Dashboard `json:"dashboard,omitempty"`
}

type GetIPResponse struct {
	IP *IP `json:"ip,omitempty"`
}

type GetImageResponse struct {
	Image *Image `json:"image,omitempty"`
}

type GetSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup `json:"security_group,omitempty"`
}

type GetSecurityGroupRuleResponse struct {
	Rule *SecurityGroupRule `json:"rule,omitempty"`
}

type GetServerResponse struct {
	Server *Server `json:"server,omitempty"`
}

type GetServerTypesAvailabilityResponse struct {
	Servers map[string]GetServerTypesAvailabilityResponseAvailability `json:"servers,omitempty"`
}

type GetSnapshotResponse struct {
	Snapshot *Snapshot `json:"snapshot,omitempty"`
}

type GetVolumeResponse struct {
	Volume *Volume `json:"volume,omitempty"`
}

type IP struct {
	ID string `json:"id,omitempty"`

	Address net.IP `json:"address,omitempty"`

	Reverse *string `json:"reverse,omitempty"`

	Server *ServerSummary `json:"server,omitempty"`

	Organization string `json:"organization,omitempty"`
}

type Image struct {
	ID string `json:"id,omitempty"`

	Name string `json:"name,omitempty"`
	// Arch
	//
	// Default value: x86_64
	Arch Arch `json:"arch,omitempty"`

	CreationDate time.Time `json:"creation_date,omitempty"`

	ModificationDate time.Time `json:"modification_date,omitempty"`

	DefaultBootscript *Bootscript `json:"default_bootscript,omitempty"`

	ExtraVolumes map[string]*Volume `json:"extra_volumes,omitempty"`

	FromServer *ServerSummary `json:"from_server,omitempty"`

	Organization string `json:"organization,omitempty"`

	Public bool `json:"public,omitempty"`

	RootVolume *VolumeTemplate `json:"root_volume,omitempty"`
	// State
	//
	// Default value: available
	State ImageState `json:"state,omitempty"`
}

type ListBootscriptsResponse struct {
	Bootscripts []*Bootscript `json:"bootscripts,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type ListComputeClustersResponse struct {
	ComputeClusters []*ComputeCluster `json:"compute_clusters,omitempty"`
}

type ListImagesResponse struct {
	Images []*Image `json:"images,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type ListIpsResponse struct {
	Ips []*IP `json:"ips,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type ListSecurityGroupRulesResponse struct {
	Rules []*SecurityGroupRule `json:"rules,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type ListSecurityGroupsResponse struct {
	SecurityGroups []*SecurityGroup `json:"security_groups,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type ListServerActionsResponse struct {
	Actions []ServerAction `json:"actions,omitempty"`
}

type ListServerUserDataResponse struct {
	UserData []string `json:"user_data,omitempty"`
}

type ListServersResponse struct {
	Servers []*Server `json:"servers,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type ListServersTypesResponse struct {
	Servers map[string]*ServerType `json:"servers,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type ListSnapshotsResponse struct {
	Snapshots []*Snapshot `json:"snapshots,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type ListVolumesResponse struct {
	Volumes []*Volume `json:"volumes,omitempty"`

	TotalCount uint32 `json:"total_count,omitempty"`
}

type NullableStringValue struct {
	Null bool `json:"null,omitempty"`

	Value string `json:"value,omitempty"`
}

type SecurityGroup struct {
	// ID display the security groups' unique ID
	ID string `json:"id,omitempty"`
	// Name display the security groups name
	Name string `json:"name,omitempty"`
	// CreationDate display the security group creation date
	CreationDate time.Time `json:"creation_date,omitempty"`
	// ModificationDate display the security group modification date
	ModificationDate time.Time `json:"modification_date,omitempty"`
	// Description display the security groups description
	Description string `json:"description,omitempty"`
	// EnableDefaultSecurity display if the security group is set as default
	EnableDefaultSecurity bool `json:"enable_default_security,omitempty"`
	// InboundDefaultPolicy display the default inbound policy
	//
	// Default value: accept
	InboundDefaultPolicy SecurityGroupPolicy `json:"inbound_default_policy,omitempty"`
	// Organization display the security groups organization ID
	Organization string `json:"organization,omitempty"`
	// OrganizationDefault display if the security group is set as organization default
	OrganizationDefault bool `json:"organization_default,omitempty"`
	// OutboundDefaultPolicy display the default outbound policy
	//
	// Default value: accept
	OutboundDefaultPolicy SecurityGroupPolicy `json:"outbound_default_policy,omitempty"`
	// Servers list of servers attached to this security group
	Servers []*ServerSummary `json:"servers,omitempty"`
	// Stateful true if the security group is stateful
	Stateful bool `json:"stateful,omitempty"`
}

type SecurityGroupRule struct {
	ID string `json:"id,omitempty"`
	// Protocol
	//
	// Default value: TCP
	Protocol SecurityGroupRuleProtocol `json:"protocol,omitempty"`
	// Direction
	//
	// Default value: inbound
	Direction SecurityGroupRuleDirection `json:"direction,omitempty"`
	// Action
	//
	// Default value: accept
	Action SecurityGroupRuleAction `json:"action,omitempty"`

	IPRange string `json:"ip_range,omitempty"`

	DestPortFrom *uint32 `json:"dest_port_from,omitempty"`

	DestPortTo *uint32 `json:"dest_port_to,omitempty"`

	Position uint32 `json:"position,omitempty"`

	Editable bool `json:"editable,omitempty"`
}

type SecurityGroupSummary struct {
	ID string `json:"id,omitempty"`

	Name string `json:"name,omitempty"`
}

type Server struct {
	// ID display the server unique ID
	ID string `json:"id,omitempty"`
	// Image provide information on the server image
	Image *Image `json:"image,omitempty"`
	// Name display the server name
	Name string `json:"name,omitempty"`
	// Organization display the server organization
	Organization string `json:"organization,omitempty"`
	// PrivateIP display the server private IP address
	PrivateIP *string `json:"private_ip,omitempty"`
	// PublicIP display the server public IP address
	PublicIP *ServerIP `json:"public_ip,omitempty"`
	// State display the server state
	//
	// Default value: running
	State ServerState `json:"state,omitempty"`
	// BootType display the server boot type
	//
	// Default value: local
	BootType ServerBootType `json:"boot_type,omitempty"`
	// Tags display the server associated tags
	Tags []string `json:"tags,omitempty"`
	// Volumes display the server volumes
	Volumes map[string]*Volume `json:"volumes,omitempty"`
	// Bootscript display the server bootscript
	Bootscript *Bootscript `json:"bootscript,omitempty"`
	// CommercialType display the server commercial type (e.g. GP1-M)
	CommercialType string `json:"commercial_type,omitempty"`
	// CreationDate display the server creation date
	CreationDate time.Time `json:"creation_date,omitempty"`
	// DynamicIPRequired display if a dynamic IP is required
	DynamicIPRequired bool `json:"dynamic_ip_required,omitempty"`
	// EnableIPv6 display if IPv6 is enabled
	EnableIPv6 bool `json:"enable_ipv6,omitempty"`
	// ExtraNetworks display information about additional network interfaces
	ExtraNetworks []string `json:"extra_networks,omitempty"`
	// Hostname display the server host name
	Hostname string `json:"hostname,omitempty"`
	// AllowedActions provide as list of allowed actions on the server
	AllowedActions []ServerAction `json:"allowed_actions,omitempty"`
	// Arch display the server arch
	//
	// Default value: x86_64
	Arch Arch `json:"arch,omitempty"`
	// IPv6 display the server IPv6 address
	IPv6 *ServerIPv6 `json:"ipv6,omitempty"`
	// Location display the server location
	Location *ServerLocation `json:"location,omitempty"`
	// Maintenances display the server planned maintenances
	Maintenances []*ServerMaintenance `json:"maintenances,omitempty"`
	// ModificationDate display the server modification date
	ModificationDate time.Time `json:"modification_date,omitempty"`
	// Protected display the server protection option is activated
	Protected bool `json:"protected,omitempty"`
	// SecurityGroup display the server security group
	SecurityGroup *SecurityGroupSummary `json:"security_group,omitempty"`
	// StateDetail display the server state_detail
	StateDetail string `json:"state_detail,omitempty"`
	// ComputeCluster display the server ComputeCluster
	ComputeCluster *ComputeCluster `json:"compute_cluster,omitempty"`
}

type ServerActionResponse struct {
	Task *Task `json:"task,omitempty"`
}

type ServerIP struct {
	// ID display the unique ID of the IP address
	ID string `json:"id,omitempty"`
	// Address display the server public IPv4 IP-Address
	Address net.IP `json:"address,omitempty"`
	// Dynamic display information if the IP address will be considered as dynamic
	Dynamic bool `json:"dynamic,omitempty"`
}

type ServerIPv6 struct {
	// Address display the server IPv6 IP-Address
	Address net.IP `json:"address,omitempty"`
	// Gateway display the IPv6 IP-addresses gateway
	Gateway string `json:"gateway,omitempty"`
	// Netmask display the IPv6 IP-addresses CIDR netmask
	Netmask string `json:"netmask,omitempty"`
}

type ServerLocation struct {
	ClusterID string `json:"cluster_id,omitempty"`

	HypervisorID string `json:"hypervisor_id,omitempty"`

	NodeID string `json:"node_id,omitempty"`

	PlatformID string `json:"platform_id,omitempty"`

	ZoneID string `json:"zone_id,omitempty"`
}

type ServerMaintenance struct {
}

type ServerSummary struct {
	ID string `json:"id,omitempty"`

	Name string `json:"name,omitempty"`
}

type ServerType struct {
	MonthlyPrice float32 `json:"monthly_price,omitempty"`

	HourlyPrice float32 `json:"hourly_price,omitempty"`

	AltNames []string `json:"alt_names,omitempty"`

	PerVolumeConstraint *ServerTypeVolumeConstraintsByType `json:"per_volume_constraint,omitempty"`

	VolumesConstraint *ServerTypeVolumeConstraintSizes `json:"volumes_constraint,omitempty"`

	Ncpus uint32 `json:"ncpus,omitempty"`

	Gpu *uint64 `json:"gpu,omitempty"`

	RAM uint64 `json:"ram,omitempty"`
	// Arch
	//
	// Default value: x86_64
	Arch Arch `json:"arch,omitempty"`

	Baremetal bool `json:"baremetal,omitempty"`

	Network *ServerTypeNetwork `json:"network,omitempty"`
}

type ServerTypeNetwork struct {
	Interfaces []*ServerTypeNetworkInterface `json:"interfaces,omitempty"`

	SumInternalBandwidth *uint64 `json:"sum_internal_bandwidth,omitempty"`

	SumInternetBandwidth *uint64 `json:"sum_internet_bandwidth,omitempty"`

	IPv6Support bool `json:"ipv6_support,omitempty"`
}

type ServerTypeNetworkInterface struct {
	InternalBandwidth *uint64 `json:"internal_bandwidth,omitempty"`

	InternetBandwidth *uint64 `json:"internet_bandwidth,omitempty"`
}

type ServerTypeVolumeConstraintSizes struct {
	MinSize uint64 `json:"min_size,omitempty"`

	MaxSize uint64 `json:"max_size,omitempty"`
}

type ServerTypeVolumeConstraintsByType struct {
	LSSD *ServerTypeVolumeConstraintSizes `json:"l_ssd,omitempty"`
}

type SetComputeClusterResponse struct {
	ComputeClusterID string `json:"compute_cluster_id,omitempty"`
}

type SetComputeClusterServersResponse struct {
	Servers []*ComputeClusterServer `json:"servers,omitempty"`
}

type Snapshot struct {
	ID string `json:"id,omitempty"`

	Name string `json:"name,omitempty"`

	Organization string `json:"organization,omitempty"`
	// VolumeType
	//
	// Default value: l_ssd
	VolumeType VolumeType `json:"volume_type,omitempty"`

	Size uint64 `json:"size,omitempty"`
	// State
	//
	// Default value: available
	State SnapshotState `json:"state,omitempty"`

	BaseVolume *SnapshotBaseVolume `json:"base_volume,omitempty"`

	CreationDate time.Time `json:"creation_date,omitempty"`

	ModificationDate time.Time `json:"modification_date,omitempty"`
}

type SnapshotBaseVolume struct {
	ID string `json:"id,omitempty"`

	Name string `json:"name,omitempty"`
}

type Task struct {
	// ID the unique ID of the task
	ID string `json:"id,omitempty"`
	// Description the description of the task
	Description string `json:"description,omitempty"`

	HrefFrom string `json:"href_from,omitempty"`

	HrefResult string `json:"href_result,omitempty"`
	// Progress show the progress of the task in percent
	Progress int32 `json:"progress,omitempty"`
	// StartedAt display the task start date
	StartedAt time.Time `json:"started_at,omitempty"`
	// Status display the task status
	//
	// Default value: pending
	Status TaskStatus `json:"status,omitempty"`
	// TerminatedAt display the task end date
	TerminatedAt time.Time `json:"terminated_at,omitempty"`
}

type UpdateComputeClusterResponse struct {
	ComputeClusterID string `json:"compute_cluster_id,omitempty"`
}

type UpdateComputeClusterServersResponse struct {
	Servers []*ComputeClusterServer `json:"servers,omitempty"`
}

type UpdateIPResponse struct {
	IP *IP `json:"ip,omitempty"`
}

type UpdateServerResponse struct {
	Server *Server `json:"server,omitempty"`
}

type Volume struct {
	// ID display the volumes unique ID
	ID string `json:"id,omitempty"`
	// Name display the volumes names
	Name string `json:"name,omitempty"`
	// ExportURI show the volumes NBD export URI
	ExportURI string `json:"export_uri,omitempty"`
	// Organization display the volumes organization
	Organization string `json:"organization,omitempty"`
	// Server display information about the server attached to the volume
	Server *ServerSummary `json:"server,omitempty"`
	// Size display the volumes disk size
	Size uint64 `json:"size,omitempty"`
	// VolumeType display the volumes type
	//
	// Default value: l_ssd
	VolumeType VolumeType `json:"volume_type,omitempty"`
	// CreationDate display the volumes creation date
	CreationDate time.Time `json:"creation_date,omitempty"`
	// ModificationDate display the volumes modification date
	ModificationDate time.Time `json:"modification_date,omitempty"`
	// State display the volumes state
	//
	// Default value: available
	State VolumeState `json:"state,omitempty"`
}

type VolumeTemplate struct {
	// ID display the volumes unique ID
	ID string `json:"id,omitempty"`
	// Name display the volumes name
	Name string `json:"name,omitempty"`
	// Size display the volumes disk size
	Size uint64 `json:"size,omitempty"`
	// VolumeType display the volumes type
	//
	// Default value: l_ssd
	VolumeType VolumeType `json:"volume_type,omitempty"`
	// Organization the organization ID
	Organization string `json:"organization,omitempty"`
}

type setIPResponse struct {
	IP *IP `json:"ip,omitempty"`
}

type setImageResponse struct {
	Image *Image `json:"image,omitempty"`
}

type setSecurityGroupResponse struct {
	SecurityGroup *SecurityGroup `json:"security_group,omitempty"`
}

type setSecurityGroupRuleResponse struct {
	Rule *SecurityGroupRule `json:"rule,omitempty"`
}

type setServerResponse struct {
	Server *Server `json:"server,omitempty"`
}

type setSnapshotResponse struct {
	Snapshot *Snapshot `json:"snapshot,omitempty"`
}

// Service API

type GetServerTypesAvailabilityRequest struct {
	Zone scw.Zone `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`
}

// GetServerTypesAvailability get availability
//
// Get availibility for all server types
func (s *API) GetServerTypesAvailability(req *GetServerTypesAvailabilityRequest, opts ...scw.RequestOption) (*GetServerTypesAvailabilityResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/products/servers/availability",
		Query:   query,
		Headers: http.Header{},
	}

	var resp GetServerTypesAvailabilityResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListServersTypesRequest struct {
	Zone scw.Zone `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListServersTypes list server types
//
// Get server types technical details
func (s *API) ListServersTypes(req *ListServersTypesRequest, opts ...scw.RequestOption) (*ListServersTypesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/products/servers",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListServersTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListServersRequest struct {
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`

	Name *string `json:"-"`
}

// ListServers list servers
func (s *API) ListServers(req *ListServersRequest, opts ...scw.RequestOption) (*ListServersResponse, error) {
	var err error

	defaultOrganization, exist := s.client.GetDefaultProjectID()
	if (req.Organization == nil || *req.Organization == "") && exist {
		req.Organization = &defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers",
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
func (r *ListServersResponse) UnsafeGetTotalCount() int {
	return int(r.TotalCount)
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListServersResponse) UnsafeAppend(res interface{}) (int, scw.SdkError) {
	results, ok := res.(*ListServersResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Servers = append(r.Servers, results.Servers...)
	r.TotalCount += uint32(len(results.Servers))
	return len(results.Servers), nil
}

type CreateServerRequest struct {
	Zone scw.Zone `json:"-"`
	// Name display the server name
	Name string `json:"name,omitempty"`
	// DynamicIPRequired define if a dynamic IP is required for the instance
	DynamicIPRequired *bool `json:"dynamic_ip_required,omitempty"`
	// CommercialType define the server commercial type (i.e. GP1-S)
	CommercialType string `json:"commercial_type,omitempty"`
	// Image define the server image id
	Image string `json:"image,omitempty"`
	// Volumes define the volumes attached to the server
	Volumes map[string]*VolumeTemplate `json:"volumes,omitempty"`
	// EnableIPv6 define if IPv6 is enabled on the server
	EnableIPv6 bool `json:"enable_ipv6,omitempty"`
	// PublicIP define the public IPv4 attached to the server
	PublicIP string `json:"public_ip,omitempty"`
	// BootType define the boot type you want to use
	//
	// Default value: local
	BootType ServerBootType `json:"boot_type,omitempty"`
	// Organization define the server organization
	Organization string `json:"organization,omitempty"`
	// Tags define the server tags
	Tags []string `json:"tags,omitempty"`
	// SecurityGroup define the security group id
	SecurityGroup string `json:"security_group,omitempty"`
	// ComputeCluster computeCluster key if server must be part of a ComputeCluster
	ComputeCluster string `json:"compute_cluster,omitempty"`
}

// CreateServer create server
func (s *API) CreateServer(req *CreateServerRequest, opts ...scw.RequestOption) (*CreateServerResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateServerResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteServerRequest struct {
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`
}

// DeleteServer delete server
//
// Delete a server with the given id
func (s *API) DeleteServer(req *DeleteServerRequest, opts ...scw.RequestOption) error {
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type GetServerRequest struct {
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`
}

// GetServer get server
//
// Get the details of a specified Server
func (s *API) GetServer(req *GetServerRequest, opts ...scw.RequestOption) (*GetServerResponse, error) {
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
		Headers: http.Header{},
	}

	var resp GetServerResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type setServerRequest struct {
	Zone scw.Zone `json:"-"`
	// ID display the server unique ID
	ID string `json:"-"`
	// Name display the server name
	Name string `json:"name"`
	// Organization display the server organization
	Organization string `json:"organization"`
	// AllowedActions provide as list of allowed actions on the server
	AllowedActions []ServerAction `json:"allowed_actions"`
	// Tags display the server associated tags
	Tags []string `json:"tags"`
	// CommercialType display the server commercial type (e.g. GP1-M)
	CommercialType string `json:"commercial_type"`
	// CreationDate display the server creation date
	CreationDate time.Time `json:"creation_date"`
	// DynamicIPRequired display if a dynamic IP is required
	DynamicIPRequired bool `json:"dynamic_ip_required"`
	// EnableIPv6 display if IPv6 is enabled
	EnableIPv6 bool `json:"enable_ipv6"`
	// ExtraNetworks display information about additional network interfaces
	ExtraNetworks []string `json:"extra_networks"`
	// Hostname display the server host name
	Hostname string `json:"hostname"`
	// Image provide information on the server image
	Image *Image `json:"image"`
	// Protected display the server protection option is activated
	Protected bool `json:"protected"`
	// PrivateIP display the server private IP address
	PrivateIP *string `json:"private_ip"`
	// PublicIP display the server public IP address
	PublicIP *ServerIP `json:"public_ip"`
	// ModificationDate display the server modification date
	ModificationDate time.Time `json:"modification_date"`
	// State display the server state
	//
	// Default value: running
	State ServerState `json:"state"`
	// Location display the server location
	Location *ServerLocation `json:"location"`
	// IPv6 display the server IPv6 address
	IPv6 *ServerIPv6 `json:"ipv6"`
	// Bootscript display the server bootscript
	Bootscript *Bootscript `json:"bootscript"`
	// BootType display the server boot type
	//
	// Default value: local
	BootType ServerBootType `json:"boot_type"`
	// Volumes display the server volumes
	Volumes map[string]*Volume `json:"volumes"`
	// SecurityGroup display the server security group
	SecurityGroup *SecurityGroupSummary `json:"security_group"`
	// Maintenances display the server planned maintenances
	Maintenances []*ServerMaintenance `json:"maintenances"`
	// StateDetail display the server state_detail
	StateDetail string `json:"state_detail"`
	// Arch display the server arch
	//
	// Default value: x86_64
	Arch Arch `json:"arch"`
	// ComputeCluster display the server ComputeCluster
	ComputeCluster *ComputeCluster `json:"compute_cluster"`
}

func (s *API) setServer(req *setServerRequest, opts ...scw.RequestOption) (*setServerResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ID) == "" {
		return nil, errors.New("field ID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setServerResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type updateServerRequest struct {
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`

	Name *string `json:"name,omitempty"`
	// BootType
	//
	// Default value: local
	BootType ServerBootType `json:"boot_type,omitempty"`

	Tags *[]string `json:"tags,omitempty"`

	Volumes *map[string]*VolumeTemplate `json:"volumes,omitempty"`

	Bootscript *Bootscript `json:"bootscript,omitempty"`

	DynamicIPRequired *bool `json:"dynamic_ip_required,omitempty"`

	EnableIPv6 *bool `json:"enable_ipv6,omitempty"`

	ExtraNetworks *[]string `json:"extra_networks,omitempty"`

	Protected *bool `json:"protected,omitempty"`

	SecurityGroup *SecurityGroupSummary `json:"security_group,omitempty"`

	ComputeCluster *NullableStringValue `json:"compute_cluster,omitempty"`
}

// updateServer update server
func (s *API) updateServer(req *updateServerRequest, opts ...scw.RequestOption) (*UpdateServerResponse, error) {
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateServerResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListServerActionsRequest struct {
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`
}

// ListServerActions list server actions
//
// Liste all actions that can currently be performed on a server
func (s *API) ListServerActions(req *ListServerActionsRequest, opts ...scw.RequestOption) (*ListServerActionsResponse, error) {
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/action",
		Headers: http.Header{},
	}

	var resp ListServerActionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ServerActionRequest struct {
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`
	// Action
	//
	// Default value: poweron
	Action ServerAction `json:"action,omitempty"`
}

// ServerAction perform action
//
// Perform power related actions on a server
func (s *API) ServerAction(req *ServerActionRequest, opts ...scw.RequestOption) (*ServerActionResponse, error) {
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/action",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ServerActionResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListServerUserDataRequest struct {
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`
}

// ListServerUserData list user data
//
// List all user data keys register on a given server
func (s *API) ListServerUserData(req *ListServerUserDataRequest, opts ...scw.RequestOption) (*ListServerUserDataResponse, error) {
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/user_data",
		Headers: http.Header{},
	}

	var resp ListServerUserDataResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteServerUserDataRequest struct {
	Zone scw.Zone `json:"-"`

	ServerID string `json:"-"`

	Key string `json:"-"`
}

// DeleteServerUserData delete user data
//
// Delete the given key from a server user data
func (s *API) DeleteServerUserData(req *DeleteServerUserDataRequest, opts ...scw.RequestOption) error {
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

	if fmt.Sprint(req.Key) == "" {
		return errors.New("field Key cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/servers/" + fmt.Sprint(req.ServerID) + "/user_data/" + fmt.Sprint(req.Key) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListImagesRequest struct {
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`

	Name *string `json:"-"`

	Public bool `json:"-"`

	Arch *string `json:"-"`
}

// ListImages list images
//
// List all images available in an account
func (s *API) ListImages(req *ListImagesRequest, opts ...scw.RequestOption) (*ListImagesResponse, error) {
	var err error

	defaultOrganization, exist := s.client.GetDefaultProjectID()
	if (req.Organization == nil || *req.Organization == "") && exist {
		req.Organization = &defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "public", req.Public)
	parameter.AddToQuery(query, "arch", req.Arch)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListImagesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeGetTotalCount() int {
	return int(r.TotalCount)
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListImagesResponse) UnsafeAppend(res interface{}) (int, scw.SdkError) {
	results, ok := res.(*ListImagesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Images = append(r.Images, results.Images...)
	r.TotalCount += uint32(len(results.Images))
	return len(results.Images), nil
}

type GetImageRequest struct {
	Zone scw.Zone `json:"-"`

	ImageID string `json:"-"`
}

// GetImage get image
//
// Get details of an image with the given id
func (s *API) GetImage(req *GetImageRequest, opts ...scw.RequestOption) (*GetImageResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return nil, errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images/" + fmt.Sprint(req.ImageID) + "",
		Headers: http.Header{},
	}

	var resp GetImageResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateImageRequest struct {
	Zone scw.Zone `json:"-"`

	Name string `json:"name,omitempty"`

	RootVolume string `json:"root_volume,omitempty"`
	// Arch
	//
	// Default value: x86_64
	Arch Arch `json:"arch,omitempty"`

	DefaultBootscript *Bootscript `json:"default_bootscript,omitempty"`

	ExtraVolumes map[string]*Volume `json:"extra_volumes,omitempty"`

	Organization string `json:"organization,omitempty"`

	Public bool `json:"public,omitempty"`
}

// CreateImage create image
func (s *API) CreateImage(req *CreateImageRequest, opts ...scw.RequestOption) (*CreateImageResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateImageResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type setImageRequest struct {
	Zone scw.Zone `json:"-"`

	ID string `json:"-"`

	Name string `json:"name"`
	// Arch
	//
	// Default value: x86_64
	Arch Arch `json:"arch"`

	CreationDate time.Time `json:"creation_date"`

	ModificationDate time.Time `json:"modification_date"`

	DefaultBootscript *Bootscript `json:"default_bootscript"`

	ExtraVolumes map[string]*Volume `json:"extra_volumes"`

	FromServer *ServerSummary `json:"from_server"`

	Organization string `json:"organization"`

	Public bool `json:"public"`

	RootVolume *VolumeTemplate `json:"root_volume"`
	// State
	//
	// Default value: available
	State ImageState `json:"state"`
}

// setImage update image
//
// Replace all image properties with an image message
func (s *API) setImage(req *setImageRequest, opts ...scw.RequestOption) (*setImageResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ID) == "" {
		return nil, errors.New("field ID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images/" + fmt.Sprint(req.ID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setImageResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteImageRequest struct {
	Zone scw.Zone `json:"-"`

	ImageID string `json:"-"`
}

// DeleteImage delete image
//
// Delete the image with the given id
func (s *API) DeleteImage(req *DeleteImageRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ImageID) == "" {
		return errors.New("field ImageID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/images/" + fmt.Sprint(req.ImageID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListSnapshotsRequest struct {
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`

	Name *string `json:"-"`
}

// ListSnapshots list snapshots
func (s *API) ListSnapshots(req *ListSnapshotsRequest, opts ...scw.RequestOption) (*ListSnapshotsResponse, error) {
	var err error

	defaultOrganization, exist := s.client.GetDefaultProjectID()
	if (req.Organization == nil || *req.Organization == "") && exist {
		req.Organization = &defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSnapshotsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSnapshotsResponse) UnsafeGetTotalCount() int {
	return int(r.TotalCount)
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSnapshotsResponse) UnsafeAppend(res interface{}) (int, scw.SdkError) {
	results, ok := res.(*ListSnapshotsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Snapshots = append(r.Snapshots, results.Snapshots...)
	r.TotalCount += uint32(len(results.Snapshots))
	return len(results.Snapshots), nil
}

type CreateSnapshotRequest struct {
	Zone scw.Zone `json:"-"`

	VolumeID string `json:"volume_id,omitempty"`

	Organization string `json:"organization,omitempty"`

	Name string `json:"name,omitempty"`
}

// CreateSnapshot create snapshot
func (s *API) CreateSnapshot(req *CreateSnapshotRequest, opts ...scw.RequestOption) (*CreateSnapshotResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateSnapshotResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetSnapshotRequest struct {
	Zone scw.Zone `json:"-"`

	SnapshotID string `json:"-"`
}

// GetSnapshot get snapshot
//
// Get details of a snapshot with the given id
func (s *API) GetSnapshot(req *GetSnapshotRequest, opts ...scw.RequestOption) (*GetSnapshotResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SnapshotID) == "" {
		return nil, errors.New("field SnapshotID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
		Headers: http.Header{},
	}

	var resp GetSnapshotResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type setSnapshotRequest struct {
	Zone scw.Zone `json:"-"`

	ID string `json:"-"`

	Name string `json:"name"`

	Organization string `json:"organization"`
	// VolumeType
	//
	// Default value: l_ssd
	VolumeType VolumeType `json:"volume_type"`

	Size uint64 `json:"size"`
	// State
	//
	// Default value: available
	State SnapshotState `json:"state"`

	BaseVolume *SnapshotBaseVolume `json:"base_volume"`

	CreationDate time.Time `json:"creation_date"`

	ModificationDate time.Time `json:"modification_date"`
}

// setSnapshot update snapshot
//
// Replace all snapshot properties with a snapshot message
func (s *API) setSnapshot(req *setSnapshotRequest, opts ...scw.RequestOption) (*setSnapshotResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ID) == "" {
		return nil, errors.New("field ID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.ID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setSnapshotResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteSnapshotRequest struct {
	Zone scw.Zone `json:"-"`

	SnapshotID string `json:"-"`
}

// DeleteSnapshot delete snapshot
//
// Delete the snapshot with the given id
func (s *API) DeleteSnapshot(req *DeleteSnapshotRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SnapshotID) == "" {
		return errors.New("field SnapshotID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/snapshots/" + fmt.Sprint(req.SnapshotID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListVolumesRequest struct {
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`

	Name *string `json:"-"`
}

// ListVolumes list volumes
func (s *API) ListVolumes(req *ListVolumesRequest, opts ...scw.RequestOption) (*ListVolumesResponse, error) {
	var err error

	defaultOrganization, exist := s.client.GetDefaultProjectID()
	if (req.Organization == nil || *req.Organization == "") && exist {
		req.Organization = &defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListVolumesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListVolumesResponse) UnsafeGetTotalCount() int {
	return int(r.TotalCount)
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListVolumesResponse) UnsafeAppend(res interface{}) (int, scw.SdkError) {
	results, ok := res.(*ListVolumesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Volumes = append(r.Volumes, results.Volumes...)
	r.TotalCount += uint32(len(results.Volumes))
	return len(results.Volumes), nil
}

type CreateVolumeRequest struct {
	Zone scw.Zone `json:"-"`

	Name string `json:"name,omitempty"`

	Organization string `json:"organization,omitempty"`
	// VolumeType
	//
	// Default value: l_ssd
	VolumeType VolumeType `json:"volume_type,omitempty"`

	// Precisely one of BaseSnapshot, BaseVolume, Size must be set.
	Size *uint64 `json:"size,omitempty"`

	// Precisely one of BaseSnapshot, BaseVolume, Size must be set.
	BaseVolume *string `json:"base_volume,omitempty"`

	// Precisely one of BaseSnapshot, BaseVolume, Size must be set.
	BaseSnapshot *string `json:"base_snapshot,omitempty"`
}

func (m *CreateVolumeRequest) GetFrom() From {
	switch {
	case m.Size != nil:
		return FromSize{*m.Size}
	case m.BaseVolume != nil:
		return FromBaseVolume{*m.BaseVolume}
	case m.BaseSnapshot != nil:
		return FromBaseSnapshot{*m.BaseSnapshot}
	}
	return nil
}

// CreateVolume create volume
func (s *API) CreateVolume(req *CreateVolumeRequest, opts ...scw.RequestOption) (*CreateVolumeResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateVolumeResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetVolumeRequest struct {
	Zone scw.Zone `json:"-"`

	VolumeID string `json:"-"`
}

// GetVolume get volume
//
// Get details of a volume with the given id
func (s *API) GetVolume(req *GetVolumeRequest, opts ...scw.RequestOption) (*GetVolumeResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.VolumeID) == "" {
		return nil, errors.New("field VolumeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes/" + fmt.Sprint(req.VolumeID) + "",
		Headers: http.Header{},
	}

	var resp GetVolumeResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteVolumeRequest struct {
	Zone scw.Zone `json:"-"`

	VolumeID string `json:"-"`
}

// DeleteVolume delete volume
//
// Delete the volume with the given id
func (s *API) DeleteVolume(req *DeleteVolumeRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.VolumeID) == "" {
		return errors.New("field VolumeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/volumes/" + fmt.Sprint(req.VolumeID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListSecurityGroupsRequest struct {
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`

	Name *string `json:"-"`
}

// ListSecurityGroups list security groups
//
// List all security groups available in an account
func (s *API) ListSecurityGroups(req *ListSecurityGroupsRequest, opts ...scw.RequestOption) (*ListSecurityGroupsResponse, error) {
	var err error

	defaultOrganization, exist := s.client.GetDefaultProjectID()
	if (req.Organization == nil || *req.Organization == "") && exist {
		req.Organization = &defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "name", req.Name)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSecurityGroupsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSecurityGroupsResponse) UnsafeGetTotalCount() int {
	return int(r.TotalCount)
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSecurityGroupsResponse) UnsafeAppend(res interface{}) (int, scw.SdkError) {
	results, ok := res.(*ListSecurityGroupsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.SecurityGroups = append(r.SecurityGroups, results.SecurityGroups...)
	r.TotalCount += uint32(len(results.SecurityGroups))
	return len(results.SecurityGroups), nil
}

type CreateSecurityGroupRequest struct {
	Zone scw.Zone `json:"-"`

	Name string `json:"name,omitempty"`

	Description string `json:"description,omitempty"`

	Organization string `json:"organization,omitempty"`

	OrganizationDefault bool `json:"organization_default,omitempty"`

	Stateful bool `json:"stateful,omitempty"`
	// InboundDefaultPolicy
	//
	// Default value: accept
	InboundDefaultPolicy SecurityGroupPolicy `json:"inbound_default_policy,omitempty"`
	// OutboundDefaultPolicy
	//
	// Default value: accept
	OutboundDefaultPolicy SecurityGroupPolicy `json:"outbound_default_policy,omitempty"`
}

// CreateSecurityGroup create security group
func (s *API) CreateSecurityGroup(req *CreateSecurityGroupRequest, opts ...scw.RequestOption) (*CreateSecurityGroupResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateSecurityGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetSecurityGroupRequest struct {
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`
}

// GetSecurityGroup get security group
//
// Get the details of a Security Group with the given id
func (s *API) GetSecurityGroup(req *GetSecurityGroupRequest, opts ...scw.RequestOption) (*GetSecurityGroupResponse, error) {
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

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "",
		Headers: http.Header{},
	}

	var resp GetSecurityGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteSecurityGroupRequest struct {
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`
}

// DeleteSecurityGroup delete security group
func (s *API) DeleteSecurityGroup(req *DeleteSecurityGroupRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return errors.New("field SecurityGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type setSecurityGroupRequest struct {
	Zone scw.Zone `json:"-"`
	// ID display the security groups' unique ID
	ID string `json:"-"`
	// Name display the security groups name
	Name string `json:"name"`
	// Description display the security groups description
	Description string `json:"description"`
	// EnableDefaultSecurity display if the security group is set as default
	EnableDefaultSecurity bool `json:"enable_default_security"`
	// InboundDefaultPolicy display the default inbound policy
	//
	// Default value: accept
	InboundDefaultPolicy SecurityGroupPolicy `json:"inbound_default_policy"`
	// OutboundDefaultPolicy display the default outbound policy
	//
	// Default value: accept
	OutboundDefaultPolicy SecurityGroupPolicy `json:"outbound_default_policy"`
	// Organization display the security groups organization ID
	Organization string `json:"organization"`
	// OrganizationDefault display if the security group is set as organization default
	OrganizationDefault bool `json:"organization_default"`
	// CreationDate display the security group creation date
	CreationDate time.Time `json:"creation_date"`
	// ModificationDate display the security group modification date
	ModificationDate time.Time `json:"modification_date"`
	// Servers list of servers attached to this security group
	Servers []*ServerSummary `json:"servers"`
	// Stateful true if the security group is stateful
	Stateful bool `json:"stateful"`
}

// setSecurityGroup update security group
//
// Replace all security group properties with a security group message
func (s *API) setSecurityGroup(req *setSecurityGroupRequest, opts ...scw.RequestOption) (*setSecurityGroupResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ID) == "" {
		return nil, errors.New("field ID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.ID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setSecurityGroupResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListSecurityGroupRulesRequest struct {
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListSecurityGroupRules list rules
func (s *API) ListSecurityGroupRules(req *ListSecurityGroupRulesRequest, opts ...scw.RequestOption) (*ListSecurityGroupRulesResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return nil, errors.New("field SecurityGroupID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListSecurityGroupRulesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListSecurityGroupRulesResponse) UnsafeGetTotalCount() int {
	return int(r.TotalCount)
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListSecurityGroupRulesResponse) UnsafeAppend(res interface{}) (int, scw.SdkError) {
	results, ok := res.(*ListSecurityGroupRulesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Rules = append(r.Rules, results.Rules...)
	r.TotalCount += uint32(len(results.Rules))
	return len(results.Rules), nil
}

type CreateSecurityGroupRuleRequest struct {
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`
	// Protocol
	//
	// Default value: TCP
	Protocol SecurityGroupRuleProtocol `json:"protocol,omitempty"`
	// Direction
	//
	// Default value: inbound
	Direction SecurityGroupRuleDirection `json:"direction,omitempty"`
	// Action
	//
	// Default value: accept
	Action SecurityGroupRuleAction `json:"action,omitempty"`

	IPRange string `json:"ip_range,omitempty"`

	DestPortFrom *uint32 `json:"dest_port_from,omitempty"`

	DestPortTo *uint32 `json:"dest_port_to,omitempty"`

	Position uint32 `json:"position,omitempty"`

	Editable bool `json:"editable,omitempty"`
}

// CreateSecurityGroupRule create rule
func (s *API) CreateSecurityGroupRule(req *CreateSecurityGroupRuleRequest, opts ...scw.RequestOption) (*CreateSecurityGroupRuleResponse, error) {
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

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateSecurityGroupRuleResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteSecurityGroupRuleRequest struct {
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`

	SecurityGroupRuleID string `json:"-"`
}

// DeleteSecurityGroupRule delete rule
//
// Delete a security group rule with the given id
func (s *API) DeleteSecurityGroupRule(req *DeleteSecurityGroupRuleRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupID) == "" {
		return errors.New("field SecurityGroupID cannot be empty in request")
	}

	if fmt.Sprint(req.SecurityGroupRuleID) == "" {
		return errors.New("field SecurityGroupRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules/" + fmt.Sprint(req.SecurityGroupRuleID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type GetSecurityGroupRuleRequest struct {
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`

	SecurityGroupRuleID string `json:"-"`
}

// GetSecurityGroupRule get rule
//
// Get details of a security group rule with the given id
func (s *API) GetSecurityGroupRule(req *GetSecurityGroupRuleRequest, opts ...scw.RequestOption) (*GetSecurityGroupRuleResponse, error) {
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

	if fmt.Sprint(req.SecurityGroupRuleID) == "" {
		return nil, errors.New("field SecurityGroupRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules/" + fmt.Sprint(req.SecurityGroupRuleID) + "",
		Headers: http.Header{},
	}

	var resp GetSecurityGroupRuleResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type setSecurityGroupRuleRequest struct {
	Zone scw.Zone `json:"-"`

	SecurityGroupID string `json:"-"`

	SecurityGroupRuleID string `json:"-"`

	ID string `json:"id"`
	// Protocol
	//
	// Default value: TCP
	Protocol SecurityGroupRuleProtocol `json:"protocol"`
	// Direction
	//
	// Default value: inbound
	Direction SecurityGroupRuleDirection `json:"direction"`
	// Action
	//
	// Default value: accept
	Action SecurityGroupRuleAction `json:"action"`

	IPRange string `json:"ip_range"`

	DestPortFrom *uint32 `json:"dest_port_from"`

	DestPortTo *uint32 `json:"dest_port_to"`

	Position uint32 `json:"position"`

	Editable bool `json:"editable"`
}

// setSecurityGroupRule update security group rule
func (s *API) setSecurityGroupRule(req *setSecurityGroupRuleRequest, opts ...scw.RequestOption) (*setSecurityGroupRuleResponse, error) {
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

	if fmt.Sprint(req.SecurityGroupRuleID) == "" {
		return nil, errors.New("field SecurityGroupRuleID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/security_groups/" + fmt.Sprint(req.SecurityGroupID) + "/rules/" + fmt.Sprint(req.SecurityGroupRuleID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setSecurityGroupRuleResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListComputeClustersRequest struct {
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListComputeClusters list compute-clusters
//
// List all compute-clusters
func (s *API) ListComputeClusters(req *ListComputeClustersRequest, opts ...scw.RequestOption) (*ListComputeClustersResponse, error) {
	var err error

	defaultOrganization, exist := s.client.GetDefaultProjectID()
	if (req.Organization == nil || *req.Organization == "") && exist {
		req.Organization = &defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListComputeClustersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateComputeClusterRequest struct {
	Zone scw.Zone `json:"-"`

	Name string `json:"name,omitempty"`

	Organization string `json:"organization,omitempty"`
	// PolicyMode
	//
	// Default value: optional
	PolicyMode ComputeClusterPolicyMode `json:"policy_mode,omitempty"`
	// PolicyType
	//
	// Default value: low_latency
	PolicyType ComputeClusterPolicyType `json:"policy_type,omitempty"`
}

// CreateComputeCluster create compute-cluster
//
// Create a new compute-cluster
func (s *API) CreateComputeCluster(req *CreateComputeClusterRequest, opts ...scw.RequestOption) (*CreateComputeClusterResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateComputeClusterResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetComputeClusterRequest struct {
	Zone scw.Zone `json:"-"`

	ComputeClusterID string `json:"-"`
}

// GetComputeCluster get compute-cluster
//
// Get the given compute-cluster
func (s *API) GetComputeCluster(req *GetComputeClusterRequest, opts ...scw.RequestOption) (*GetComputeClusterResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ComputeClusterID) == "" {
		return nil, errors.New("field ComputeClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters/" + fmt.Sprint(req.ComputeClusterID) + "",
		Headers: http.Header{},
	}

	var resp GetComputeClusterResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type SetComputeClusterRequest struct {
	Zone scw.Zone `json:"-"`

	ComputeClusterID string `json:"-"`

	Name string `json:"name"`

	Organization string `json:"organization"`
	// PolicyMode
	//
	// Default value: optional
	PolicyMode ComputeClusterPolicyMode `json:"policy_mode"`
	// PolicyType
	//
	// Default value: low_latency
	PolicyType ComputeClusterPolicyType `json:"policy_type"`
}

// SetComputeCluster set compute-cluster
//
// Set all parameters of the given compute-cluster
func (s *API) SetComputeCluster(req *SetComputeClusterRequest, opts ...scw.RequestOption) (*SetComputeClusterResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ComputeClusterID) == "" {
		return nil, errors.New("field ComputeClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters/" + fmt.Sprint(req.ComputeClusterID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp SetComputeClusterResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateComputeClusterRequest struct {
	Zone scw.Zone `json:"-"`

	ComputeClusterID string `json:"-"`

	Name *string `json:"name,omitempty"`

	Organization *string `json:"organization,omitempty"`
	// PolicyMode
	//
	// Default value: optional
	PolicyMode ComputeClusterPolicyMode `json:"policy_mode,omitempty"`
	// PolicyType
	//
	// Default value: low_latency
	PolicyType ComputeClusterPolicyType `json:"policy_type,omitempty"`
}

// UpdateComputeCluster update compute-cluster
//
// Update one or more parameter of the given compute-cluster
func (s *API) UpdateComputeCluster(req *UpdateComputeClusterRequest, opts ...scw.RequestOption) (*UpdateComputeClusterResponse, error) {
	var err error

	defaultOrganization, exist := s.client.GetDefaultProjectID()
	if (req.Organization == nil || *req.Organization == "") && exist {
		req.Organization = &defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ComputeClusterID) == "" {
		return nil, errors.New("field ComputeClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters/" + fmt.Sprint(req.ComputeClusterID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateComputeClusterResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteComputeClusterRequest struct {
	Zone scw.Zone `json:"-"`

	ComputeClusterID string `json:"-"`
}

// DeleteComputeCluster delete the given compute-cluster
//
// Delete the given compute-cluster
func (s *API) DeleteComputeCluster(req *DeleteComputeClusterRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ComputeClusterID) == "" {
		return errors.New("field ComputeClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters/" + fmt.Sprint(req.ComputeClusterID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type GetComputeClusterServersRequest struct {
	Zone scw.Zone `json:"-"`

	ComputeClusterID string `json:"-"`
}

// GetComputeClusterServers get compute-cluster servers
//
// Get all servers belonging to the given compute-cluster
func (s *API) GetComputeClusterServers(req *GetComputeClusterServersRequest, opts ...scw.RequestOption) (*GetComputeClusterServersResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ComputeClusterID) == "" {
		return nil, errors.New("field ComputeClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters/" + fmt.Sprint(req.ComputeClusterID) + "/servers",
		Headers: http.Header{},
	}

	var resp GetComputeClusterServersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type SetComputeClusterServersRequest struct {
	Zone scw.Zone `json:"-"`

	ComputeClusterID string `json:"-"`
}

// SetComputeClusterServers set compute-cluster servers
//
// Set all servers belonging to the given compute-cluster
func (s *API) SetComputeClusterServers(req *SetComputeClusterServersRequest, opts ...scw.RequestOption) (*SetComputeClusterServersResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ComputeClusterID) == "" {
		return nil, errors.New("field ComputeClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters/" + fmt.Sprint(req.ComputeClusterID) + "/servers",
		Headers: http.Header{},
	}

	var resp SetComputeClusterServersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateComputeClusterServersRequest struct {
	Zone scw.Zone `json:"-"`

	ComputeClusterID string `json:"-"`
}

// UpdateComputeClusterServers update compute-cluster servers
//
// Update all servers belonging to the given compute-cluster
func (s *API) UpdateComputeClusterServers(req *UpdateComputeClusterServersRequest, opts ...scw.RequestOption) (*UpdateComputeClusterServersResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ComputeClusterID) == "" {
		return nil, errors.New("field ComputeClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters/" + fmt.Sprint(req.ComputeClusterID) + "/servers",
		Headers: http.Header{},
	}

	var resp UpdateComputeClusterServersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteComputeClusterServersRequest struct {
	Zone scw.Zone `json:"-"`

	ComputeClusterID string `json:"-"`
}

// DeleteComputeClusterServers delete compute-cluster servers
//
// Delete all servers from the given compute-cluster
func (s *API) DeleteComputeClusterServers(req *DeleteComputeClusterServersRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ComputeClusterID) == "" {
		return errors.New("field ComputeClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/compute_clusters/" + fmt.Sprint(req.ComputeClusterID) + "/servers",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListIpsRequest struct {
	Zone scw.Zone `json:"-"`

	Organization string `json:"-"`

	Name *string `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListIps list IPs
func (s *API) ListIps(req *ListIpsRequest, opts ...scw.RequestOption) (*ListIpsResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListIpsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListIpsResponse) UnsafeGetTotalCount() int {
	return int(r.TotalCount)
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListIpsResponse) UnsafeAppend(res interface{}) (int, scw.SdkError) {
	results, ok := res.(*ListIpsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Ips = append(r.Ips, results.Ips...)
	r.TotalCount += uint32(len(results.Ips))
	return len(results.Ips), nil
}

type CreateIPRequest struct {
	Zone scw.Zone `json:"-"`

	Organization string `json:"organization,omitempty"`

	Server *string `json:"server,omitempty"`
}

// CreateIP reseve an IP
func (s *API) CreateIP(req *CreateIPRequest, opts ...scw.RequestOption) (*CreateIPResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
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
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp CreateIPResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetIPRequest struct {
	Zone scw.Zone `json:"-"`

	IPID string `json:"-"`
}

// GetIP get IP
//
// Get details of an IP with the given id
func (s *API) GetIP(req *GetIPRequest, opts ...scw.RequestOption) (*GetIPResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	var resp GetIPResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type setIPRequest struct {
	Zone scw.Zone `json:"-"`

	ID string `json:"-"`

	Address net.IP `json:"address"`

	Reverse *string `json:"reverse"`

	Server *ServerSummary `json:"server"`

	Organization string `json:"organization"`
}

func (s *API) setIP(req *setIPRequest, opts ...scw.RequestOption) (*setIPResponse, error) {
	var err error

	if req.Organization == "" {
		defaultOrganization, _ := s.client.GetDefaultProjectID()
		req.Organization = defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.ID) == "" {
		return nil, errors.New("field ID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PUT",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.ID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp setIPResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type updateIPRequest struct {
	Zone scw.Zone `json:"-"`

	IPID string `json:"-"`

	Reverse *NullableStringValue `json:"reverse,omitempty"`

	Server *NullableStringValue `json:"server,omitempty"`
}

// updateIP update IP
func (s *API) updateIP(req *updateIPRequest, opts ...scw.RequestOption) (*UpdateIPResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return nil, errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp UpdateIPResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteIPRequest struct {
	Zone scw.Zone `json:"-"`

	IPID string `json:"-"`
}

// DeleteIP delete IP
//
// Delete the IP with the given id
func (s *API) DeleteIP(req *DeleteIPRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.IPID) == "" {
		return errors.New("field IPID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/ips/" + fmt.Sprint(req.IPID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ListBootscriptsRequest struct {
	Zone scw.Zone `json:"-"`

	Arch *string `json:"-"`

	Title *string `json:"-"`

	Default *bool `json:"-"`

	Public *bool `json:"-"`

	PerPage *int32 `json:"-"`

	Page *int32 `json:"-"`
}

// ListBootscripts list bootscripts
func (s *API) ListBootscripts(req *ListBootscriptsRequest, opts ...scw.RequestOption) (*ListBootscriptsResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	defaultPerPage, exist := s.client.GetDefaultPageSize()
	if (req.PerPage == nil || *req.PerPage == 0) && exist {
		req.PerPage = &defaultPerPage
	}

	query := url.Values{}
	parameter.AddToQuery(query, "arch", req.Arch)
	parameter.AddToQuery(query, "title", req.Title)
	parameter.AddToQuery(query, "default", req.Default)
	parameter.AddToQuery(query, "public", req.Public)
	parameter.AddToQuery(query, "per_page", req.PerPage)
	parameter.AddToQuery(query, "page", req.Page)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/bootscripts",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListBootscriptsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListBootscriptsResponse) UnsafeGetTotalCount() int {
	return int(r.TotalCount)
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListBootscriptsResponse) UnsafeAppend(res interface{}) (int, scw.SdkError) {
	results, ok := res.(*ListBootscriptsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Bootscripts = append(r.Bootscripts, results.Bootscripts...)
	r.TotalCount += uint32(len(results.Bootscripts))
	return len(results.Bootscripts), nil
}

type GetBootscriptRequest struct {
	Zone scw.Zone `json:"-"`

	BootscriptID string `json:"-"`
}

// GetBootscript get bootscripts
//
// Get details of a bootscript with the given id
func (s *API) GetBootscript(req *GetBootscriptRequest, opts ...scw.RequestOption) (*GetBootscriptResponse, error) {
	var err error

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	if fmt.Sprint(req.BootscriptID) == "" {
		return nil, errors.New("field BootscriptID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/bootscripts/" + fmt.Sprint(req.BootscriptID) + "",
		Headers: http.Header{},
	}

	var resp GetBootscriptResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetDashboardRequest struct {
	Zone scw.Zone `json:"-"`

	Organization *string `json:"-"`
}

func (s *API) GetDashboard(req *GetDashboardRequest, opts ...scw.RequestOption) (*GetDashboardResponse, error) {
	var err error

	defaultOrganization, exist := s.client.GetDefaultProjectID()
	if (req.Organization == nil || *req.Organization == "") && exist {
		req.Organization = &defaultOrganization
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization", req.Organization)

	if fmt.Sprint(req.Zone) == "" {
		return nil, errors.New("field Zone cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/instance/v1/zones/" + fmt.Sprint(req.Zone) + "/dashboard",
		Query:   query,
		Headers: http.Header{},
	}

	var resp GetDashboardResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type From interface {
	isFrom()
}

type FromSize struct {
	Value uint64
}

func (FromSize) isFrom() {
}

type FromBaseVolume struct {
	Value string
}

func (FromBaseVolume) isFrom() {
}

type FromBaseSnapshot struct {
	Value string
}

func (FromBaseSnapshot) isFrom() {
}
