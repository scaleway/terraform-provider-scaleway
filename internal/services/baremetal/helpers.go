package baremetal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
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

var errNoPrivateNetworkOption = errors.New("attaching a private network requires enabling an offer-compatible private network option")

// customDiffPrivateNetworkOption checks that the private_network option has been set if there is a private_network
func customDiffPrivateNetworkOption() schema.CustomizeDiffFunc {
	return func(ctx context.Context, diff *schema.ResourceDiff, m any) error {
		if _, ok := diff.GetOk("private_network"); !ok {
			return nil
		}

		raw, ok := diff.GetOk("options")
		if !ok {
			return errNoPrivateNetworkOption
		}

		diffOptions, err := expandOptions(raw)
		if err != nil {
			return err
		}

		api := baremetal.NewAPI(meta.ExtractScwClient(m))

		zone, err := meta.ExtractZone(diff, m)
		if err != nil {
			return err
		}

		localizedOfferID, ok := diff.GetOk("offer")
		if !ok {
			return errors.New("listing options for offer: no offer attribute to retrieve options from")
		}

		offerID := locality.ExpandID(localizedOfferID)

		offerOptions, err := api.ListOptions(&baremetal.ListOptionsRequest{
			Zone:     zone,
			OfferID:  new(offerID),
			PageSize: new(uint32(100)),
		}, scw.WithAllPages())
		if err != nil {
			return fmt.Errorf("listing options for offer %q: %w", localizedOfferID, err)
		}

		hasPrivateNetworkOptions := false

		for _, diffOption := range diffOptions {
			for _, offerOption := range offerOptions.Options {
				tflog.Debug(ctx, "checking option", map[string]any{"id": offerOption.ID, "name": offerOption.Name, "private_network": offerOption.PrivateNetwork})

				if offerOption.ID != diffOption.ID {
					continue
				}

				if offerOption.PrivateNetwork != nil {
					hasPrivateNetworkOptions = true
				}

				break
			}

			if hasPrivateNetworkOptions {
				break
			}
		}

		if !hasPrivateNetworkOptions {
			return errNoPrivateNetworkOption
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
	return !slices.Contains(offer.IncompatibleOsIDs, os.ID)
}
