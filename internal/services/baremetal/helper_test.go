package baremetal_test

import (
	"os"

	"github.com/scaleway/scaleway-sdk-go/api/baremetal/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func IsOfferAvailable(offerName string, zone scw.Zone, tt *acctest.TestTools) bool {
	api := baremetal.NewAPI(tt.Meta.ScwClient())
	offer, _ := api.GetOfferByName(&baremetal.GetOfferByNameRequest{
		OfferName: offerName,
		Zone:      zone,
	})

	return offer.Stock == baremetal.OfferStockAvailable
}
