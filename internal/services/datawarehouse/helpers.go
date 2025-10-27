package datawarehouse

import (
	"time"

	datawarehouseapi "github.com/scaleway/scaleway-sdk-go/api/datawarehouse/v1beta1"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
)

const (
	defaultWaitRetryInterval = 30 * time.Second
)

func NewAPI(m any) *datawarehouseapi.API {
	return datawarehouseapi.NewAPI(meta.ExtractScwClient(m))
}
