package instance

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	// InstanceServerStateStopped transient state of the instance event stop
	InstanceServerStateStopped = "stopped"
	// InstanceServerStateStarted transient state of the instance event start
	InstanceServerStateStarted = "started"
	// InstanceServerStateStandby transient state of the instance event waiting third action or rescue mode
	InstanceServerStateStandby = "standby"

	DefaultInstanceServerWaitTimeout        = 10 * time.Minute
	defaultInstancePrivateNICWaitTimeout    = 10 * time.Minute
	defaultInstanceVolumeDeleteTimeout      = 10 * time.Minute
	defaultInstanceSecurityGroupTimeout     = 1 * time.Minute
	defaultInstanceSecurityGroupRuleTimeout = 1 * time.Minute
	defaultInstancePlacementGroupTimeout    = 1 * time.Minute
	defaultInstanceIPTimeout                = 1 * time.Minute
	defaultInstanceIPReverseDNSTimeout      = 10 * time.Minute
	defaultInstanceRetryInterval            = 5 * time.Second

	defaultInstanceSnapshotWaitTimeout = 1 * time.Hour

	defaultInstanceImageTimeout = 1 * time.Hour
)

// newAPIWithZone returns a new instance API and the zone for a Create request
func newAPIWithZone(d *schema.ResourceData, m interface{}) (*instance.API, scw.Zone, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}
	return instanceAPI, zone, nil
}

// NewAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m interface{}, zonedID string) (*instance.API, scw.Zone, string, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonedID)
	if err != nil {
		return nil, "", "", err
	}
	return instanceAPI, zone, ID, nil
}

// NewAPIWithZoneAndNestedID returns an instance API with zone and inner/outer ID extracted from the state
func NewAPIWithZoneAndNestedID(m interface{}, zonedNestedID string) (*instance.API, scw.Zone, string, string, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))

	zone, innerID, outerID, err := zonal.ParseNestedID(zonedNestedID)
	if err != nil {
		return nil, "", "", "", err
	}
	return instanceAPI, zone, innerID, outerID, nil
}

// orderVolumes return an ordered slice based on the volume map key "0", "1", "2",...
func orderVolumes(v map[string]*instance.Volume) []*instance.Volume {
	indexes := make([]string, 0, len(v))
	for index := range v {
		indexes = append(indexes, index)
	}
	sort.Strings(indexes)

	orderedVolumes := make([]*instance.Volume, 0, len(indexes))
	for _, index := range indexes {
		orderedVolumes = append(orderedVolumes, v[index])
	}
	return orderedVolumes
}

// sortVolumeServer return an ordered slice based on the volume map key "0", "1", "2",...
func sortVolumeServer(v map[string]*instance.VolumeServer) []*instance.VolumeServer {
	indexes := make([]string, 0, len(v))
	for index := range v {
		indexes = append(indexes, index)
	}
	sort.Strings(indexes)

	sortedVolumes := make([]*instance.VolumeServer, 0, len(indexes))
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
		return "", errors.New("server is locked, please contact Scaleway support: https://console.scaleway.com/support/tickets")
	}
	return "", errors.New("server is in an invalid state, someone else might be executing action at the same time")
}

// serverStateExpand converts terraform state to an API state or return an error.
func serverStateExpand(rawState string) (instance.ServerState, error) {
	apiState, exist := map[string]instance.ServerState{
		InstanceServerStateStopped: instance.ServerStateStopped,
		InstanceServerStateStandby: instance.ServerStateStoppedInPlace,
		InstanceServerStateStarted: instance.ServerStateRunning,
	}[rawState]

	if !exist {
		return "", errors.New("server is in a transient state, someone else might be executing another action at the same time")
	}

	return apiState, nil
}

func reachState(ctx context.Context, api *BlockAndInstanceAPI, zone scw.Zone, serverID string, toState instance.ServerState) error {
	response, err := api.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
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
		if volume.VolumeType == block.BlockVolumeType {
			_, err := api.blockAPI.WaitForVolumeAndReferences(&blockSDK.WaitForVolumeAndReferencesRequest{
				VolumeID:      volume.ID,
				Zone:          zone,
				RetryInterval: transport.DefaultWaitRetryInterval,
			})
			if err != nil {
				return err
			}
		} else if volume.State != instance.VolumeServerStateAvailable {
			_, err = api.WaitForVolume(&instance.WaitForVolumeRequest{
				Zone:          zone,
				VolumeID:      volume.ID,
				RetryInterval: transport.DefaultWaitRetryInterval,
			})
			if err != nil {
				return err
			}
		}
	}

	for _, a := range actions {
		err = api.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
			ServerID:      serverID,
			Action:        a,
			Zone:          zone,
			Timeout:       scw.TimeDurationPtr(DefaultInstanceServerWaitTimeout),
			RetryInterval: transport.DefaultWaitRetryInterval,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// getServerType is a util to get a instance.ServerType by its commercialType
func getServerType(ctx context.Context, apiInstance *instance.API, zone scw.Zone, commercialType string) *instance.ServerType {
	serverType, err := apiInstance.GetServerType(&instance.GetServerTypeRequest{
		Zone: zone,
		Name: commercialType,
	})
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("cannot get server types: %s", err))
	} else {
		if serverType == nil {
			tflog.Warn(ctx, "unrecognized server type: "+commercialType)
		}
		return serverType
	}

	return nil
}

// validateLocalVolumeSizes validates the total size of local volumes.
func validateLocalVolumeSizes(volumes map[string]*instance.VolumeServerTemplate, serverType *instance.ServerType, commercialType string) error {
	// Calculate local volume total size.
	var localVolumeTotalSize scw.Size
	for _, volume := range volumes {
		if volume.VolumeType == instance.VolumeVolumeTypeLSSD && volume.Size != nil {
			localVolumeTotalSize += *volume.Size
		}
	}

	volumeConstraint := serverType.VolumesConstraint

	// If no root volume provided, count the default root volume size added by the API.
	if rootVolume := volumes["0"]; rootVolume == nil {
		localVolumeTotalSize += volumeConstraint.MinSize
	}

	if localVolumeTotalSize < volumeConstraint.MinSize || localVolumeTotalSize > volumeConstraint.MaxSize {
		minSize := humanize.Bytes(uint64(volumeConstraint.MinSize))
		if volumeConstraint.MinSize == volumeConstraint.MaxSize {
			return fmt.Errorf("%s total local volume size must be equal to %s", commercialType, minSize)
		}

		maxSize := humanize.Bytes(uint64(volumeConstraint.MaxSize))
		return fmt.Errorf("%s total local volume size must be between %s and %s", commercialType, minSize, maxSize)
	}

	return nil
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
		privateNetworkID := locality.ExpandID(zonedID.(string))
		if pnExist {
			region, err := server.Zone.Region()
			if err != nil {
				return nil, err
			}
			currentPN, err := vpcAPI.GetPrivateNetwork(&vpc.GetPrivateNetworkRequest{
				PrivateNetworkID: locality.ExpandID(privateNetworkID),
				Region:           region,
			}, scw.WithContext(ctx))
			if err != nil {
				return nil, err
			}
			query := &instance.CreatePrivateNICRequest{
				Zone:             server.Zone,
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

func newPrivateNICHandler(api *instance.API, server string, zone scw.Zone) (*privateNICsHandler, error) {
	handler := &privateNICsHandler{
		instanceAPI: api,
		serverID:    server,
		zone:        zone,
	}
	return handler, handler.flatPrivateNICs()
}

func (ph *privateNICsHandler) detach(ctx context.Context, o interface{}, timeout time.Duration) error {
	oPtr := types.ExpandStringPtr(o)
	if oPtr != nil && len(*oPtr) > 0 {
		idPN := locality.ExpandID(*oPtr)
		// check if old private network still exist on instance server
		if p, ok := ph.privateNICsMap[idPN]; ok {
			_, err := waitForPrivateNIC(ctx, ph.instanceAPI, ph.zone, ph.serverID, locality.ExpandID(p.ID), timeout)
			if err != nil {
				return err
			}
			// detach private NIC
			err = ph.instanceAPI.DeletePrivateNIC(&instance.DeletePrivateNICRequest{
				PrivateNicID: locality.ExpandID(p.ID),
				Zone:         ph.zone,
				ServerID:     ph.serverID,
			},
				scw.WithContext(ctx))
			if err != nil {
				return err
			}

			_, err = ph.instanceAPI.WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
				ServerID:      ph.serverID,
				PrivateNicID:  p.ID,
				Zone:          ph.zone,
				Timeout:       &timeout,
				RetryInterval: scw.TimeDurationPtr(defaultInstanceRetryInterval),
			})
			if err != nil && !httperrors.Is404(err) {
				return err
			}
		}
	}

	return nil
}

func (ph *privateNICsHandler) attach(ctx context.Context, n interface{}, timeout time.Duration) error {
	if nPtr := types.ExpandStringPtr(n); nPtr != nil {
		// check if new private network was already attached on instance server
		privateNetworkID := locality.ExpandID(*nPtr)
		if _, ok := ph.privateNICsMap[privateNetworkID]; !ok {
			pn, err := ph.instanceAPI.CreatePrivateNIC(&instance.CreatePrivateNICRequest{
				Zone:             ph.zone,
				ServerID:         ph.serverID,
				PrivateNetworkID: privateNetworkID,
			})
			if err != nil {
				return err
			}

			_, err = waitForPrivateNIC(ctx, ph.instanceAPI, ph.zone, ph.serverID, pn.PrivateNic.ID, timeout)
			if err != nil {
				return err
			}

			_, err = waitForMACAddress(ctx, ph.instanceAPI, ph.zone, ph.serverID, pn.PrivateNic.ID, timeout)
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
			continue
		}
		privateNetworks = append(privateNetworks, keyRaw.(map[string]interface{}))
	}
	return d.Set("private_network", privateNetworks)
}

func (ph *privateNICsHandler) get(key string) (interface{}, error) {
	loc, id, err := locality.ParseLocalizedID(key)
	if err != nil {
		return nil, err
	}
	pn, ok := ph.privateNICsMap[id]
	if !ok {
		return nil, fmt.Errorf("could not find private network ID %s on locality %s", key, loc)
	}
	return map[string]interface{}{
		"pn_id":       key,
		"mac_address": pn.MacAddress,
		"status":      pn.State.String(),
		"zone":        loc,
		"pnic_id":     pn.ID,
	}, nil
}

func getSnapshotsFromIDs(ctx context.Context, snapIDs []interface{}, instanceAPI *instance.API) ([]*instance.GetSnapshotResponse, error) {
	snapResponses := []*instance.GetSnapshotResponse(nil)
	for _, snapID := range snapIDs {
		zone, id, err := zonal.ParseID(snapID.(string))
		if err != nil {
			return nil, err
		}
		snapshot, err := instanceAPI.GetSnapshot(&instance.GetSnapshotRequest{
			Zone:       zone,
			SnapshotID: id,
		}, scw.WithContext(ctx))
		if err != nil {
			return snapResponses, fmt.Errorf("extra volumes : could not find snapshot with id %s", snapID)
		}
		snapResponses = append(snapResponses, snapshot)
	}
	return snapResponses, nil
}

func formatImageLabel(imageUUID string) string {
	return strings.ReplaceAll(imageUUID, "-", "_")
}

func IsIPReverseDNSResolveError(err error) bool {
	invalidArgError := &scw.InvalidArgumentsError{}

	if !errors.As(err, &invalidArgError) {
		return false
	}

	for _, fields := range invalidArgError.Details {
		if fields.ArgumentName == "reverse" {
			return true
		}
	}

	return false
}

func retryUpdateReverseDNS(ctx context.Context, instanceAPI *instance.API, req *instance.UpdateIPRequest, timeout time.Duration) error {
	timeoutChannel := time.After(timeout)

	for {
		select {
		case <-time.After(defaultInstanceRetryInterval):
			_, err := instanceAPI.UpdateIP(req, scw.WithContext(ctx))
			if err != nil && IsIPReverseDNSResolveError(err) {
				continue
			}
			return err
		case <-timeoutChannel:
			_, err := instanceAPI.UpdateIP(req, scw.WithContext(ctx))
			return err
		}
	}
}

// instanceIPHasMigrated check if instance migrate from ip_id to ip_ids
// should be used if ip_id has changed
// will return true if the id removed from ip_id is present in ip_ids
func instanceIPHasMigrated(d *schema.ResourceData) bool {
	oldIP, newIP := d.GetChange("ip_id")
	// ip_id should have been removed
	if newIP != "" {
		return false
	}

	// ip_ids should have been added
	if !d.HasChange("ip_ids") {
		return false
	}

	ipIDs := types.ExpandStrings(d.Get("ip_ids"))
	for _, ipID := range ipIDs {
		if ipID == oldIP {
			return true
		}
	}

	return false
}

func instanceServerAdditionalVolumeTemplate(api *BlockAndInstanceAPI, zone scw.Zone, volumeID string) (*instance.VolumeServerTemplate, error) {
	vol, err := api.GetUnknownVolume(&GetUnknownVolumeRequest{
		VolumeID: locality.ExpandID(volumeID),
		Zone:     zone,
	})
	if err != nil {
		return nil, err
	}
	return vol.VolumeTemplate(), nil
}

func prepareRootVolume(rootVolumeI map[string]any, serverType *instance.ServerType, image string) *UnknownVolume {
	serverTypeCanBootOnBlock := serverType.VolumesConstraint.MaxSize == 0

	rootVolumeIsBootVolume := types.ExpandBoolPtr(types.GetMapValue[bool](rootVolumeI, "boot"))
	rootVolumeType := types.GetMapValue[string](rootVolumeI, "volume_type")
	sizeInput := types.GetMapValue[int](rootVolumeI, "size_in_gb")
	rootVolumeID := zonal.ExpandID(types.GetMapValue[string](rootVolumeI, "volume_id")).ID

	// If the rootVolumeType is not defined, define it depending on the offer
	if rootVolumeType == "" {
		if serverTypeCanBootOnBlock {
			rootVolumeType = instance.VolumeVolumeTypeBSSD.String()
		} else {
			rootVolumeType = instance.VolumeVolumeTypeLSSD.String()
		}
	}

	rootVolumeName := ""
	if image == "" { // When creating an instance from an image, volume should not have a name
		rootVolumeName = types.NewRandomName("vol")
	}

	var rootVolumeSize *scw.Size
	if sizeInput == 0 && rootVolumeType == instance.VolumeVolumeTypeLSSD.String() {
		// Compute the rootVolumeSize so it will be valid against the local volume constraints
		// It wouldn't be valid if another local volume is added, but in this case
		// the user would be informed that it does not fulfill the local volume constraints
		rootVolumeSize = scw.SizePtr(serverType.VolumesConstraint.MaxSize)
	} else if sizeInput > 0 {
		rootVolumeSize = scw.SizePtr(scw.Size(uint64(sizeInput) * gb))
	}

	return &UnknownVolume{
		Name:               rootVolumeName,
		ID:                 rootVolumeID,
		InstanceVolumeType: instance.VolumeVolumeType(rootVolumeType),
		Size:               rootVolumeSize,
		Boot:               rootVolumeIsBootVolume,
	}
}
