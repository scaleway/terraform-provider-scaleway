package mongodb

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mongodb "github.com/scaleway/scaleway-sdk-go/api/mongodb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	defaultMongodbInstanceTimeout           = 30 * time.Minute
	defaultMongodbSnapshotTimeout           = 30 * time.Minute
	defaultWaitMongodbInstanceRetryInterval = 10 * time.Second
)

const (
	defaultVolumeSize = 5
)

func newAPI(m any) *mongodb.API {
	return mongodb.NewAPI(meta.ExtractScwClient(m))
}

func newAPIWithRegion(d *schema.ResourceData, m any) (*mongodb.API, scw.Region, error) {
	region, err := meta.ExtractRegion(d, m)
	if err != nil {
		return nil, "", err
	}

	return newAPI(m), region, nil
}

// NewAPIWithRegionAndID returns a mongoDB API with region and ID extracted from the state
func NewAPIWithRegionAndID(m any, id string) (*mongodb.API, scw.Region, string, error) {
	region, ID, err := regional.ParseID(id)
	if err != nil {
		return nil, "", "", err
	}

	return newAPI(m), region, ID, nil
}

func waitForInstance(ctx context.Context, api *mongodb.API, region scw.Region, id string, timeout time.Duration) (*mongodb.Instance, error) {
	retryInterval := defaultWaitMongodbInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForInstance(&mongodb.WaitForInstanceRequest{
		Timeout:       scw.TimeDurationPtr(timeout),
		InstanceID:    id,
		Region:        region,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

func waitForSnapshot(ctx context.Context, api *mongodb.API, region scw.Region, instanceID string, snapshotID string, timeout time.Duration) (*mongodb.Snapshot, error) {
	retryInterval := defaultWaitMongodbInstanceRetryInterval
	if transport.DefaultWaitRetryInterval != nil {
		retryInterval = *transport.DefaultWaitRetryInterval
	}

	return api.WaitForSnapshot(&mongodb.WaitForSnapshotRequest{
		Timeout:       scw.TimeDurationPtr(timeout),
		InstanceID:    instanceID,
		SnapshotID:    snapshotID,
		Region:        region,
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
}

// expandUserRoles converts Terraform roles to SDK UserRole slice
func expandUserRoles(rolesSet *schema.Set) []*mongodb.UserRole {
	if rolesSet == nil || rolesSet.Len() == 0 {
		return nil
	}

	roles := make([]*mongodb.UserRole, 0, rolesSet.Len())

	for _, roleInterface := range rolesSet.List() {
		roleMap := roleInterface.(map[string]any)

		userRole := &mongodb.UserRole{
			Role: mongodb.UserRoleRole(roleMap["role"].(string)),
		}

		if dbName, ok := roleMap["database_name"]; ok && dbName.(string) != "" {
			userRole.DatabaseName = types.ExpandStringPtr(dbName)
		}

		if anyDB, ok := roleMap["any_database"]; ok && anyDB.(bool) {
			userRole.AnyDatabase = scw.BoolPtr(true)
		}

		roles = append(roles, userRole)
	}

	return roles
}

// flattenUserRoles converts SDK UserRole slice to Terraform roles
func flattenUserRoles(roles []*mongodb.UserRole) []any {
	if len(roles) == 0 {
		return nil
	}

	result := make([]any, 0, len(roles))

	for _, role := range roles {
		roleMap := map[string]any{
			"role": string(role.Role),
		}

		if role.DatabaseName != nil {
			roleMap["database_name"] = *role.DatabaseName
		}

		if role.AnyDatabase != nil && *role.AnyDatabase {
			roleMap["any_database"] = true
		}

		result = append(result, roleMap)
	}

	return result
}
