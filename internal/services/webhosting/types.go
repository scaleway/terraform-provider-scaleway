package webhosting

import (
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func flattenOfferPrice(price *scw.Money) any {
	return price.String()
}
