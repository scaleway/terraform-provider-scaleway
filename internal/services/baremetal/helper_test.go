package baremetal_test

import (
	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func IsOfferAvailable(offerID string, zone scw.Zone, tt *acctest.TestTools) bool {
	api := baremetal.NewAPI(tt.Meta.ScwClient())
	offer, _ := api.GetOffer(&baremetal.GetOfferRequest{
		Zone:    zone,
		OfferID: offerID,
	})
	if offer.Stock == baremetal.OfferStockAvailable {
		return true
	}
	return false
}
