package instance

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	blockSDK "github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	instanceV2 "github.com/scaleway/scaleway-sdk-go/api/instance/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/marketplace/v2"
	product_catalog "github.com/scaleway/scaleway-sdk-go/api/product_catalog/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	scwvalidation "github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/block"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/instance/instancehelpers"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/transport"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

// newAPIWithZone returns a new instance API and the zone for a Create request
func newAPIWithZone(d *schema.ResourceData, m any) (*instance.API, scw.Zone, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return instanceAPI, zone, nil
}

// newAPIV2WithZone returns a new instance API v2 and the zone for a Create request
func newAPIV2WithZone(d *schema.ResourceData, m any) (*instanceV2.API, scw.Zone, error) {
	instanceAPI := instanceV2.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return instanceAPI, zone, nil
}

// NewAPIWithZoneAndID returns an instance API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m any, zonedID string) (*instance.API, scw.Zone, string, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonedID)
	if err != nil {
		return nil, "", "", err
	}

	return instanceAPI, zone, ID, nil
}

// NewAPIV2WithZoneAndID returns an instance API v2 with zone and ID extracted from the state
func NewAPIV2WithZoneAndID(m any, zonedID string) (*instanceV2.API, scw.Zone, string, error) {
	instanceAPI := instanceV2.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(zonedID)
	if err != nil {
		return nil, "", "", err
	}

	return instanceAPI, zone, ID, nil
}

// NewAPIWithZoneAndNestedID returns an instance API with zone and inner/outer ID extracted from the state
func NewAPIWithZoneAndNestedID(m any, zonedNestedID string) (*instance.API, scw.Zone, string, string, error) {
	instanceAPI := instance.NewAPI(meta.ExtractScwClient(m))

	zone, innerID, outerID, err := zonal.ParseNestedID(zonedNestedID)
	if err != nil {
		return nil, "", "", "", err
	}

	return instanceAPI, zone, innerID, outerID, nil
}

// NewAPIV2WithZoneAndNestedID returns an instance API v2 with zone and inner/outer ID extracted from the state
func NewAPIV2WithZoneAndNestedID(m any, zonedNestedID string) (*instanceV2.API, scw.Zone, string, string, error) {
	instanceAPI := instanceV2.NewAPI(meta.ExtractScwClient(m))

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
	if rawState == "" {
		return instance.ServerStateRunning, nil
	}

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

func reachState(ctx context.Context, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, serverID string, toState instance.ServerState) error {
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
		{instance.ServerStateStopping, instance.ServerStateStopped}:       {}, // Already stopping, just wait
	}

	actions, exist := transitionMap[[2]instance.ServerState{fromState, toState}]
	if !exist {
		return fmt.Errorf("don't know how to reach state %s from state %s for server %s", toState, fromState, serverID)
	}

	// We need to check that all volumes are ready
	for _, volume := range response.Server.Volumes {
		if volume.VolumeType == block.BlockVolumeType {
			_, err := api.BlockAPI.WaitForVolumeAndReferences(&blockSDK.WaitForVolumeAndReferencesRequest{
				VolumeID:      volume.ID,
				Zone:          zone,
				RetryInterval: transport.DefaultWaitRetryInterval,
			}, scw.WithContext(ctx))
			if err != nil {
				return err
			}
		} else if volume.State != nil && *volume.State != instance.VolumeServerStateAvailable {
			_, err = api.WaitForVolume(&instance.WaitForVolumeRequest{
				Zone:          zone,
				VolumeID:      volume.ID,
				RetryInterval: transport.DefaultWaitRetryInterval,
			}, scw.WithContext(ctx))
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
			Timeout:       new(DefaultInstanceServerWaitTimeout),
			RetryInterval: transport.DefaultWaitRetryInterval,
		}, scw.WithContext(ctx))
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
	ctx context.Context, data any,
	server *instance.Server, vpcAPI *vpc.API,
) ([]*instanceV2.CreatePrivateNetworkInterfaceRequest, error) {
	if data == nil {
		return nil, nil
	}

	var res []*instanceV2.CreatePrivateNetworkInterfaceRequest

	for _, pn := range data.([]any) {
		r := pn.(map[string]any)
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

			query := &instanceV2.CreatePrivateNetworkInterfaceRequest{
				Zone:             server.Zone,
				ServerID:         new(server.ID),
				PrivateNetworkID: currentPN.ID,
				ProjectID:        server.Project,
			}
			res = append(res, query)
		}
	}

	return res, nil
}

type privateNICsHandler struct {
	instanceAPI    *instanceV2.API
	instanceAPIV1  *instance.API
	serverID       string
	privateNICsMap map[string]*instanceV2.PrivateNetworkInterfaceSummary
	zone           scw.Zone
	projectID      string
}

func newPrivateNICHandler(api *instanceV2.API, apiV1 *instance.API, serverID string, zone scw.Zone, projectID string) (*privateNICsHandler, error) {
	handler := &privateNICsHandler{
		instanceAPI:   api,
		instanceAPIV1: apiV1,
		serverID:      serverID,
		zone:          zone,
		projectID:     projectID,
	}

	return handler, handler.flatPrivateNICs()
}

func (ph *privateNICsHandler) detach(ctx context.Context, o any, timeout time.Duration) error {
	oPtr := types.ExpandStringPtr(o)
	if oPtr != nil && len(*oPtr) > 0 {
		idPN := locality.ExpandID(*oPtr)
		// check if old private network still exist on instance server
		if p, ok := ph.privateNICsMap[idPN]; ok {
			_, err := waitForPrivateNIC(ctx, ph.instanceAPI, ph.zone, locality.ExpandID(p.ID), timeout)
			if err != nil {
				return err
			}
			// detach private NIC
			err = ph.instanceAPI.DeletePrivateNetworkInterface(&instanceV2.DeletePrivateNetworkInterfaceRequest{
				PrivateNetworkInterfaceID: locality.ExpandID(p.ID),
				Zone:                      ph.zone,
			},
				scw.WithContext(ctx))
			if err != nil {
				return err
			}

			_, err = ph.instanceAPI.WaitForPrivateNetworkInterface(&instanceV2.WaitForPrivateNetworkInterfaceRequest{
				PrivateNetworkInterfaceID: p.ID,
				Zone:                      ph.zone,
				Timeout:                   &timeout,
				RetryInterval:             new(instancehelpers.DefaultInstanceRetryInterval),
			})
			if err != nil && !httperrors.Is404(err) {
				return err
			}
		}
	}

	return nil
}

func (ph *privateNICsHandler) attach(ctx context.Context, n any, timeout time.Duration) error {
	if nPtr := types.ExpandStringPtr(n); nPtr != nil {
		// check if new private network was already attached on instance server
		privateNetworkID := locality.ExpandID(*nPtr)
		if _, ok := ph.privateNICsMap[privateNetworkID]; !ok {
			pn, err := ph.instanceAPI.CreatePrivateNetworkInterface(&instanceV2.CreatePrivateNetworkInterfaceRequest{
				Zone:             ph.zone,
				ServerID:         new(ph.serverID),
				PrivateNetworkID: privateNetworkID,
			})
			if err != nil {
				return err
			}

			_, err = waitForPrivateNIC(ctx, ph.instanceAPI, ph.zone, pn.ID, timeout)
			if err != nil {
				return err
			}

			err = waitForMACAddress(ctx, ph.instanceAPIV1, ph.zone, ph.serverID, pn.ID, timeout)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ph *privateNICsHandler) set(d *schema.ResourceData) error {
	raw := d.Get("private_network")
	privateNetworks := []map[string]any(nil)

	for index := range raw.([]any) {
		pnKey := fmt.Sprintf("private_network.%d.pn_id", index)
		keyValue := d.Get(pnKey)

		keyRaw, err := ph.get(keyValue.(string))
		if err != nil {
			continue
		}

		privateNetworks = append(privateNetworks, keyRaw.(map[string]any))
	}

	return d.Set("private_network", privateNetworks)
}

func (ph *privateNICsHandler) get(key string) (any, error) {
	loc, id, _ := locality.ParseLocalizedID(key)
	if loc == "" {
		loc = ph.zone.String()
	}

	pn, ok := ph.privateNICsMap[id]
	if !ok {
		return nil, fmt.Errorf("could not find private network ID %s on locality %s", key, loc)
	}

	return map[string]any{
		"pn_id":       key,
		"mac_address": pn.MacAddress,
		"status":      pn.Status.String(),
		"zone":        loc,
		"pnic_id":     pn.ID,
	}, nil
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
		case <-time.After(instancehelpers.DefaultInstanceRetryInterval):
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

func instanceServerAdditionalVolumeTemplate(api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, volumeID string) (*instance.VolumeServerTemplate, error) {
	vol, err := api.GetUnknownVolume(&instancehelpers.GetUnknownVolumeRequest{
		VolumeID: locality.ExpandID(volumeID),
		Zone:     zone,
	})
	if err != nil {
		return nil, err
	}

	return vol.VolumeTemplate(), nil
}

func prepareRootVolume(rootVolumeI map[string]any, serverType *instance.ServerType, image string) *instancehelpers.UnknownVolume {
	rootVolumeIsBootVolume := types.ExpandBoolPtr(types.GetMapValue[bool](rootVolumeI, "boot"))
	rootVolumeType := types.GetMapValue[string](rootVolumeI, "volume_type")
	sizeInput := types.GetMapValue[int](rootVolumeI, "size_in_gb")
	rootVolumeID := zonal.ExpandID(types.GetMapValue[string](rootVolumeI, "volume_id")).ID

	rootVolumeName := ""
	if image == "" { // When creating an instance from an image, volume should not have a name
		rootVolumeName = types.NewRandomName("vol")
	}

	var rootVolumeSize *scw.Size
	if sizeInput == 0 && rootVolumeType == instance.VolumeVolumeTypeLSSD.String() {
		// Compute the rootVolumeSize so it will be valid against the local volume constraints
		// It wouldn't be valid if another local volume is added, but in this case
		// the user would be informed that it does not fulfill the local volume constraints
		rootVolumeSize = new(serverType.VolumesConstraint.MaxSize)
	} else if sizeInput > 0 {
		rootVolumeSize = new(scw.Size(uint64(sizeInput) * gb))
	}

	return &instancehelpers.UnknownVolume{
		Name:               rootVolumeName,
		ID:                 rootVolumeID,
		InstanceVolumeType: instance.VolumeVolumeType(rootVolumeType),
		Size:               rootVolumeSize,
		Boot:               rootVolumeIsBootVolume,
	}
}

func attachNewFileSystem(ctx context.Context, newIDs map[string]struct{}, oldIDs map[string]struct{}, api *instance.API, zone scw.Zone, server *instance.Server) error {
	for id := range newIDs {
		if _, alreadyAttached := oldIDs[id]; !alreadyAttached {
			_, err := api.AttachServerFileSystem(&instance.AttachServerFileSystemRequest{
				Zone:         zone,
				ServerID:     server.ID,
				FilesystemID: locality.ExpandID(id),
			})
			if err != nil {
				return fmt.Errorf("error attaching filesystem %s: %w", id, err)
			}

			_, err = waitForFilesystems(ctx, api, zone, server.ID, DefaultInstanceServerWaitTimeout)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func detachOldFileSystem(ctx context.Context, oldIDs map[string]struct{}, newIDs map[string]struct{}, api *instance.API, zone scw.Zone, server *instance.Server) error {
	for id := range oldIDs {
		if _, stillPresent := newIDs[id]; !stillPresent {
			_, err := api.DetachServerFileSystem(&instance.DetachServerFileSystemRequest{
				Zone:         zone,
				ServerID:     server.ID,
				FilesystemID: locality.ExpandID(id),
			})
			if err != nil {
				return fmt.Errorf("error detaching filesystem %s: %w", id, err)
			}

			_, err = waitForFilesystems(ctx, api, zone, server.ID, DefaultInstanceServerWaitTimeout)
			if err != nil && !httperrors.Is404(err) {
				return err
			}
		}
	}

	return nil
}

func collectFilesystemIDs(fsList []any, target map[string]struct{}) {
	for _, fs := range fsList {
		if fsMap, ok := fs.(map[string]any); ok {
			id := fsMap["filesystem_id"].(string)
			target[id] = struct{}{}
		}
	}
}

func DeleteASGServers(
	ctx context.Context,
	api *instance.API,
	zone scw.Zone,
	groupID string,
	timeout time.Duration,
) error {
	resp, err := api.ListServers(&instance.ListServersRequest{
		Zone: zone,
		Name: types.ExpandStringPtr(groupID),
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	for _, srv := range resp.Servers {
		switch srv.State {
		case instance.ServerStateRunning:
			if _, err = api.ServerAction(&instance.ServerActionRequest{
				Zone:     zone,
				ServerID: srv.ID,
				Action:   instance.ServerActionTerminate,
			}, scw.WithContext(ctx)); err != nil {
				return err
			}
		case instance.ServerStateStopped, instance.ServerStateStoppedInPlace:
			if err = api.DeleteServer(&instance.DeleteServerRequest{
				Zone:     zone,
				ServerID: srv.ID,
			}, scw.WithContext(ctx)); err != nil {
				return err
			}
		}

		_, err := api.WaitForServer(&instance.WaitForServerRequest{
			Zone:     zone,
			ServerID: srv.ID,
			Timeout:  new(timeout),
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return err
		}
	}

	return nil
}

// FindIPInList looks for an IP in a list to avoid using the deprecated server.PublicIP field
func FindIPInList(ipID string, ips []*instance.ServerIP) *instance.ServerIP {
	id := zonal.ExpandID(ipID).ID
	for _, ip := range ips {
		if ip.ID == id {
			return ip
		}
	}

	return nil
}

func detachFileSystemDelete(ctx context.Context, filesystems any, api *instance.API, zone scw.Zone, id string) diag.Diagnostics {
	fsList := filesystems.([]any)
	for i, fsRaw := range fsList {
		fsMap := fsRaw.(map[string]any)

		fsIDRaw, ok := fsMap["filesystem_id"]
		if !ok || fsIDRaw == nil {
			return diag.Errorf("filesystem_id is missing or nil for filesystem at index %d", i)
		}

		fsID := fsIDRaw.(string)

		newFileSystemID := types.ExpandStringPtr(fsID)
		if newFileSystemID == nil {
			return diag.Errorf("failed to expand filesystem_id pointer at index %d", i)
		}

		_, err := api.DetachServerFileSystem(&instance.DetachServerFileSystemRequest{
			Zone:         zone,
			ServerID:     id,
			FilesystemID: locality.ExpandID(*newFileSystemID),
		})
		if err != nil {
			return diag.FromErr(err)
		}

		_, err = waitForFilesystems(ctx, api, zone, id, DefaultInstanceServerWaitTimeout)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func detachPrivateNetworkDelete(ctx context.Context, d *schema.ResourceData, pnRaw any, apiV2 *instanceV2.API, apiV1 *instance.API, zone scw.Zone, id string, projectID string) diag.Diagnostics {
	ph, err := newPrivateNICHandler(apiV2, apiV1, id, zone, projectID)
	if err != nil {
		return diag.FromErr(err)
	}

	for index := range pnRaw.([]any) {
		pnKey := fmt.Sprintf("private_network.%d.pn_id", index)
		pn := d.Get(pnKey)

		err := ph.detach(ctx, pn, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func terminateServer(ctx context.Context, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, id string, timeout time.Duration) error {
	// reach running state (mandatory for termination)
	err := reachState(ctx, api, zone, id, instance.ServerStateRunning)
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	err = api.ServerActionAndWait(&instance.ServerActionAndWaitRequest{
		Zone:     zone,
		ServerID: id,
		Action:   instance.ServerActionTerminate,
		Timeout:  &timeout,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	return nil
}

func deleteServer(ctx context.Context, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, id string, timeout time.Duration) error {
	_, err := waitForServer(ctx, api.API, zone, id, timeout)
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	// reach stopped state
	err = reachState(ctx, api, zone, id, instance.ServerStateStopped)
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	err = api.DeleteServer(&instance.DeleteServerRequest{
		Zone:     zone,
		ServerID: id,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	_, err = waitForServer(ctx, api.API, zone, id, timeout)
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	return nil
}

func instanceServerCanMigrate(api *instance.API, server *instance.Server, requestedType string) error {
	var localVolumeSize scw.Size

	for _, volume := range server.Volumes {
		if volume.VolumeType == instance.VolumeServerVolumeTypeLSSD && volume.Size != nil {
			localVolumeSize += *volume.Size
		}
	}

	serverType, err := api.GetServerType(&instance.GetServerTypeRequest{
		Zone: server.Zone,
		Name: requestedType,
	})
	if err != nil {
		return err
	}

	if serverType.VolumesConstraint != nil &&
		(localVolumeSize > serverType.VolumesConstraint.MaxSize) ||
		(localVolumeSize < serverType.VolumesConstraint.MinSize) {
		return fmt.Errorf("local volume total size does not respect type constraint, expected beteween (%dGB, %dGB), got %sGB",
			serverType.VolumesConstraint.MinSize/scw.GB,
			serverType.VolumesConstraint.MaxSize/scw.GB,
			localVolumeSize/scw.GB)
	}

	return nil
}

func customDiffInstanceRootVolumeSize(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	if !diff.HasChange("root_volume.0.size_in_gb") || diff.Id() == "" {
		return nil
	}

	instanceAPI, zone, id, err := NewAPIWithZoneAndID(meta, diff.Id())
	if err != nil {
		return err
	}

	resp, err := instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to check server root volume type: %w", err)
	}

	if rootVolume, hasRootVolume := resp.Server.Volumes["0"]; hasRootVolume {
		if rootVolume.VolumeType == instance.VolumeServerVolumeTypeLSSD {
			return diff.ForceNew("root_volume.0.size_in_gb")
		}
	}

	return nil
}

func customDiffInstanceServerType(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	if !diff.HasChange("type") || diff.Id() == "" {
		return nil
	}

	if diff.Get("replace_on_type_change").(bool) {
		return diff.ForceNew("type")
	}

	instanceAPI, zone, id, err := NewAPIWithZoneAndID(meta, diff.Id())
	if err != nil {
		return err
	}

	_, newValue := diff.GetChange("type")
	newType := newValue.(string)

	resp, err := instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to check server type change: %w", err)
	}

	err = instanceServerCanMigrate(instanceAPI, resp.Server, newType)
	if err != nil {
		return fmt.Errorf("cannot change server type: %w", err)
	}

	return nil
}

func customDiffInstanceServerImage(ctx context.Context, diff *schema.ResourceDiff, m any) error {
	if diff.Get("image") == "" || !diff.HasChange("image") || diff.Id() == "" {
		return nil
	}

	// We get the server to fetch the UUID of the image
	instanceAPI, zone, id, err := NewAPIWithZoneAndID(m, diff.Id())
	if err != nil {
		return err
	}

	server, err := instanceAPI.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: id,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	// If 'image' field is defined by the user and server.Image is empty, we should create a new server
	if server.Server.Image == nil {
		return diff.ForceNew("image")
	}

	// We get the image as it is defined by the user
	image := regional.ExpandID(diff.Get("image").(string))
	if scwvalidation.IsUUID(image.ID) {
		if image.ID == zonal.ExpandID(server.Server.Image.ID).ID {
			return nil
		}
	}

	// If image is a label, we check that server.Image.ID matches the label in case the user has edited
	// the image with another tool.
	marketplaceAPI := marketplace.NewAPI(meta.ExtractScwClient(m))

	marketplaceImage, err := marketplaceAPI.GetLocalImage(&marketplace.GetLocalImageRequest{
		LocalImageID: server.Server.Image.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		// If UUID is not in marketplace, then it's an image change
		if httperrors.Is404(err) {
			return diff.ForceNew("image")
		}

		return err
	}

	if marketplaceImage.Label != image.ID {
		return diff.ForceNew("image")
	}

	return nil
}

func ResourceInstanceServerMigrate(ctx context.Context, d *schema.ResourceData, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, id string) error {
	server, err := waitForServer(ctx, api.API, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("failed to wait for server before changing server type: %w", err)
	}

	beginningState := server.State

	err = reachState(ctx, api, zone, id, instance.ServerStateStopped)
	if err != nil {
		return fmt.Errorf("failed to stop server before changing server type: %w", err)
	}

	_, err = api.UpdateServer(&instance.UpdateServerRequest{
		Zone:           zone,
		ServerID:       id,
		CommercialType: types.ExpandStringPtr(d.Get("type")),
	})
	if err != nil {
		return errors.New("failed to change server type server")
	}

	err = reachState(ctx, api, zone, id, beginningState)
	if err != nil {
		return fmt.Errorf("failed to start server after changing server type: %w", err)
	}

	return nil
}

func ResourceInstanceServerUpdateIPs(ctx context.Context, d *schema.ResourceData, instanceAPI *instance.API, zone scw.Zone, id string, attribute string) error {
	server, err := waitForServer(ctx, instanceAPI, zone, id, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return err
	}

	var schemaIPs []any

	switch attribute {
	case "ip_id":
		schemaIP := d.Get(attribute).(string)
		schemaIPs = append(schemaIPs, schemaIP)
	case "ip_ids":
		schemaIPs = d.Get(attribute).([]any)
	}

	requestedIPs := make(map[string]bool, len(schemaIPs))

	// Gather request IPs in a map
	for _, rawIP := range schemaIPs {
		requestedIPs[locality.ExpandID(rawIP)] = false
	}

	// Detach all IPs that are not requested and set to true the one that are already attached
	for _, ip := range server.PublicIPs {
		_, isRequested := requestedIPs[ip.ID]
		if isRequested {
			requestedIPs[ip.ID] = true
		} else {
			_, err := instanceAPI.UpdateIP(&instance.UpdateIPRequest{
				Zone: zone,
				IP:   ip.ID,
				Server: &instance.NullableStringValue{
					Null: true,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to detach IP: %w", err)
			}
		}
	}

	// Attach all remaining IPs that are not attached
	for ipID, isAttached := range requestedIPs {
		if isAttached {
			continue
		}

		_, err := instanceAPI.UpdateIP(&instance.UpdateIPRequest{
			Zone: zone,
			IP:   ipID,
			Server: &instance.NullableStringValue{
				Value: server.ID,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to attach IP: %w", err)
		}
	}

	return nil
}

func ResourceInstanceServerUpdateRootVolumeIOPS(ctx context.Context, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, serverID string, iops *uint32) diag.Diagnostics {
	res, err := api.GetServer(&instance.GetServerRequest{
		Zone:     zone,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	rootVolume, exists := res.Server.Volumes["0"]
	if exists {
		_, err := api.BlockAPI.UpdateVolume(&blockSDK.UpdateVolumeRequest{
			Zone:     zone,
			VolumeID: rootVolume.ID,
			PerfIops: iops,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.Diagnostics{{
				Severity:      diag.Warning,
				Summary:       "Failed to update root_volume iops",
				Detail:        err.Error(),
				AttributePath: cty.GetAttrPath("root_volume.0.sbs_iops"),
			}}
		}
	} else {
		return diag.Diagnostics{{
			Severity:      diag.Warning,
			Summary:       "Failed to find root_volume",
			Detail:        "Failed to update root_volume IOPS",
			AttributePath: cty.GetAttrPath("root_volume.0.sbs_iops"),
		}}
	}

	return nil
}

// instanceServerVolumesUpdate updates root_volume size and returns the list of volumes templates that should be updated for the server.
// It uses root_volume and additional_volume_ids to build the volumes templates.
func instanceServerVolumesUpdate(ctx context.Context, d *schema.ResourceData, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, serverIsStopped bool) (map[string]*instance.VolumeServerTemplate, error) {
	volumes := map[string]*instance.VolumeServerTemplate{}
	raw, hasAdditionalVolumes := d.GetOk("additional_volume_ids")

	if d.HasChange("root_volume.0.size_in_gb") {
		err := api.ResizeUnknownVolume(&instancehelpers.ResizeUnknownVolumeRequest{
			VolumeID: zonal.ExpandID(d.Get("root_volume.0.volume_id")).ID,
			Zone:     zone,
			Size:     new(scw.Size(d.Get("root_volume.0.size_in_gb").(int)) * scw.GB),
		}, scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}
	}

	volumes["0"] = &instance.VolumeServerTemplate{
		ID:   new(zonal.ExpandID(d.Get("root_volume.0.volume_id")).ID),
		Name: new(types.NewRandomName("vol")), // name is ignored by the API, any name will work here
		Boot: types.ExpandBoolPtr(d.Get("root_volume.0.boot")),
	}

	if !hasAdditionalVolumes {
		raw = []any{} // Set an empty list if not volumes exist
	}

	for i, volumeID := range raw.([]any) {
		volumeHasChange := d.HasChange("additional_volume_ids." + strconv.Itoa(i))

		volume, err := api.GetUnknownVolume(&instancehelpers.GetUnknownVolumeRequest{
			VolumeID: zonal.ExpandID(volumeID).ID,
			Zone:     zone,
		}, scw.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to get updated volume: %w", err)
		}

		// local volumes can only be added when the server is stopped
		if volumeHasChange && !serverIsStopped && volume.IsLocal() && volume.IsAttached() {
			return nil, errors.New("instance must be stopped to change local volumes")
		}

		volumes[strconv.Itoa(i+1)] = volume.VolumeTemplate()
	}

	return volumes, nil
}

func GetEndOfServiceDate(ctx context.Context, client *scw.Client, zone scw.Zone, commercialType string) (string, error) {
	api := product_catalog.NewPublicCatalogAPI(client)

	products, err := api.ListPublicCatalogProducts(&product_catalog.PublicCatalogAPIListPublicCatalogProductsRequest{
		Zone: &zone,
		ProductTypes: []product_catalog.ListPublicCatalogProductsRequestProductType{
			product_catalog.ListPublicCatalogProductsRequestProductTypeInstance,
		},
		APIIDs: []string{commercialType},
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("could not list product catalog entries: %w", err)
	}

	if products.TotalCount != 1 {
		return "", fmt.Errorf("expected exactly 1 PCU entry for %q, got %d", commercialType, products.TotalCount)
	}

	return products.Products[0].EndOfLifeAt.Format(time.DateOnly), nil
}

func renameRootVolumeIfNeeded(d *schema.ResourceData, api *instancehelpers.BlockAndInstanceAPI, zone scw.Zone, volumes map[string]*instance.VolumeServer) error {
	if volumes == nil || volumes["0"] == nil {
		return nil
	}

	if rootVolumeName, setByUser := meta.GetRawConfigForKey(d, "root_volume.0.name", cty.String); setByUser {
		if volumes["0"].Name == nil || *volumes["0"].Name != rootVolumeName {
			err := api.RenameUnknownVolume(&instancehelpers.RenameUnknownVolumeRequest{
				Zone:     zone,
				VolumeID: volumes["0"].ID,
				Name:     new(rootVolumeName.(string)),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func errorCheck(err error, message string) bool {
	return strings.Contains(err.Error(), message)
}
