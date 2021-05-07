package scaleway

import (
	rdb "github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type RdbAPIInterface interface {
	GetServiceInfo(req *rdb.GetServiceInfoRequest, opts ...scw.RequestOption) (*scw.ServiceInfo, error)
	ListDatabaseEngines(req *rdb.ListDatabaseEnginesRequest, opts ...scw.RequestOption) (*rdb.ListDatabaseEnginesResponse, error)
	ListNodeTypes(req *rdb.ListNodeTypesRequest, opts ...scw.RequestOption) (*rdb.ListNodeTypesResponse, error)
	ListDatabaseBackups(req *rdb.ListDatabaseBackupsRequest, opts ...scw.RequestOption) (*rdb.ListDatabaseBackupsResponse, error)
	CreateDatabaseBackup(req *rdb.CreateDatabaseBackupRequest, opts ...scw.RequestOption) (*rdb.DatabaseBackup, error)
	GetDatabaseBackup(req *rdb.GetDatabaseBackupRequest, opts ...scw.RequestOption) (*rdb.DatabaseBackup, error)
	UpdateDatabaseBackup(req *rdb.UpdateDatabaseBackupRequest, opts ...scw.RequestOption) (*rdb.DatabaseBackup, error)
	DeleteDatabaseBackup(req *rdb.DeleteDatabaseBackupRequest, opts ...scw.RequestOption) (*rdb.DatabaseBackup, error)
	RestoreDatabaseBackup(req *rdb.RestoreDatabaseBackupRequest, opts ...scw.RequestOption) (*rdb.DatabaseBackup, error)
	ExportDatabaseBackup(req *rdb.ExportDatabaseBackupRequest, opts ...scw.RequestOption) (*rdb.DatabaseBackup, error)
	UpgradeInstance(req *rdb.UpgradeInstanceRequest, opts ...scw.RequestOption) (*rdb.Instance, error)
	ListInstances(req *rdb.ListInstancesRequest, opts ...scw.RequestOption) (*rdb.ListInstancesResponse, error)
	GetInstance(req *rdb.GetInstanceRequest, opts ...scw.RequestOption) (*rdb.Instance, error)
	CreateInstance(req *rdb.CreateInstanceRequest, opts ...scw.RequestOption) (*rdb.Instance, error)
	UpdateInstance(req *rdb.UpdateInstanceRequest, opts ...scw.RequestOption) (*rdb.Instance, error)
	DeleteInstance(req *rdb.DeleteInstanceRequest, opts ...scw.RequestOption) (*rdb.Instance, error)
	CloneInstance(req *rdb.CloneInstanceRequest, opts ...scw.RequestOption) (*rdb.Instance, error)
	GetInstanceCertificate(req *rdb.GetInstanceCertificateRequest, opts ...scw.RequestOption) (*scw.File, error)
	RenewInstanceCertificate(req *rdb.RenewInstanceCertificateRequest, opts ...scw.RequestOption) error
	GetInstanceMetrics(req *rdb.GetInstanceMetricsRequest, opts ...scw.RequestOption) (*rdb.InstanceMetrics, error)
	PrepareInstanceLogs(req *rdb.PrepareInstanceLogsRequest, opts ...scw.RequestOption) (*rdb.PrepareInstanceLogsResponse, error)
	ListInstanceLogs(req *rdb.ListInstanceLogsRequest, opts ...scw.RequestOption) (*rdb.ListInstanceLogsResponse, error)
	GetInstanceLog(req *rdb.GetInstanceLogRequest, opts ...scw.RequestOption) (*rdb.InstanceLog, error)
	AddInstanceSettings(req *rdb.AddInstanceSettingsRequest, opts ...scw.RequestOption) (*rdb.AddInstanceSettingsResponse, error)
	DeleteInstanceSettings(req *rdb.DeleteInstanceSettingsRequest, opts ...scw.RequestOption) (*rdb.DeleteInstanceSettingsResponse, error)
	SetInstanceSettings(req *rdb.SetInstanceSettingsRequest, opts ...scw.RequestOption) (*rdb.SetInstanceSettingsResponse, error)
	ListInstanceACLRules(req *rdb.ListInstanceACLRulesRequest, opts ...scw.RequestOption) (*rdb.ListInstanceACLRulesResponse, error)
	AddInstanceACLRules(req *rdb.AddInstanceACLRulesRequest, opts ...scw.RequestOption) (*rdb.AddInstanceACLRulesResponse, error)
	SetInstanceACLRules(req *rdb.SetInstanceACLRulesRequest, opts ...scw.RequestOption) (*rdb.SetInstanceACLRulesResponse, error)
	DeleteInstanceACLRules(req *rdb.DeleteInstanceACLRulesRequest, opts ...scw.RequestOption) (*rdb.DeleteInstanceACLRulesResponse, error)
	ListUsers(req *rdb.ListUsersRequest, opts ...scw.RequestOption) (*rdb.ListUsersResponse, error)
	CreateUser(req *rdb.CreateUserRequest, opts ...scw.RequestOption) (*rdb.User, error)
	UpdateUser(req *rdb.UpdateUserRequest, opts ...scw.RequestOption) (*rdb.User, error)
	DeleteUser(req *rdb.DeleteUserRequest, opts ...scw.RequestOption) error
	ListDatabases(req *rdb.ListDatabasesRequest, opts ...scw.RequestOption) (*rdb.ListDatabasesResponse, error)
	CreateDatabase(req *rdb.CreateDatabaseRequest, opts ...scw.RequestOption) (*rdb.Database, error)
	DeleteDatabase(req *rdb.DeleteDatabaseRequest, opts ...scw.RequestOption) error
	ListPrivileges(req *rdb.ListPrivilegesRequest, opts ...scw.RequestOption) (*rdb.ListPrivilegesResponse, error)
	SetPrivilege(req *rdb.SetPrivilegeRequest, opts ...scw.RequestOption) (*rdb.Privilege, error)
	ListSnapshots(req *rdb.ListSnapshotsRequest, opts ...scw.RequestOption) (*rdb.ListSnapshotsResponse, error)
	GetSnapshot(req *rdb.GetSnapshotRequest, opts ...scw.RequestOption) (*rdb.Snapshot, error)
	CreateSnapshot(req *rdb.CreateSnapshotRequest, opts ...scw.RequestOption) (*rdb.Snapshot, error)
	UpdateSnapshot(req *rdb.UpdateSnapshotRequest, opts ...scw.RequestOption) (*rdb.Snapshot, error)
	DeleteSnapshot(req *rdb.DeleteSnapshotRequest, opts ...scw.RequestOption) (*rdb.Snapshot, error)
	CreateInstanceFromSnapshot(req *rdb.CreateInstanceFromSnapshotRequest, opts ...scw.RequestOption) (*rdb.Instance, error)
	WaitForInstance(req *rdb.WaitForInstanceRequest, opts ...scw.RequestOption) (*rdb.Instance, error)
}
