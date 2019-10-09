// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

package k8s

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

// API this API allows you to manage your kapsule clusters.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type ClusterStatus string

const (
	// ClusterStatusUnknown is [insert doc].
	ClusterStatusUnknown = ClusterStatus("unknown")
	// ClusterStatusCreating is [insert doc].
	ClusterStatusCreating = ClusterStatus("creating")
	// ClusterStatusReady is [insert doc].
	ClusterStatusReady = ClusterStatus("ready")
	// ClusterStatusDeleting is [insert doc].
	ClusterStatusDeleting = ClusterStatus("deleting")
	// ClusterStatusDeleted is [insert doc].
	ClusterStatusDeleted = ClusterStatus("deleted")
	// ClusterStatusUpdating is [insert doc].
	ClusterStatusUpdating = ClusterStatus("updating")
	// ClusterStatusWarning is [insert doc].
	ClusterStatusWarning = ClusterStatus("warning")
	// ClusterStatusError is [insert doc].
	ClusterStatusError = ClusterStatus("error")
	// ClusterStatusLocked is [insert doc].
	ClusterStatusLocked = ClusterStatus("locked")
)

func (enum ClusterStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum ClusterStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ClusterStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ClusterStatus(ClusterStatus(tmp).String())
	return nil
}

type ClusterSubStatus string

const (
	// ClusterSubStatusNoDetails is [insert doc].
	ClusterSubStatusNoDetails = ClusterSubStatus("no_details")
	// ClusterSubStatusDeployLoadbalancer is [insert doc].
	ClusterSubStatusDeployLoadbalancer = ClusterSubStatus("deploy_loadbalancer")
	// ClusterSubStatusDeployEtcd is [insert doc].
	ClusterSubStatusDeployEtcd = ClusterSubStatus("deploy_etcd")
	// ClusterSubStatusDeployControlplane is [insert doc].
	ClusterSubStatusDeployControlplane = ClusterSubStatus("deploy_controlplane")
	// ClusterSubStatusDeployNodes is [insert doc].
	ClusterSubStatusDeployNodes = ClusterSubStatus("deploy_nodes")
	// ClusterSubStatusUpdatingEtcd is [insert doc].
	ClusterSubStatusUpdatingEtcd = ClusterSubStatus("updating_etcd")
	// ClusterSubStatusUpdatingControlplane is [insert doc].
	ClusterSubStatusUpdatingControlplane = ClusterSubStatus("updating_controlplane")
)

func (enum ClusterSubStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "no_details"
	}
	return string(enum)
}

func (enum ClusterSubStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ClusterSubStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ClusterSubStatus(ClusterSubStatus(tmp).String())
	return nil
}

type ListClustersRequestOrderBy string

const (
	// ListClustersRequestOrderByCreatedAtAsc is [insert doc].
	ListClustersRequestOrderByCreatedAtAsc = ListClustersRequestOrderBy("created_at_asc")
	// ListClustersRequestOrderByCreatedAtDesc is [insert doc].
	ListClustersRequestOrderByCreatedAtDesc = ListClustersRequestOrderBy("created_at_desc")
	// ListClustersRequestOrderByUpdatedAtAsc is [insert doc].
	ListClustersRequestOrderByUpdatedAtAsc = ListClustersRequestOrderBy("updated_at_asc")
	// ListClustersRequestOrderByUpdatedAtDesc is [insert doc].
	ListClustersRequestOrderByUpdatedAtDesc = ListClustersRequestOrderBy("updated_at_desc")
	// ListClustersRequestOrderByNameAsc is [insert doc].
	ListClustersRequestOrderByNameAsc = ListClustersRequestOrderBy("name_asc")
	// ListClustersRequestOrderByNameDesc is [insert doc].
	ListClustersRequestOrderByNameDesc = ListClustersRequestOrderBy("name_desc")
	// ListClustersRequestOrderByStatusAsc is [insert doc].
	ListClustersRequestOrderByStatusAsc = ListClustersRequestOrderBy("status_asc")
	// ListClustersRequestOrderByStatusDesc is [insert doc].
	ListClustersRequestOrderByStatusDesc = ListClustersRequestOrderBy("status_desc")
	// ListClustersRequestOrderByVersionAsc is [insert doc].
	ListClustersRequestOrderByVersionAsc = ListClustersRequestOrderBy("version_asc")
	// ListClustersRequestOrderByVersionDesc is [insert doc].
	ListClustersRequestOrderByVersionDesc = ListClustersRequestOrderBy("version_desc")
)

func (enum ListClustersRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListClustersRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListClustersRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListClustersRequestOrderBy(ListClustersRequestOrderBy(tmp).String())
	return nil
}

type ListNodesRequestOrderBy string

const (
	// ListNodesRequestOrderByCreatedAtAsc is [insert doc].
	ListNodesRequestOrderByCreatedAtAsc = ListNodesRequestOrderBy("created_at_asc")
	// ListNodesRequestOrderByCreatedAtDesc is [insert doc].
	ListNodesRequestOrderByCreatedAtDesc = ListNodesRequestOrderBy("created_at_desc")
)

func (enum ListNodesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListNodesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListNodesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListNodesRequestOrderBy(ListNodesRequestOrderBy(tmp).String())
	return nil
}

type ListPoolsRequestOrderBy string

const (
	// ListPoolsRequestOrderByCreatedAtAsc is [insert doc].
	ListPoolsRequestOrderByCreatedAtAsc = ListPoolsRequestOrderBy("created_at_asc")
	// ListPoolsRequestOrderByCreatedAtDesc is [insert doc].
	ListPoolsRequestOrderByCreatedAtDesc = ListPoolsRequestOrderBy("created_at_desc")
	// ListPoolsRequestOrderByUpdatedAtAsc is [insert doc].
	ListPoolsRequestOrderByUpdatedAtAsc = ListPoolsRequestOrderBy("updated_at_asc")
	// ListPoolsRequestOrderByUpdatedAtDesc is [insert doc].
	ListPoolsRequestOrderByUpdatedAtDesc = ListPoolsRequestOrderBy("updated_at_desc")
	// ListPoolsRequestOrderByNameAsc is [insert doc].
	ListPoolsRequestOrderByNameAsc = ListPoolsRequestOrderBy("name_asc")
	// ListPoolsRequestOrderByNameDesc is [insert doc].
	ListPoolsRequestOrderByNameDesc = ListPoolsRequestOrderBy("name_desc")
	// ListPoolsRequestOrderByStatusAsc is [insert doc].
	ListPoolsRequestOrderByStatusAsc = ListPoolsRequestOrderBy("status_asc")
	// ListPoolsRequestOrderByStatusDesc is [insert doc].
	ListPoolsRequestOrderByStatusDesc = ListPoolsRequestOrderBy("status_desc")
	// ListPoolsRequestOrderByVersionAsc is [insert doc].
	ListPoolsRequestOrderByVersionAsc = ListPoolsRequestOrderBy("version_asc")
	// ListPoolsRequestOrderByVersionDesc is [insert doc].
	ListPoolsRequestOrderByVersionDesc = ListPoolsRequestOrderBy("version_desc")
)

func (enum ListPoolsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListPoolsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListPoolsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListPoolsRequestOrderBy(ListPoolsRequestOrderBy(tmp).String())
	return nil
}

type NodeStatus string

const (
	// NodeStatusUnknown is [insert doc].
	NodeStatusUnknown = NodeStatus("unknown")
	// NodeStatusCreating is [insert doc].
	NodeStatusCreating = NodeStatus("creating")
	// NodeStatusRebuilding is [insert doc].
	NodeStatusRebuilding = NodeStatus("rebuilding")
	// NodeStatusNotready is [insert doc].
	NodeStatusNotready = NodeStatus("notready")
	// NodeStatusReady is [insert doc].
	NodeStatusReady = NodeStatus("ready")
	// NodeStatusDeleting is [insert doc].
	NodeStatusDeleting = NodeStatus("deleting")
	// NodeStatusDeleted is [insert doc].
	NodeStatusDeleted = NodeStatus("deleted")
	// NodeStatusWarning is [insert doc].
	NodeStatusWarning = NodeStatus("warning")
	// NodeStatusError is [insert doc].
	NodeStatusError = NodeStatus("error")
	// NodeStatusLocked is [insert doc].
	NodeStatusLocked = NodeStatus("locked")
	// NodeStatusRebooting is [insert doc].
	NodeStatusRebooting = NodeStatus("rebooting")
)

func (enum NodeStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum NodeStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *NodeStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = NodeStatus(NodeStatus(tmp).String())
	return nil
}

type PoolStatus string

const (
	// PoolStatusUnknown is [insert doc].
	PoolStatusUnknown = PoolStatus("unknown")
	// PoolStatusCreating is [insert doc].
	PoolStatusCreating = PoolStatus("creating")
	// PoolStatusReady is [insert doc].
	PoolStatusReady = PoolStatus("ready")
	// PoolStatusDeleting is [insert doc].
	PoolStatusDeleting = PoolStatus("deleting")
	// PoolStatusDeleted is [insert doc].
	PoolStatusDeleted = PoolStatus("deleted")
	// PoolStatusUpdating is [insert doc].
	PoolStatusUpdating = PoolStatus("updating")
	// PoolStatusScalling is [insert doc].
	PoolStatusScalling = PoolStatus("scalling")
	// PoolStatusWarning is [insert doc].
	PoolStatusWarning = PoolStatus("warning")
	// PoolStatusError is [insert doc].
	PoolStatusError = PoolStatus("error")
	// PoolStatusLocked is [insert doc].
	PoolStatusLocked = PoolStatus("locked")
)

func (enum PoolStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum PoolStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PoolStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PoolStatus(PoolStatus(tmp).String())
	return nil
}

type Cluster struct {
	// ID display the cluster unique ID
	ID string `json:"id"`
	// Region display the cluster region
	Region scw.Region `json:"region"`
	// Name display the cluster name
	Name string `json:"name"`
	// Description display the cluster description
	Description string `json:"description"`
	// OrganizationID display the cluster organization
	OrganizationID string `json:"organization_id"`
	// Tags display the cluster associated tags
	Tags []string `json:"tags"`
	// Status
	//
	// Default value: unknown
	Status ClusterStatus `json:"status"`
	// SubStatus
	//
	// Default value: no_details
	SubStatus ClusterSubStatus `json:"sub_status"`
	// Version display the cluster version
	Version string `json:"version"`
	// Cni display the cni model
	Cni string `json:"cni"`
	// ClusterURL display the cluster URL
	ClusterURL string `json:"cluster_url"`
	// DNSWildcard display the dns wildcard associated with the cluster
	DNSWildcard string `json:"dns_wildcard"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`

	CurrentCoreCount uint32 `json:"current_core_count"`

	CurrentNodeCount uint32 `json:"current_node_count"`

	CurrentMemCount uint64 `json:"current_mem_count"`

	AutoscalerConfig *ClusterAutoscalerConfig `json:"autoscaler_config"`
}

type ClusterAutoscalerConfig struct {
	ScaleDownDisabled bool `json:"scale_down_disabled"`

	ScaleDownDelayAfterAdd string `json:"scale_down_delay_after_add"`

	Estimator string `json:"estimator"`

	Expander string `json:"expander"`

	IgnoreDaemonsetsUtilization bool `json:"ignore_daemonsets_utilization"`

	BalanceSimilarNodeGroups bool `json:"balance_similar_node_groups"`

	ExpendablePodsPriorityCutoff int32 `json:"expendable_pods_priority_cutoff"`
}

type CreateClusterRequestAutoscalerConfig struct {
	ScaleDownDisabled *bool `json:"scale_down_disabled"`

	ScaleDownDelayAfterAdd *string `json:"scale_down_delay_after_add"`

	Estimator *string `json:"estimator"`

	Expander *string `json:"expander"`

	IgnoreDaemonsetsUtilization *bool `json:"ignore_daemonsets_utilization"`

	BalanceSimilarNodeGroups *bool `json:"balance_similar_node_groups"`

	ExpendablePodsPriorityCutoff *int32 `json:"expendable_pods_priority_cutoff"`
}

type CreateClusterRequestDefaultPoolConfig struct {
	NodeType string `json:"node_type"`

	PlacementGroupID *string `json:"placement_group_id"`

	Autoscaling bool `json:"autoscaling"`

	Size uint32 `json:"size"`

	MinSize *uint32 `json:"min_size"`

	MaxSize *uint32 `json:"max_size"`

	ContainerRuntime *string `json:"container_runtime"`

	Autohealing bool `json:"autohealing"`
}

type ListClusterAvailableVersionsResponse struct {
	Versions []*Version `json:"versions"`
}

type ListClustersResponse struct {
	TotalCount uint32 `json:"total_count"`

	Clusters []*Cluster `json:"clusters"`
}

type ListNodesResponse struct {
	TotalCount uint32 `json:"total_count"`

	Nodes []*Node `json:"nodes"`
}

type ListPoolsResponse struct {
	TotalCount uint32 `json:"total_count"`

	Pools []*Pool `json:"pools"`
}

type ListVersionsResponse struct {
	Versions []*Version `json:"versions"`
}

type Node struct {
	// ID display node unique ID
	ID string `json:"id"`
	// PoolID display pool unique ID
	PoolID string `json:"pool_id"`
	// ClusterID display cluster unique ID
	ClusterID string `json:"cluster_id"`

	Region scw.Region `json:"region"`
	// Name display node name
	Name string `json:"name"`
	// PublicIPV4 display the servers public IPv4 address
	PublicIPV4 *string `json:"public_ip_v4"`
	// PublicIPV6 display the servers public IPv6 address
	PublicIPV6 *string `json:"public_ip_v6"`
	// NpdStatus display kubernetes node conditions
	NpdStatus map[string]string `json:"npd_status"`
	// Status
	//
	// Default value: unknown
	Status NodeStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

type Pool struct {
	// ID display pool unique ID
	ID string `json:"id"`
	// ClusterID display cluster unique ID
	ClusterID string `json:"cluster_id"`
	// Region display the cluster region
	Region scw.Region `json:"region"`
	// Name display pool name
	Name string `json:"name"`
	// Status
	//
	// Default value: unknown
	Status PoolStatus `json:"status"`
	// Version display pool version
	Version string `json:"version"`
	// NodeType display nodes commercial type (e.g. GP1-M)
	NodeType string `json:"node_type"`
	// Autoscaling enable or disable autoscaling
	Autoscaling bool `json:"autoscaling"`
	// Autohealing enable or disable autohealing
	Autohealing bool `json:"autohealing"`
	// Size target number of nodes
	Size uint32 `json:"size"`
	// MinSize display lower limit for this pool
	MinSize uint32 `json:"min_size"`
	// MaxSize display upper limit for this pool
	MaxSize uint32 `json:"max_size"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`

	CurrentCoreCount uint32 `json:"current_core_count"`

	CurrentNodeCount uint32 `json:"current_node_count"`

	CurrentMemCount uint64 `json:"current_mem_count"`

	ContainerRuntime string `json:"container_runtime"`
}

type ResetClusterAdminTokenResponse struct {
}

type UpdateClusterRequestAutoscalerConfig struct {
	ScaleDownDisabled *bool `json:"scale_down_disabled"`

	ScaleDownDelayAfterAdd *string `json:"scale_down_delay_after_add"`

	Estimator *string `json:"estimator"`

	Expander *string `json:"expander"`

	IgnoreDaemonsetsUtilization *bool `json:"ignore_daemonsets_utilization"`

	BalanceSimilarNodeGroups *bool `json:"balance_similar_node_groups"`

	ExpendablePodsPriorityCutoff *int32 `json:"expendable_pods_priority_cutoff"`
}

type Version struct {
	Name string `json:"name"`

	Description string `json:"description"`

	Label string `json:"label"`

	Cni []string `json:"cni"`

	Ingress []string `json:"ingress"`

	Monitoring bool `json:"monitoring"`

	Region scw.Region `json:"region"`
}

// Service API

type ListClustersRequest struct {
	Region scw.Region `json:"-"`
	// OrderBy you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListClustersRequestOrderBy `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize set the maximum list size
	PageSize *int32 `json:"-"`
	// Name filter clusters per name
	Name *string `json:"-"`
	// OrganizationID filter cluster by organization
	OrganizationID *string `json:"-"`
	// Status filter cluster by status
	//
	// Default value: unknown
	Status ClusterStatus `json:"-"`
}

// ListClusters list all your clusters
func (s *API) ListClusters(req *ListClustersRequest, opts ...scw.RequestOption) (*ListClustersResponse, error) {
	var err error

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if (req.OrganizationID == nil || *req.OrganizationID == "") && exist {
		req.OrganizationID = &defaultOrganizationID
	}

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
	parameter.AddToQuery(query, "status", req.Status)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListClustersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListClustersResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListClustersResponse) UnsafeAppend(res interface{}) (uint32, scw.SdkError) {
	results, ok := res.(*ListClustersResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Clusters = append(r.Clusters, results.Clusters...)
	r.TotalCount += uint32(len(results.Clusters))
	return uint32(len(results.Clusters)), nil
}

type CreateClusterRequest struct {
	Region scw.Region `json:"-"`
	// OrganizationID organization owning the resource
	OrganizationID string `json:"organization_id"`
	// Name cluster name
	Name string `json:"name"`
	// Description description
	Description string `json:"description"`
	// Tags list of keyword
	Tags []string `json:"tags"`
	// Version set the cluster version (you can get available versions by calling ListVersions)
	Version string `json:"version"`
	// Cni set the Container Network Interface
	Cni string `json:"cni"`
	// EnableDashboard enable or disable Kubernetes dashboard preinstallation
	EnableDashboard bool `json:"enable_dashboard"`
	// Ingress preinstall an ingress controller into your cluster
	Ingress string `json:"ingress"`

	DefaultPoolConfig *CreateClusterRequestDefaultPoolConfig `json:"default_pool_config"`

	AutoscalerConfig *CreateClusterRequestAutoscalerConfig `json:"autoscaler_config"`
}

// CreateCluster create a new cluster
//
// Create a new kubernetes cluster.
func (s *API) CreateCluster(req *CreateClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
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
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetClusterRequest struct {
	Region scw.Region `json:"-"`

	ClusterID string `json:"-"`
}

// GetCluster get cluster details
//
// Get the cluster details associated with the given id.
func (s *API) GetCluster(req *GetClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "",
		Headers: http.Header{},
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateClusterRequest struct {
	Region scw.Region `json:"-"`
	// ClusterID cluster ID
	ClusterID string `json:"-"`
	// Description description
	Description *string `json:"description"`
	// Tags list of keyword
	Tags *[]string `json:"tags"`

	AutoscalerConfig *UpdateClusterRequestAutoscalerConfig `json:"autoscaler_config"`
}

// UpdateCluster update an existing cluster
//
// Update the cluster associated with the given id.
func (s *API) UpdateCluster(req *UpdateClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteClusterRequest struct {
	Region scw.Region `json:"-"`

	ClusterID string `json:"-"`
}

// DeleteCluster delete an existing cluster
//
// Delete the cluster associated with the given id.
func (s *API) DeleteCluster(req *DeleteClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "",
		Headers: http.Header{},
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpgradeClusterRequest struct {
	Region scw.Region `json:"-"`

	ClusterID string `json:"-"`

	Version string `json:"version"`

	UpgradePools bool `json:"upgrade_pools"`
}

// UpgradeCluster upgrade an existing cluster
//
// Upgrade the cluster associated with the given id.
func (s *API) UpgradeCluster(req *UpgradeClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/upgrade",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListClusterAvailableVersionsRequest struct {
	Region scw.Region `json:"-"`

	ClusterID string `json:"-"`
}

func (s *API) ListClusterAvailableVersions(req *ListClusterAvailableVersionsRequest, opts ...scw.RequestOption) (*ListClusterAvailableVersionsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/available-versions",
		Headers: http.Header{},
	}

	var resp ListClusterAvailableVersionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type getClusterKubeConfigRequest struct {
	Region scw.Region `json:"-"`

	ClusterID string `json:"-"`
}

// getClusterKubeConfig download kubeconfig
func (s *API) getClusterKubeConfig(req *getClusterKubeConfigRequest, opts ...scw.RequestOption) (*scw.File, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/kubeconfig",
		Headers: http.Header{},
	}

	var resp scw.File

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ResetClusterAdminTokenRequest struct {
	Region scw.Region `json:"-"`

	ClusterID string `json:"-"`
}

// ResetClusterAdminToken revoke and renew your admin token
//
// Revoke and renew your cluster admin token, you will have to download kubeconfig again.
func (s *API) ResetClusterAdminToken(req *ResetClusterAdminTokenRequest, opts ...scw.RequestOption) (*ResetClusterAdminTokenResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/reset-admin-token",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ResetClusterAdminTokenResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListPoolsRequest struct {
	Region scw.Region `json:"-"`
	// ClusterID display the cluster unique ID
	ClusterID string `json:"-"`
	// OrderBy you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListPoolsRequestOrderBy `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize set the maximum list size
	PageSize *int32 `json:"-"`
	// Name filter pools per name
	Name *string `json:"-"`
}

// ListPools list all your cluster pools
func (s *API) ListPools(req *ListPoolsRequest, opts ...scw.RequestOption) (*ListPoolsResponse, error) {
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

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/pools",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListPoolsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPoolsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPoolsResponse) UnsafeAppend(res interface{}) (uint32, scw.SdkError) {
	results, ok := res.(*ListPoolsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Pools = append(r.Pools, results.Pools...)
	r.TotalCount += uint32(len(results.Pools))
	return uint32(len(results.Pools)), nil
}

type CreatePoolRequest struct {
	Region scw.Region `json:"-"`

	ClusterID string `json:"-"`

	Name string `json:"name"`

	NodeType string `json:"node_type"`

	PlacementGroupID *string `json:"placement_group_id"`

	Autoscaling bool `json:"autoscaling"`

	Size uint32 `json:"size"`

	MinSize *uint32 `json:"min_size"`

	MaxSize *uint32 `json:"max_size"`

	ContainerRuntime *string `json:"container_runtime"`

	Autohealing bool `json:"autohealing"`
}

// CreatePool create a new pool
//
// Create a new pool in your cluster.
func (s *API) CreatePool(req *CreatePoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/pools",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetPoolRequest struct {
	Region scw.Region `json:"-"`

	PoolID string `json:"-"`
}

// GetPool get pool details
//
// Get the pool details associated with the given id.
func (s *API) GetPool(req *GetPoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "",
		Headers: http.Header{},
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpgradePoolRequest struct {
	Region scw.Region `json:"-"`

	PoolID string `json:"-"`

	Version string `json:"version"`
}

// UpgradePool upgrade an existing cluster pool
//
// Upgrade the pool associated with the given id.
func (s *API) UpgradePool(req *UpgradePoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "/upgrade",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdatePoolRequest struct {
	Region scw.Region `json:"-"`

	PoolID string `json:"-"`

	Autoscaling *bool `json:"autoscaling"`

	Size *uint32 `json:"size"`

	MinSize *uint32 `json:"min_size"`

	MaxSize *uint32 `json:"max_size"`

	Autohealing *bool `json:"autohealing"`
}

// UpdatePool update an existing cluster pool
//
// Update the pool associated with the given id (nodes will be replaced one by one, quotas must be set to allow user to have -at least- one more node than the size of its current pool).
func (s *API) UpdatePool(req *UpdatePoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeletePoolRequest struct {
	Region scw.Region `json:"-"`

	PoolID string `json:"-"`
}

// DeletePool delete an existing cluster pool
//
// Delete the pool associated with the given id.
func (s *API) DeletePool(req *DeletePoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "",
		Headers: http.Header{},
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListNodesRequest struct {
	Region scw.Region `json:"-"`
	// ClusterID cluster unique ID
	ClusterID string `json:"-"`
	// PoolID filter nodes by pool id
	PoolID *string `json:"-"`
	// OrderBy you can order the response by created_at asc/desc or name asc/desc
	//
	// Default value: created_at_asc
	OrderBy ListNodesRequestOrderBy `json:"-"`
	// Page page number
	Page *int32 `json:"-"`
	// PageSize set the maximum list size
	PageSize *int32 `json:"-"`
	// Name filter nodes by name
	Name *string `json:"-"`
	// Status filter nodes by status
	//
	// Default value: unknown
	Status NodeStatus `json:"-"`
}

// ListNodes list all your cluster nodes
func (s *API) ListNodes(req *ListNodesRequest, opts ...scw.RequestOption) (*ListNodesResponse, error) {
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
	parameter.AddToQuery(query, "pool_id", req.PoolID)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "status", req.Status)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/nodes",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListNodesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListNodesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListNodesResponse) UnsafeAppend(res interface{}) (uint32, scw.SdkError) {
	results, ok := res.(*ListNodesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Nodes = append(r.Nodes, results.Nodes...)
	r.TotalCount += uint32(len(results.Nodes))
	return uint32(len(results.Nodes)), nil
}

type GetNodeRequest struct {
	Region scw.Region `json:"-"`

	NodeID string `json:"-"`
}

// GetNode get node details
//
// Get the node associated with the given id.
func (s *API) GetNode(req *GetNodeRequest, opts ...scw.RequestOption) (*Node, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NodeID) == "" {
		return nil, errors.New("field NodeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/nodes/" + fmt.Sprint(req.NodeID) + "",
		Headers: http.Header{},
	}

	var resp Node

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ReplaceNodeRequest struct {
	Region scw.Region `json:"-"`

	NodeID string `json:"-"`
}

// ReplaceNode replace a node by another
//
// Replace a node by another (first the node is deleted, then a new one is created).
func (s *API) ReplaceNode(req *ReplaceNodeRequest, opts ...scw.RequestOption) (*Node, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NodeID) == "" {
		return nil, errors.New("field NodeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/nodes/" + fmt.Sprint(req.NodeID) + "/replace",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Node

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RebootNodeRequest struct {
	Region scw.Region `json:"-"`

	NodeID string `json:"-"`
}

// RebootNode reboot node
//
// Reboot node.
func (s *API) RebootNode(req *RebootNodeRequest, opts ...scw.RequestOption) (*Node, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NodeID) == "" {
		return nil, errors.New("field NodeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/nodes/" + fmt.Sprint(req.NodeID) + "/reboot",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Node

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListVersionsRequest struct {
	Region scw.Region `json:"-"`
}

// ListVersions list available versions
func (s *API) ListVersions(req *ListVersionsRequest, opts ...scw.RequestOption) (*ListVersionsResponse, error) {
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
		Path:    "/k8s/v1beta3/regions/" + fmt.Sprint(req.Region) + "/versions",
		Headers: http.Header{},
	}

	var resp ListVersionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
