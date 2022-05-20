package scaleway

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	// InstanceServerStateStopped transient state of the instance event stop
	InstanceServerStateStopped = "stopped"
	// InstanceServerStateStarted transient state of the instance event start
	InstanceServerStateStarted = "started"
	// InstanceServerStateStandby transient state of the instance event waiting third action or rescue mode
	InstanceServerStateStandby = "standby"

	defaultInstanceServerWaitTimeout        = 10 * time.Minute
	defaultInstancePrivateNICWaitTimeout    = 10 * time.Minute
	defaultInstanceVolumeDeleteTimeout      = 10 * time.Minute
	defaultInstanceSecurityGroupTimeout     = 1 * time.Minute
	defaultInstanceSecurityGroupRuleTimeout = 1 * time.Minute
	defaultInstancePlacementGroupTimeout    = 1 * time.Minute
	defaultInstanceIPTimeout                = 1 * time.Minute
	defaultInstanceRetryInterval            = 5 * time.Second

	defaultInstanceSnapshotWaitTimeout = 1 * time.Hour
)

// instanceAPIWithZone returns a new instance API and the zone for a Create request
func instanceAPIWithZone(d *schema.ResourceData, m interface{}) (*instance.API, scw.Zone, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, err := extractZone(d, meta)
	if err != nil {
		return nil, "", err
	}
	return instanceAPI, zone, nil
}

// instanceAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func instanceAPIWithZoneAndID(m interface{}, zonedID string) (*instance.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(zonedID)
	if err != nil {
		return nil, "", "", err
	}
	return instanceAPI, zone, ID, nil
}

// instanceAPIWithZoneAndNestedID returns an instance API with zone and inner/outer ID extracted from the state
func instanceAPIWithZoneAndNestedID(m interface{}, zonedNestedID string) (*instance.API, scw.Zone, string, string, error) {
	meta := m.(*Meta)
	instanceAPI := instance.NewAPI(meta.scwClient)

	zone, innerID, outerID, err := parseZonedNestedID(zonedNestedID)
	if err != nil {
		return nil, "", "", "", err
	}
	return instanceAPI, zone, innerID, outerID, nil
}

// orderVolumes return an ordered slice based on the volume map key "0", "1", "2",...
func orderVolumes(v map[string]*instance.Volume) []*instance.Volume {
	var indexes []string
	for index := range v {
		indexes = append(indexes, index)
	}
	sort.Strings(indexes)
	var orderedVolumes []*instance.Volume
	for _, index := range indexes {
		orderedVolumes = append(orderedVolumes, v[index])
	}
	return orderedVolumes
}

// sortVolumeServer return an ordered slice based on the volume map key "0", "1", "2",...
func sortVolumeServer(v map[string]*instance.VolumeServer) []*instance.VolumeServer {
	var indexes []string
	for index := range v {
		indexes = append(indexes, index)
	}
	sort.Strings(indexes)
	var sortedVolumes []*instance.VolumeServer
	for _, index := range indexes {
		sortedVolumes = append(sortedVolumes, v[index])
	}
	return sortedVolumes
}

// serverStateFlatten converts the API state to terraform state or return an error.
func serverStateFlatten(fromState instance.ServerState) (string, error) {
	switch fromState {
	case instance.ServerStateStopped:
		return InstanceServerStateStopped, nil
	case instance.ServerStateStoppedInPlace:
		return InstanceServerStateStandby, nil
	case instance.ServerStateRunning:
		return InstanceServerStateStarted, nil
	case instance.ServerStateLocked:
		return "", fmt.Errorf("server is locked, please contact Scaleway support: https://console.scaleway.com/support/tickets")
	}
	return "", fmt.Errorf("server is in an invalid state, someone else might be executing action at the same time")
}

// serverStateExpand converts terraform state to an API state or return an error.
func serverStateExpand(rawState string) (instance.ServerState, error) {
	apiState, exist := map[string]instance.ServerState{
		InstanceServerStateStopped: instance.ServerStateStopped,
		InstanceServerStateStandby: instance.ServerStateStoppedInPlace,
		InstanceServerStateStarted: instance.ServerStateRunning,
	}[rawState]

	if !exist {
		return "", fmt.Errorf("server is in a transient state, someone else might be executing another action at the same time")
	}

	return apiState, nil
}

func reachState(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, serverID string, toState instance.ServerState) error {
	response, err := instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("error getting server %s: %s", serverID, err)
	}
	fromState := response.Server.State

	if response.Server.State == toState {
		return nil
	}

	transitionMap := map[[2]instance.ServerState][]instance.ServerAction{
		{instance.ServerStateStopped, instance.ServerStateRunning}:        {instance.ServerActionPoweron},
		{instance.ServerStateStopped, instance.ServerStateStoppedInPlace}: {instance.ServerActionPoweron, instance.ServerActionStopInPlace},
		{instance.ServerStateRunning, instance.ServerStateStopped}:        {instance.ServerActionPoweroff},
		{instance.ServerStateRunning, instance.ServerStateStoppedInPlace}: {instance.ServerActionStopInPlace},
		{instance.ServerStateStoppedInPlace, instance.ServerStateRunning}: {instance.ServerActionPoweron},
		{instance.ServerStateStoppedInPlace, instance.ServerStateStopped}: {instance.ServerActionPoweron, instance.ServerActionPoweroff},
	}

	actions, exist := transitionMap[[2]instance.ServerState{fromState, toState}]
	if !exist {
		return fmt.Errorf("don't know how to reach state %s from state %s for server %s", toState, fromState, serverID)
	}

	// We need to check that all volumes are ready
	for _, volume := range response.Server.Volumes {
		if volume.State != instance.VolumeServerStateAvailable {
			_, err = instanceAPI.WaitForVolume(&instance.WaitForVolumeRequest{
				Zone:          zone,
				VolumeID:      volume.ID,
				RetryInterval: DefaultWaitRetryInterval,
			})
			if err != nil {
				return fmt.Errorf("volume %s is not ready: %s", volume.ID, err)
			}
		}
	}

	for _, a := range actions {
		err = instanceAPI.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			ServerID:      serverID,
			Action:        a,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(defaultInstanceServerWaitTimeout),
			RetryInterval: DefaultWaitRetryInterval,
		})
		if err != nil {
			return fmt.Errorf("failed to reach state %s from state %s for server %s: %s", toState, fromState, serverID, err)
		}
	}
	return nil
}

// getServerType is a util to get a instance.ServerType by its commercialType
func getServerType(ctx context.Context, apiInstance *instance.API, zone scw.Zone, commercialType string) *instance.ServerType {
	serverType := (*instance.ServerType)(nil)

	serverTypesRes, err := apiInstance.ListServersTypes(&instance.ListServersTypesRequest{
		Zone: zone,
	})
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("cannot get server types: %s", err))
	} else {
		serverType = serverTypesRes.Servers[commercialType]
		if serverType == nil {
			tflog.Warn(ctx, fmt.Sprintf("unrecognized server type: %s", commercialType))
		}
	}

	return serverType
}

// validateLocalVolumeSizes validates the total size of local volumes.
func validateLocalVolumeSizes(volumes map[string]*instance.VolumeServerTemplate, serverType *instance.ServerType, commercialType string) error {
	// Calculate local volume total size.
	var localVolumeTotalSize scw.Size
	for _, volume := range volumes {
		if volume.VolumeType == instance.VolumeVolumeTypeLSSD {
			localVolumeTotalSize += volume.Size
		}
	}

	volumeConstraint := serverType.VolumesConstraint

	// If no root volume provided, count the default root volume size added by the API.
	if rootVolume := volumes["0"]; rootVolume == nil {
		localVolumeTotalSize += volumeConstraint.MinSize
	}

	if localVolumeTotalSize < volumeConstraint.MinSize || localVolumeTotalSize > volumeConstraint.MaxSize {
		min := humanize.Bytes(uint64(volumeConstraint.MinSize))
		if volumeConstraint.MinSize == volumeConstraint.MaxSize {
			return fmt.Errorf("%s total local volume size must be equal to %s", commercialType, min)
		}

		max := humanize.Bytes(uint64(volumeConstraint.MaxSize))
		return fmt.Errorf("%s total local volume size must be between %s and %s", commercialType, min, max)
	}

	return nil
}

// sanitizeVolumeMap removes extra data for API validation.
//
// On the api side, there are two possibles validation schemas for volumes and the validator will be chosen dynamically depending on the passed JSON request
// - With an image (in that case the root volume can be skipped because it is taken from the image)
// - Without an image (in that case, the root volume must be defined)
func sanitizeVolumeMap(serverName string, volumes map[string]*instance.VolumeServerTemplate) map[string]*instance.VolumeServerTemplate {
	m := make(map[string]*instance.VolumeServerTemplate)

	for index, v := range volumes {
		// Remove extra data for API validation.
		switch {
		// If a volume already got an ID it is passed as it to the API without specifying the volume type.
		// TODO: Fix once instance accept volume type in the schema validation
		case v.ID != "":
			v = &instance.VolumeServerTemplate{
				ID:   v.ID,
				Name: v.Name,
			}
		// For the root volume (index 0) if the size is 0, it is considered as a volume created from an image.
		// The size is not passed to the API, so it's computed by the API
		case index == "0" && v.Size == 0:
			v = &instance.VolumeServerTemplate{
				VolumeType: v.VolumeType,
				Boot:       v.Boot,
			}
		// If none of the above conditions are met, the volume is passed as it to the API
		default:
		}
		m[index] = v
	}

	return m
}

func preparePrivateNIC(
	ctx context.Context, data interface{},
	server *instance.Server, vpcAPI *vpc.API,
) ([]*instance.CreatePrivateNICRequest, error) {
	if data == nil {
		return nil, nil
	}

	var res []*instance.CreatePrivateNICRequest

	for _, pn := range data.([]interface{}) {
		r := pn.(map[string]interface{})
		zonedID, pnExist := r["pn_id"]
		privateNetworkID := expandID(zonedID.(string))
		if pnExist {
			currentPN, err := vpcAPI.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
				PrivateNetworkID: expandID(privateNetworkID),
				Zone:             server.Zone,
			}, scw.WithContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("unable to get private network %s: %s", privateNetworkID, err)
			}
			query := &instance.CreatePrivateNICRequest{
				Zone:             currentPN.Zone,
				ServerID:         server.ID,
				PrivateNetworkID: currentPN.ID,
			}
			res = append(res, query)
		}
	}

	return res, nil
}

type privateNICsHandler struct {
	instanceAPI    *instance.API
	serverID       string
	privateNICsMap map[string]*instance.PrivateNIC
	zone           scw.Zone
}

func newPrivateNICHandler(ctx context.Context, api *instance.API, server string, zone scw.Zone) (*privateNICsHandler, error) {
	handler := &privateNICsHandler{
		instanceAPI: api,
		serverID:    server,
		zone:        zone,
	}
	return handler, handler.flatPrivateNICs()
}

func (ph *privateNICsHandler) flatPrivateNICs() error {
	privateNICsMap := make(map[string]*instance.PrivateNIC)
	res, err := ph.instanceAPI.ListPrivateNICs(&instance.ListPrivateNICsRequest{Zone: ph.zone, ServerID: ph.serverID})
	if err != nil {
		return fmt.Errorf("failed to list private NICs: %w", err)
	}
	for _, p := range res.PrivateNics {
		privateNICsMap[p.PrivateNetworkID] = p
	}

	ph.privateNICsMap = privateNICsMap
	return nil
}

func (ph *privateNICsHandler) detach(ctx context.Context, o interface{}, timeout time.Duration) error {
	oPtr := expandStringPtr(o)
	if oPtr != nil && len(*oPtr) > 0 {
		idPN := expandID(*oPtr)
		// check if old private network still exist on instance server
		if p, ok := ph.privateNICsMap[idPN]; ok {
			_, err := waitForPrivateNIC(ctx, ph.instanceAPI, ph.zone, ph.serverID, expandID(p.ID), timeout)
			if err != nil {
				return err
			}
			// detach private NIC
			err = ph.instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
				PrivateNicID: expandID(p.ID),
				Zone:         ph.zone,
				ServerID:     ph.serverID,
			},
				scw.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to detach private NIC: %w", err)
			}
		}
	}

	return nil
}

func (ph *privateNICsHandler) attach(ctx context.Context, n interface{}, timeout time.Duration) error {
	if nPtr := expandStringPtr(n); nPtr != nil {
		// check if new private network was already attached on instance server
		privateNetworkID := expandID(*nPtr)
		if _, ok := ph.privateNICsMap[privateNetworkID]; !ok {
			pn, err := ph.instanceAPI.CreatePrivateNIC(&instance.CreatePrivateNICRequest{
				Zone:             ph.zone,
				ServerID:         ph.serverID,
				PrivateNetworkID: privateNetworkID,
			})
			if err != nil {
				return fmt.Errorf("failed to create private NIC: %w", err)
			}

			_, err = waitForPrivateNIC(ctx, ph.instanceAPI, ph.zone, ph.serverID, pn.PrivateNic.ID, timeout)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ph *privateNICsHandler) set(d *schema.ResourceData) error {
	raw := d.Get("private_network")
	privateNetworks := []map[string]interface{}(nil)
	for index := range raw.([]interface{}) {
		pnKey := fmt.Sprintf("private_network.%d.pn_id", index)
		keyValue := d.Get(pnKey)
		keyRaw, err := ph.get(keyValue.(string))
		if err != nil {
			return err
		}
		privateNetworks = append(privateNetworks, keyRaw.(map[string]interface{}))
	}

	err := d.Set("private_network", privateNetworks)
	if err != nil {
		return fmt.Errorf("failed to set private_network: %w", err)
	}

	return nil
}

func (ph *privateNICsHandler) get(key string) (interface{}, error) {
	locality, id, err := parseLocalizedID(key)
	if err != nil {
		return nil, err
	}
	pn, ok := ph.privateNICsMap[id]
	if !ok {
		return nil, fmt.Errorf("could not find private network ID %s on locality %s", key, locality)
	}
	return map[string]interface{}{
		"pn_id":       key,
		"mac_address": pn.MacAddress,
		"status":      pn.State.String(),
		"zone":        locality,
	}, nil
}

func waitForInstanceSnapshot(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Snapshot, error) {
	retryInterval := defaultInstanceRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	snapshot, err := api.WaitForSnapshot(&instance.WaitForSnapshotRequest{
		SnapshotID:    id,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error waiting for snapshot: %w", err)
	}

	return snapshot, nil
}

func waitForInstanceVolume(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Volume, error) {
	retryInterval := defaultInstanceRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	volume, err := api.WaitForVolume(&instance.WaitForVolumeRequest{
		VolumeID:      id,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error waiting for volume: %w", err)
	}

	return volume, nil
}

func waitForInstanceServer(ctx context.Context, api *instance.API, zone scw.Zone, id string, timeout time.Duration) (*instance.Server, error) {
	retryInterval := defaultInstanceRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	server, err := api.WaitForServer(&instance.WaitForServerRequest{
		Zone:          zone,
		ServerID:      id,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: &retryInterval,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error waiting for server: %w", err)
	}

	return server, nil
}

func waitForPrivateNIC(ctx context.Context, instanceAPI *instance.API, zone scw.Zone, serverID string, privateNICID string, timeout time.Duration) (*instance.PrivateNIC, error) {
	retryInterval := defaultInstanceRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	nic, err := instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
		ServerID:      serverID,
		PrivateNicID:  privateNICID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(timeout),
		RetryInterval: scw.TimeDurationPtr(retryInterval),
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error waiting for private NIC: %w", err)
	}

	return nic, nil
}
