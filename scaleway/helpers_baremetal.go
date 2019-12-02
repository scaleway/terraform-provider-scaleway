package scaleway

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	baremetal "github.com/scaleway/scaleway-sdk-go/api/baremetal/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	baremetalServerWaitForTimeout   = 60 * time.Minute
	baremetalServerRetryFuncTimeout = baremetalServerWaitForTimeout + time.Minute // some RetryFunc are calling a WaitFor
)

var baremetalServerResourceTimeout = baremetalServerRetryFuncTimeout + time.Minute

// getInstanceAPIWithZone returns a new baremetal API and the zone for a Create request
func getBaremetalAPIWithZone(d *schema.ResourceData, m interface{}) (*baremetal.API, scw.Zone, error) {
	meta := m.(*Meta)
	baremetalAPI := baremetal.NewAPI(meta.scwClient)

	zone, err := getZone(d, meta)
	return baremetalAPI, zone, err
}

// getInstanceAPIWithZoneAndID returns an baremetal API with zone and ID extracted from the state
func getBaremetalAPIWithZoneAndID(m interface{}, id string) (*baremetal.API, scw.Zone, string, error) {
	meta := m.(*Meta)
	baremetalAPI := baremetal.NewAPI(meta.scwClient)

	zone, ID, err := parseZonedID(id)
	return baremetalAPI, zone, ID, err
}

// TODO: Remove it when SDK will handle it.
// getBaremetalOfferByName call baremetal API to get an offer by its exact name.
func getBaremetalOfferByName(baremetalAPI *baremetal.API, zone scw.Zone, offerName string) (*baremetal.Offer, error) {
	offerRes, err := baremetalAPI.ListOffers(&baremetal.ListOffersRequest{
		Zone: zone,
	}, scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	offerName = strings.ToUpper(offerName)
	for _, offer := range offerRes.Offers {
		if offer.Name == offerName {
			return offer, nil
		}
	}
	return nil, fmt.Errorf("cannot find the offer %s", offerName)
}

// TODO: Remove it when SDK will handle it.
// getBaremetalOfferByID call baremetal API to get an offer by its exact name.
func getBaremetalOfferByID(baremetalAPI *baremetal.API, zone scw.Zone, offerID string) (*baremetal.Offer, error) {
	offerRes, err := baremetalAPI.ListOffers(&baremetal.ListOffersRequest{
		Zone: zone,
	}, scw.WithAllPages())
	if err != nil {
		return nil, err
	}

	for _, offer := range offerRes.Offers {
		if offer.ID == offerID {
			return offer, nil
		}
	}
	return nil, fmt.Errorf("cannot find the offer %s", offerID)
}
