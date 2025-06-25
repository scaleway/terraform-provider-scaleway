package baremetal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	baremetalV3 "github.com/scaleway/scaleway-sdk-go/api/baremetal/v3"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/scaleway-sdk-go/validation"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

const (
	ServerTimeout          = 80 * time.Minute
	serverRetryFuncTimeout = ServerTimeout + time.Minute // some RetryFunc are calling a WaitFor
	defaultServerTimeout   = serverRetryFuncTimeout + time.Minute
	retryInterval          = 5 * time.Second
)

// newAPIWithZone returns a new API and the zone for a Create request
func newAPIWithZone(d *schema.ResourceData, m any) (*baremetal.API, scw.Zone, error) {
	api := baremetal.NewAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return api, zone, nil
}

// NewAPIWithZoneAndID returns an API with zone and ID extracted from the state
func NewAPIWithZoneAndID(m any, id string) (*baremetal.API, zonal.ID, error) {
	api := baremetal.NewAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, zonal.ID{}, err
	}

	return api, zonal.NewID(zone, ID), nil
}

// returns a new private network API and the zone for a Create request
func newPrivateNetworkAPIWithZone(d *schema.ResourceData, m any) (*baremetalV3.PrivateNetworkAPI, scw.Zone, error) {
	privateNetworkAPI := baremetalV3.NewPrivateNetworkAPI(meta.ExtractScwClient(m))

	zone, err := meta.ExtractZone(d, m)
	if err != nil {
		return nil, "", err
	}

	return privateNetworkAPI, zone, nil
}

// NewPrivateNetworkAPIWithZoneAndID returns a private network API with zone and ID extracted from the state
func NewPrivateNetworkAPIWithZoneAndID(m any, id string) (*baremetalV3.PrivateNetworkAPI, zonal.ID, error) {
	privateNetworkAPI := baremetalV3.NewPrivateNetworkAPI(meta.ExtractScwClient(m))

	zone, ID, err := zonal.ParseID(id)
	if err != nil {
		return nil, zonal.ID{}, err
	}

	return privateNetworkAPI, zonal.NewID(zone, ID), nil
}

func detachAllPrivateNetworkFromServer(ctx context.Context, d *schema.ResourceData, m any, serverID string) error {
	privateNetworkAPI, zone, err := newPrivateNetworkAPIWithZone(d, m)
	if err != nil {
		return err
	}

	listPrivateNetwork, err := privateNetworkAPI.ListServerPrivateNetworks(&baremetalV3.PrivateNetworkAPIListServerPrivateNetworksRequest{
		Zone:     zone,
		ServerID: &serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	for _, pn := range listPrivateNetwork.ServerPrivateNetworks {
		err := privateNetworkAPI.DeleteServerPrivateNetwork(&baremetalV3.PrivateNetworkAPIDeleteServerPrivateNetworkRequest{
			Zone:             zone,
			ServerID:         serverID,
			PrivateNetworkID: pn.PrivateNetworkID,
		}, scw.WithContext(ctx))
		if err != nil {
			return err
		}
	}

	_, err = waitForServerPrivateNetwork(ctx, privateNetworkAPI, zone, serverID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return err
	}

	return nil
}

func installServer(ctx context.Context, d *schema.ResourceData, api *baremetal.API, installServerRequest *baremetal.InstallServerRequest) error {
	installServerRequest.OsID = locality.ExpandID(d.Get("os"))
	installServerRequest.SSHKeyIDs = types.ExpandStrings(d.Get("ssh_key_ids"))

	_, err := api.InstallServer(installServerRequest, scw.WithContext(ctx))
	if err != nil {
		return err
	}

	return nil
}

func FindOfferByID(ctx context.Context, api *baremetal.API, zone scw.Zone, offerID string) (*baremetal.Offer, error) {
	subscriptionPeriods := []baremetal.OfferSubscriptionPeriod{
		baremetal.OfferSubscriptionPeriodHourly,
		baremetal.OfferSubscriptionPeriodMonthly,
	}

	for _, subscriptionPeriod := range subscriptionPeriods {
		res, err := api.ListOffers(&baremetal.ListOffersRequest{
			Zone:               zone,
			SubscriptionPeriod: subscriptionPeriod,
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		for _, offer := range res.Offers {
			if offer.ID == offerID {
				return offer, nil
			}
		}
	}

	return nil, fmt.Errorf("offer %s not found in zone %s", offerID, zone)
}

func compareOptions(slice1, slice2 []*baremetal.ServerOption) []*baremetal.ServerOption {
	var diff []*baremetal.ServerOption

	m := make(map[string]struct{}, len(slice1))
	for _, option := range slice1 {
		m[option.ID] = struct{}{}
	}
	// find the differences
	for _, option := range slice2 {
		if _, foundID := m[option.ID]; !foundID {
			diff = append(diff, option)
		} else if foundID {
			if _, foundExp := m[types.FlattenTime(option.ExpiresAt).(string)]; !foundExp {
				diff = append(diff, option)
			}
		}
	}

	return diff
}

// customDiffPrivateNetworkOption checks that the private_network option has been set if there is a private_network
func customDiffPrivateNetworkOption() func(ctx context.Context, diff *schema.ResourceDiff, i any) error {
	return func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
		var isPrivateNetworkOption bool

		_, okPrivateNetwork := diff.GetOk("private_network")

		options, optionsExist := diff.GetOk("options")
		if optionsExist {
			opSpecs, err := expandOptions(options)
			if err != nil {
				return err
			}

			for j := range opSpecs {
				// private network option ID
				if opSpecs[j].ID == "cd4158d7-2d65-49be-8803-c4b8ab6f760c" {
					isPrivateNetworkOption = true
				}
			}
		}

		if okPrivateNetwork && !isPrivateNetworkOption {
			return errors.New("private network option needs to be enabled in order to attach a private network")
		}

		return nil
	}
}

func privateNetworkSetHash(v any) int {
	m := v.(map[string]any)
	id := locality.ExpandID(m["id"].(string))

	var buf bytes.Buffer

	buf.WriteString(id)

	if ipamIPs, ok := m["ipam_ip_ids"]; ok && ipamIPs != nil {
		ipamIPsList := ipamIPs.([]any)

		var ipamIPIDs []string

		for _, ip := range ipamIPsList {
			if ipStr, ok := ip.(string); ok && ipStr != "" {
				ipamIPIDs = append(ipamIPIDs, ipStr)
			}
		}

		sort.Strings(ipamIPIDs)

		for _, ipID := range ipamIPIDs {
			buf.WriteString("-")
			buf.WriteString(ipID)
		}
	}

	return schema.HashString(buf.String())
}

func getOfferInformations(ctx context.Context, offer any, id string, i any) (*baremetal.Offer, error) {
	api, zone, err := NewAPIWithZoneAndID(i, id)
	if err != nil {
		return nil, err
	}

	if validation.IsUUID(regional.ExpandID(offer.(string)).ID) {
		offerID := regional.ExpandID(offer.(string))

		return FindOfferByID(ctx, api, zone.Zone, offerID.ID)
	} else {
		return api.GetOfferByName(&baremetal.GetOfferByNameRequest{
			OfferName: offer.(string),
			Zone:      zone.Zone,
		})
	}
}

func customDiffOffer() func(ctx context.Context, diff *schema.ResourceDiff, i any) error {
	return func(ctx context.Context, diff *schema.ResourceDiff, i any) error {
		if diff.Get("offer") == "" || !diff.HasChange("offer") || diff.Id() == "" {
			return nil
		}

		oldOffer, newOffer := diff.GetChange("offer")

		oldOfferInfo, err := getOfferInformations(ctx, oldOffer, diff.Id(), i)
		if err != nil {
			return errors.New(err.Error())
		}

		newOfferInfo, err := getOfferInformations(ctx, newOffer, diff.Id(), i)
		if err != nil {
			return errors.New(err.Error())
		}

		if oldOfferInfo.Name != newOfferInfo.Name {
			return diff.ForceNew("offer")
		}

		if oldOfferInfo.SubscriptionPeriod == baremetal.OfferSubscriptionPeriodMonthly && newOfferInfo.SubscriptionPeriod == baremetal.OfferSubscriptionPeriodHourly {
			return errors.New("invalid plan transition: you cannot transition from a monthly plan to an hourly plan. Only the reverse (hourly to monthly) is supported. Please update your configuration accordingly")
		}

		return nil
	}
}

func isOSCompatible(offer *baremetal.Offer, os *baremetal.OS) bool {
	for _, incompatible := range offer.IncompatibleOsIDs {
		if os.ID == incompatible {
			return false
		}
	}

	return true
}
